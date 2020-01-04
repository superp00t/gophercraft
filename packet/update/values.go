package update

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"sort"
	"sync"

	"github.com/davecgh/go-spew/spew"

	"github.com/superp00t/etc"
	"github.com/superp00t/etc/yo"
	"github.com/superp00t/gophercraft/guid"
)

type FieldFlags uint32

type ValueMask uint32

const (
	// ValuesCreate sends all [shareable] fields. Otherwise, it sends only fields that were changed.
	ValuesCreate ValueMask = 1 << iota
	// ValuesPrivate sends private fields when enabled
	ValuesPrivate
	// ValuesParty sends fields for party memberss
	ValuesParty

	ValuesNone ValueMask = 0
)

// ValuesDescriptor contains the offset data of a particular game version.
// type ValuesDescriptor map[Global]FieldDefinition

// ValuesBlock contains a map of global update fields to their corresponding Go types (uint32, []*uint32, byte, GUID)
type ValuesBlock struct {
	sync.Mutex // write access
	Values     map[Global]interface{}
	Changes    map[Global]bool
}

type valueKeysSorter []Global

func (v valueKeysSorter) Less(i, j int) bool {
	return v[i] < v[j]
}

func (v valueKeysSorter) Swap(i, j int) {
	iv := v[i]
	jv := v[j]
	v[i] = jv
	v[j] = iv
}

func (v valueKeysSorter) Len() int {
	return len(v)
}

func (v *ValuesBlock) SortedKeys() []Global {
	g := []Global{}

	for k := range v.Values {
		g = append(g, k)
	}

	sort.Sort(valueKeysSorter(g))

	return g
}

func (v *ValuesBlock) PrintString() string {
	str := "&ValuesBlock{{\n"

	for _, global := range v.SortedKeys() {
		str += global.String() + ": "

		data := v.Values[global]
		switch data.(type) {
		case uint32, float32, guid.GUID:
			str += fmt.Sprint(data)
		case uint8:
			str += fmt.Sprintf("0x%02X", data)
		default:
			str += spew.Sdump(data)
		}

		str += "\n"
	}

	str += "}}\n"
	return str
}

func (v *ValuesBlock) Type() BlockType {
	return Values
}

func getClassName(objectGUID guid.GUID) (string, error) {
	className := "Object"
	switch objectGUID.HighType() {
	case guid.Player:
		className = "Player"
	case guid.Corpse:
		className = "Corpse"
	case guid.Creature, guid.Pet, guid.Vehicle:
		className = "Unit"
	case guid.GameObject, guid.Mo_Transport, guid.Transport:
		className = "GameObject"
	case guid.DynamicObject:
		className = "DynamicObject"
	case guid.Item:
		className = "Item"
	default:
		return "", fmt.Errorf("update: can't find class name for %s", objectGUID.HighType())
	}

	return className, nil
}

func (d *ValuesDecoder) decodeGUID(offset uint32) (guid.GUID, error) {
	t1 := d.offsetEnabled(offset)
	t2 := d.offsetEnabled(offset + 1)
	if (t2 == false && t1 == true) || (t2 == true && t1 == false) {
		// Original MaNGOS doesn't do this. The operators are probably just trying to save a couple bytes.
		// Because players usually have a high guid = 0, and the client already has this value set to 0, it can be ignored.
		yo.Warn(fmt.Errorf("update: partial GUID detected, 0x%04X, %t, 0x%04X, %t", offset, t1, offset+1, t2))
	}

	g1 := etc.NewBuffer()

	if t1 {
		g1.WriteUint32(d.ReadUint32())
	} else {
		g1.WriteUint32(0x0000)
	}

	if t2 {
		g1.WriteUint32(d.ReadUint32())
	} else {
		g1.WriteUint32(0x0000)
	}

	g := guid.Classic(g1.ReadUint64())
	return g, nil
}

func DecodeValuesClassic(objectGUID guid.GUID, version uint32, in *etc.Buffer) (*ValuesBlock, error) {
	// Search for values descriptor for this version of the game
	descriptor := Descriptors[version]
	if descriptor == nil {
		return nil, fmt.Errorf("update: could not find descriptor data for version %d", version)
	}

	decoder := new(ValuesDecoder)
	decoder.Version = version

	// Will hold all the raw 4 byte data chunks stored in the update field dictionary.
	var dataBlocks [][]byte

	// setup values block
	valuesBlock := new(ValuesBlock)
	valuesBlock.Values = make(map[Global]interface{})
	decoder.ValuesBlock = valuesBlock

	// Load the bitmask. This contains an array of bits, if a bit == 1, then that bit corresponds to a 4-byte chunk that is being updated.
	bitmaskSize := int(in.ReadByte())
	bitmask := make([]uint32, bitmaskSize)

	for x := 0; x < bitmaskSize; x++ {
		bitmask[x] = in.ReadUint32()
	}

	var enabledOffsets []uint32

	for fieldOffsetBase, fieldIndex := range bitmask {
		for bitIndex := uint32(0); bitIndex < 32; bitIndex++ {
			if fieldIndex&(1<<bitIndex) != 0 {
				enabledOffsets = append(enabledOffsets, ((uint32(fieldOffsetBase) * 32) + bitIndex))
			}
		}
	}

	dataBlocks = make([][]byte, len(enabledOffsets))

	// In classic mode, each bit enabled in the bitmask has a corresponding 4-byte chunk.
	// Even if there is a problem with parsing the chunks, the other material in the packet will be untouched.
	for x := 0; x < len(enabledOffsets); x++ {
		dataBlocks[x] = in.ReadBytes(4)
	}

	decoder.Buffer = etc.FromBytes(bytes.Join(dataBlocks, nil))

	className, err := getClassName(objectGUID)
	if err != nil {
		return nil, err
	}

	class, err := descriptor.GetClass(className)
	if err != nil {
		return nil, err
	}

	fields := class.ExtractAllClassFields()

	decoder.EnabledOffsets = enabledOffsets
	decoder.OffsetIndex = 0

	// Post-mortem: let's scan for each offset individually, enabled or otherwise.
offsetScan:
	for offset := uint32(0); offset < class.EndOffset(); {
		field, err := queryFieldByAbsoluteOffset(fields, offset)
		if err != nil {
			// should not occur, fiddle with EndOffset to fix
			fmt.Println(err)
			offset++
			continue
		}

		if field.FieldType == Pad {
			if decoder.offsetEnabled(offset) {
				return nil, fmt.Errorf("update: padding block enabled")
			}
			offset++
			continue
		}

		toggled := decoder.offsetEnabled(offset)

		switch field.FieldType {
		case Uint32:
			if !toggled {
				offset++
				continue offsetScan
			}

			decoder.ValuesBlock.Values[field.Global] = decoder.ReadUint32()
		case Int32:
			if !toggled {
				offset++
				continue offsetScan
			}

			decoder.ValuesBlock.Values[field.Global] = decoder.ReadInt32()
		case Float32:
			if !toggled {
				offset++
				continue offsetScan
			}

			decoder.ValuesBlock.Values[field.Global] = decoder.ReadFloat32()
			// yo.Fatal(field.Global, decoder.ValuesBlock.Values[field.Global])
		case GUID:
			if toggled || decoder.offsetEnabled(offset+1) {
				g, err := decoder.decodeGUID(offset)
				if err != nil {
					return nil, err
				}

				decoder.ValuesBlock.Values[field.Global] = g
			}
			offset += 2
			continue
		case Uint32Array:
			u32 := make([]*uint32, field.SliceSize)

			anyFieldsEnabled := false // don't bother including in valuesblock if none are toggled

			for x := int64(0); x < field.SliceSize; x++ {
				off := uint32(x) + offset
				if decoder.offsetEnabled(off) {
					anyFieldsEnabled = true
					value := decoder.ReadUint32()
					u32[int(x)] = &value
				}
			}

			if anyFieldsEnabled {
				decoder.ValuesBlock.Values[field.Global] = u32
			}

			offset += uint32(field.SliceSize)
			continue offsetScan
		case Int32Array:
			i32 := make([]*int32, field.SliceSize)

			anyFieldsEnabled := false // don't bother including in valuesblock if none are toggled

			for x := int64(0); x < field.SliceSize; x++ {
				off := offset + uint32(x)
				if decoder.offsetEnabled(off) {
					anyFieldsEnabled = true
					value := decoder.ReadInt32()
					i32[int(x)] = &value
				}
			}

			if anyFieldsEnabled {
				decoder.ValuesBlock.Values[field.Global] = i32
			}

			offset += uint32(field.SliceSize)
			continue offsetScan
		case Float32Array:
			f := make([]*float32, field.SliceSize)

			anyFieldsEnabled := false

			for x := int64(0); x < field.SliceSize; x++ {
				if decoder.offsetEnabled(offset + uint32(x)) {
					float := decoder.ReadFloat32()
					f[x] = &float
					anyFieldsEnabled = true
				}
			}

			if anyFieldsEnabled {
				decoder.ValuesBlock.Values[field.Global] = f
			}

			offset += uint32(field.SliceSize)
			continue offsetScan
		case GUIDArray:
			g := make([]*guid.GUID, field.SliceSize)

			anyFieldsEnabled := false // don't bother including in valuesblock if none are toggled

			for x := int64(0); x < field.SliceSize; x++ {
				off := offset + (uint32(x) * 2)
				if decoder.offsetEnabled(off) || decoder.offsetEnabled(off+1) {
					gi, err := decoder.decodeGUID(off)
					if err != nil {
						return nil, err
					}
					g[int(x)] = &gi
					anyFieldsEnabled = true
				}
			}

			if anyFieldsEnabled {
				decoder.ValuesBlock.Values[field.Global] = g
			}

			offset += (uint32(field.SliceSize) * 2)
			continue offsetScan
		case Uint8:
			if toggled {
				bytes := decoder.ReadBytes(4)

				fieldIndex := -1
				for idx, bfield := range fields {
					if bfield.Global == field.Global {
						fieldIndex = idx
					}
				}

				for idx := fieldIndex; idx <= len(fields); idx++ {
					f := fields[idx]
					if f.FieldType != Uint8 {
						break
					}
					bIndex := idx - fieldIndex
					if bIndex == 4 {
						break
					}
					// TODO: apply type decorations
					decoder.ValuesBlock.Values[f.Global] = bytes[bIndex]
				}
			}
		case ArrayType:
			anyRowsEnabled := false

			ad := &ArrayData{}
			for _, v := range field.array.Fields {
				if v.Key != "" {
					ad.Cols = append(ad.Cols, v.Key)
				}
			}

			for x := int64(0); x < field.array.Len; x++ {
				row := []interface{}{}
				anyFieldsEnabled := false

				for _, f := range field.array.Fields {
					switch f.FieldType {
					case Pad:
						if decoder.offsetEnabled(offset) {
							panic("pad enabled")
						}
						offset++
					case Uint32:
						if decoder.offsetEnabled(offset) {
							anyFieldsEnabled = true
							row = append(row, decoder.ReadUint32())
						} else {
							row = append(row, nil)
						}
						offset++
					case Uint32Array:
						u32Array := make([]uint32, int(f.Len))
						for ux := int(0); ux < int(f.Len); ux++ {
							if decoder.offsetEnabled(offset) {
								anyFieldsEnabled = true
								u32Array[ux] = decoder.ReadUint32()
							}

							offset++
						}

						row = append(row, u32Array)
					case GUID:
						if decoder.offsetEnabled(offset) || decoder.offsetEnabled(offset+1) {
							g, err := decoder.decodeGUID(offset)
							if err != nil {
								return nil, err
							}

							anyFieldsEnabled = true

							row = append(row, g)
						} else {
							row = append(row, nil)
						}
						offset += 2
					default:
						return nil, fmt.Errorf("unknown field type %d", f.FieldType)
					}
				}

				if anyFieldsEnabled {
					anyRowsEnabled = true
				}
				ad.Rows = append(ad.Rows, row)
			}

			if anyRowsEnabled {
				valuesBlock.Values[field.Global] = ad
			}

			continue offsetScan
		}

		offset++
	}
	// switch field.FieldType {
	// case Uint32:
	// 	valuesBlock.Values[field.Global] = binary.LittleEndian.Uint32(dataBlocks[offset])
	// case Uint32Array:
	// 	slice := make([]*uint32, field.SliceSize)

	// 	for x := blockStart; x < int(field.SliceSize)-blockStart; x++ {
	// 		if decoder.offsetEnabled(offset + uint32(x)) {
	// 			u32 := binary.LittleEndian.Uint32(dataBlocks[offset])
	// 			slice[x] = &u32
	// 			decoder.OffsetIndex++
	// 		}
	// 	}

	// 	valuesBlock.Values[field.Global] = slice
	// case Float32:
	// 	valuesBlock.Values[field.Global] = math.Float32frombits(binary.LittleEndian.Uint32(dataBlocks[offset]))
	// case GUID:
	// 	if decoder.offsetEnabled(offset+1) == false {
	// 		return nil, fmt.Errorf("update: expected two blocks for GUID data")
	// 	}

	// 	offset2, _ := decoder.getOffset()

	// 	gbuf := etc.NewBuffer()
	// 	gbuf.Write(dataBlocks[offset])
	// 	gbuf.Write(dataBlocks[offset2])

	// 	g, err := guid.DecodeUnpacked(version, gbuf)
	// 	if err != nil {
	// 		return nil, err
	// 	}

	// 	valuesBlock.Values[field.Global] = g
	// }

	yo.Spew(valuesBlock)

	return valuesBlock, nil
}

func (d *ValuesDecoder) getOffset() (uint32, error) {
	if int(d.OffsetIndex) >= len(d.EnabledOffsets) {
		return 0, io.EOF
	}

	idx := d.OffsetIndex
	d.OffsetIndex++

	return d.EnabledOffsets[idx], nil
}

func (d *ValuesDecoder) offsetEnabled(offset uint32) bool {
	for _, off := range d.EnabledOffsets {
		if off == offset {
			return true
		}
	}

	return false
}

func (vb *ValuesBlock) WriteTo(objectGUID guid.GUID, e *Encoder) error {
	descriptor := Descriptors[e.Version]
	if descriptor == nil {
		return fmt.Errorf("update: WriteTo(): could not find descriptor data for version %d", e.Version)
	}

	className, err := getClassName(objectGUID)
	if err != nil {
		return err
	}

	class, err := descriptor.GetClass(className)
	if err != nil {
		return err
	}

	v := &ValuesEncoder{
		e,
		make(map[int][]byte),
	}

	fields := class.ExtractAllClassFields()

valueScan:
	for glob, value := range vb.Values {
		for _, field := range fields {
			if field.Global == glob {
				// Dont encode unchanged values,
				// except for uint8 values which must ALWAYS be included.
				// This is because if only changed bytes are included,
				// the client will interpret the zero bytes as a changed field.
				// MaNGOS handles this by storing the bytes in a continuous data block, and it would be hard to do this in Gophercraft.
				if field.FieldType != Uint8 {
					if v.enc.Mask&ValuesCreate == 0 && !vb.Changes[glob] {
						continue valueScan
					}
				}

				private := (v.enc.Mask&ValuesPrivate != 0)
				party := (v.enc.Mask&ValuesParty != 0)

				// dont encode private values
				if (field.Flags&Private != 0) && private == false {
					continue valueScan
				}

				// dont encode party only values
				if (field.Flags&Party != 0) && !(party || private) {
					continue valueScan
				}

				switch field.FieldType {
				case Int32:
					v.encodeUint32(uint32(field.AbsBlockOffset()), uint32(value.(int32)))
				case Uint32:
					v.encodeUint32(uint32(field.AbsBlockOffset()), value.(uint32))
				case Float32:
					_, ok := value.(float32)
					if !ok {
						panic(glob)
					}
					v.encodeFloat32(uint32(field.AbsBlockOffset()), value.(float32))
				case Uint8:
					v.encodeUint8(uint32(field.AbsBlockOffset()), field.BitOffset, value.(uint8))
				case GUID:
					v.encodeGUID(uint32(field.AbsBlockOffset()), value.(guid.GUID))
				case Uint32Array:
					u32 := value.([]*uint32)

					if len(u32) != int(field.SliceSize) {
						return fmt.Errorf("update: invalid uint32 pointer array passed in %s (should be %d, is %d)", glob, field.SliceSize, len(u32))
					}

					offsetBase := uint32(field.AbsBlockOffset())

					for x := 0; x < int(field.SliceSize); x++ {
						if u32[x] != nil {
							u := u32[x]
							v.encodeUint32(offsetBase+uint32(x), *u)
						}
					}
				case Int32Array:
					i32 := value.([]*int32)

					if len(i32) != int(field.SliceSize) {
						return fmt.Errorf("update: invalid int32 pointer array passed in %s (should be %d, is %d)", glob, field.SliceSize, len(i32))
					}

					offsetBase := uint32(field.AbsBlockOffset())

					for x := 0; x < int(field.SliceSize); x++ {
						if i32[x] != nil {
							i := i32[x]
							v.encodeUint32(offsetBase+uint32(x), uint32(*i))
						}
					}
				case Float32Array:
					f32 := value.([]*float32)

					if len(f32) != int(field.SliceSize) {
						return fmt.Errorf("update: invalid float32 pointer array passed in %s (should be %d, is %d)", glob, field.SliceSize, len(f32))
					}

					offsetBase := uint32(field.AbsBlockOffset())

					for x := 0; x < int(field.SliceSize); x++ {
						if f32[x] != nil {
							f := f32[x]
							v.encodeFloat32(offsetBase+uint32(x), *f)
						}
					}
				case GUIDArray:
					g := value.([]*guid.GUID)

					if len(g) != int(field.SliceSize) {
						return fmt.Errorf("update: invalid GUID pointer array passed in %s (should be %d, is %d)", glob, field.SliceSize, len(g))
					}

					offsetBase := uint32(field.AbsBlockOffset())

					for x := 0; x < int(field.SliceSize); x++ {
						if g[x] != nil {
							el := g[x]
							v.encodeGUID(offsetBase+(uint32(x)*2), *el)
						}
					}
				case ArrayType:
					offsetBase := uint32(field.AbsBlockOffset())
					offset := offsetBase

					ad := value.(*ArrayData)

					// validate input data
					for i, v := range field.array.Fields {
						if v.FieldType != Pad {
							if v.Key != ad.Cols[i] {
								return fmt.Errorf("update: field column names are mismatched. Consult packet/update/x_descriptor_%d.go's convention for %s to properly format your array data", e.Version, glob)
							}
						}
					}

					for x := int64(0); x < field.array.Len; x++ {
						var row []interface{}
						if int(x) < len(ad.Rows) {
							row = ad.Rows[x]
						}

						for i, vfield := range field.array.Fields {
							if vfield.FieldType != Pad {
								if row != nil {
									switch vfield.FieldType {
									case GUID:
										v.encodeGUID(offset, row[i].(guid.GUID))
										offset += 2
									case Uint32:
										v.encodeUint32(offset, row[i].(uint32))
										offset++
									case Float32:
										v.encodeFloat32(offset, row[i].(float32))
										offset++
									case Uint32Array:
										arrayVal := row[i].([]uint32)
										for arrayValIndex := int64(0); arrayValIndex < vfield.Len; arrayValIndex++ {
											v.encodeUint32(offset, arrayVal[int(arrayValIndex)])
											offset++
										}
									default:
										panic(vfield.FieldType.String())
									}
								}
							}
						}
					}
				default:
					panic(field.FieldType)
				}
			}
		}
	}

	// Build bitmask
	keys := []int{}

	for k := range v.Blocks {
		keys = append(keys, k)
	}

	sort.Ints(keys)

	yo.Spew(keys)

	maximum := uint32(keys[len(keys)-1])
	maskLen := ((maximum + 31) / 32)
	mask := make([]uint32, maskLen)
	for dwIndex := uint32(0); dwIndex < maskLen; dwIndex++ {
		for bitIndex := uint32(0); bitIndex < 32; bitIndex++ {
			offset := (dwIndex * 32) + bitIndex
			if v.Blocks[int(offset)] != nil {
				mask[dwIndex] |= (1 << bitIndex)
			}
		}
	}
	e.WriteByte(uint8(maskLen))
	for _, msk := range mask {
		e.WriteUint32(msk)
	}

	// Build content
	for _, key := range keys {
		block := v.Blocks[key]
		if len(block) != 4 {
			panic("block len invalid")
		}
		e.Write(v.Blocks[key])
	}

	return nil
}

func (v *ValuesEncoder) encodeGUID(offset uint32, g guid.GUID) {
	e := etc.NewBuffer()
	g.EncodeUnpacked(v.enc.Version, e)

	p1 := e.ReadUint32()
	p2 := e.ReadUint32()

	v.encodeUint32(offset, p1)
	v.encodeUint32(offset+1, p2)
}

func (ve *ValuesEncoder) encodeUint32(offset, v uint32) {
	out := make([]byte, 4)
	binary.LittleEndian.PutUint32(out, v)
	ve.Blocks[int(offset)] = out
}

func (ve *ValuesEncoder) encodeFloat32(offset uint32, v float32) {
	ve.encodeUint32(offset, math.Float32bits(v))
}

func (ve *ValuesEncoder) encodeUint8(offset uint32, bitOffset, v uint8) {
	k := int(offset)
	if ve.Blocks[k] == nil {
		ve.Blocks[k] = make([]byte, 4)
	}

	byteOffset := bitOffset / 8

	ve.Blocks[k][byteOffset] = v
}
