package update

import (
	"encoding/binary"
	"fmt"
	"reflect"

	"github.com/superp00t/gophercraft/guid"
)

type VisibilityFlags uint32

const (
	None        VisibilityFlags = 0
	Owner       VisibilityFlags = 0x01
	PartyMember VisibilityFlags = 0x02
	UnitAll     VisibilityFlags = 0x04
	Empath      VisibilityFlags = 0x08
)

type ValuesEncoder struct {
	Encoder        *Encoder
	ViewMask       VisibilityFlags
	Create         bool
	ValuesBlock    *ValuesBlock
	CurrentBitmask *Bitmask
	ChunkPos       uint32
	BitPos         uint32
	NextChunk      [4]byte
}

func (valenc *ValuesEncoder) SetCreateBits() error {
	value := valenc.ValuesBlock.StorageDescriptor.Elem()

	if err := valenc.setCreateBitsFor(value, ""); err != nil {
		return err
	}

	// panic(valenc.CurrentBitmask)

	valenc.BitPos = 0
	valenc.ChunkPos = 0
	valenc.clearNextChunk()

	return nil
}

func (valenc *ValuesEncoder) clearNextChunk() {
	copy(valenc.NextChunk[:], make([]byte, 4))
}

func (valenc *ValuesEncoder) includeValue(tag FieldTag) bool {
	if tag.IsPrivate() {
		if valenc.ViewMask&Owner != 0 {
			return true
		}
		return false
	}

	if tag.IsParty() {
		if valenc.ViewMask&PartyMember != 0 {
			return true
		}
		return false
	}

	return true
}

// this function is purely for setting the bitmask for new objects in the legacy protocol.
func (valenc *ValuesEncoder) setCreateBitsFor(value reflect.Value, tag FieldTag) error {
	if !isSubChunkType(value) && valenc.BitPos > 0 {
		// We've moved on to the next value, so commit chunk
		valenc.BitPos = 0
		valenc.ChunkPos++

		if value.Type() == chunkPadType {
			return nil
		}
	} else {
		if valenc.BitPos >= 32 {
			valenc.BitPos = 0
			valenc.ChunkPos++

			if value.Type() == chunkPadType {
				return nil
			}
		}
	}

	switch value.Type() {
	case bitPadType:
		valenc.BitPos++
		return nil
	case bytePadType:
		if valenc.BitPos%8 != 0 {
			valenc.BitPos += (8 - valenc.BitPos%8)
		} else {
			valenc.BitPos += 8
		}
		return nil
	case chunkPadType:
		valenc.BitPos = 0
		valenc.ChunkPos++
		return nil
	case alignPadType:
		valenc.BitPos = 0
		valenc.CurrentBitmask.Set(valenc.ChunkPos, true)
		valenc.ChunkPos++
		return nil
	case guidType:
		if valenc.includeValue(tag) {
			id := value.Interface().(guid.GUID)
			var bytes [8]byte
			binary.LittleEndian.PutUint64(bytes[:], id.Classic())
			if u32(bytes[:4]) != 0 {
				valenc.CurrentBitmask.Set(valenc.ChunkPos, true)
			}

			if u32(bytes[4:]) != 0 {
				valenc.CurrentBitmask.Set(valenc.ChunkPos+1, true)
			}
		}

		valenc.ChunkPos += 2
		valenc.BitPos = 0
		return nil
	}

	switch value.Kind() {
	case reflect.Struct:
		for x := 0; x < value.NumField(); x++ {
			if err := valenc.setCreateBitsFor(value.Field(x), FieldTag(value.Type().Field(x).Tag.Get("update"))); err != nil {
				return err
			}
		}
		return nil
	case reflect.Array:
		for x := 0; x < value.Len(); x++ {
			if err := valenc.setCreateBitsFor(value.Index(x), tag); err != nil {
				return err
			}
		}
	case reflect.Uint64:
		if valenc.includeValue(tag) {
			var bytes [8]byte
			binary.LittleEndian.PutUint64(bytes[:], value.Uint())
			if u32(bytes[:4]) != 0 {
				valenc.CurrentBitmask.Set(valenc.ChunkPos, true)
			}

			if u32(bytes[4:]) != 0 {
				valenc.CurrentBitmask.Set(valenc.ChunkPos+1, true)
			}
		}
		valenc.ChunkPos += 2
		valenc.BitPos = 0
	case reflect.Uint32:
		if value.Uint() != 0 && valenc.includeValue(tag) {
			valenc.CurrentBitmask.Set(valenc.ChunkPos, true)
		}
		valenc.ChunkPos++
		valenc.BitPos = 0
	case reflect.Int32:
		if value.Int() != 0 && valenc.includeValue(tag) {
			valenc.CurrentBitmask.Set(valenc.ChunkPos, true)
		}
		valenc.ChunkPos++
		valenc.BitPos = 0
	case reflect.Float32:
		valenc.BitPos = 0
		if value.Float() != 0 && valenc.includeValue(tag) {
			valenc.CurrentBitmask.Set(valenc.ChunkPos, true)
		}
		valenc.ChunkPos++
	case reflect.Bool:
		if valenc.includeValue(tag) {
			if value.Bool() {
				valenc.CurrentBitmask.Set(valenc.ChunkPos, true)
			}
		}
		valenc.BitPos++
	case reflect.Uint8:
		if valenc.BitPos%8 != 0 {
			valenc.BitPos += (8 - valenc.BitPos%8)
		}
		if valenc.BitPos == 32 {
			valenc.BitPos = 0
			valenc.ChunkPos++
		}
		if valenc.includeValue(tag) {
			if value.Uint() != 0 {
				valenc.CurrentBitmask.Set(valenc.ChunkPos, true)
			}
		}

		valenc.BitPos += 8
	case reflect.Uint16:
		if valenc.BitPos == 32 {
			valenc.BitPos = 0
			valenc.ChunkPos++
		}
		if valenc.includeValue(tag) {
			if value.Uint() != 0 {
				valenc.CurrentBitmask.Set(valenc.ChunkPos, true)
			}
		}
		valenc.BitPos += 16
	default:
		return fmt.Errorf("update: unhandled type detected while trying to write creation bitmask: %s", value.Type())
	}

	return nil
}

func (valenc *ValuesEncoder) EncodeValue(value reflect.Value, name string, tag FieldTag) error {
	quit := true

	if value.Type() == chunkPadType {
		// Sometimes update.ChunkPad is used to terminate a multi-field chunk.
		if valenc.BitPos > 0 {
			if valenc.CurrentBitmask.Enabled(valenc.ChunkPos) {
				valenc.Encoder.Write(valenc.NextChunk[:])
			}
			valenc.ChunkPos++
			valenc.BitPos = 0
			return nil
		}
		// Otherwise, it is merely treated as an empty chunk.
	}

	if !isSubChunkType(value) && valenc.BitPos > 0 || valenc.BitPos >= 32 {
		if valenc.CurrentBitmask.Enabled(valenc.ChunkPos) {
			valenc.Encoder.Write(valenc.NextChunk[:])
		}
		valenc.clearNextChunk()
		valenc.BitPos = 0
		valenc.ChunkPos++
	}

	// Uncomment to dump raw offsets
	// fmt.Printf("%s 0x%04X\n", name, valenc.ChunkPos)

	switch value.Type() {
	case guidType:
		var bytes [8]byte
		binary.LittleEndian.PutUint64(bytes[:], value.Interface().(guid.GUID).Classic())

		if valenc.CurrentBitmask.Enabled(valenc.ChunkPos) {
			if _, err := valenc.Encoder.Write(bytes[:4]); err != nil {
				return err
			}
		}

		if valenc.CurrentBitmask.Enabled(valenc.ChunkPos + 1) {
			if _, err := valenc.Encoder.Write(bytes[4:]); err != nil {
				return err
			}
		}

		valenc.ChunkPos += 2
		valenc.BitPos = 0
	case chunkPadType:
		if valenc.CurrentBitmask.Enabled(valenc.ChunkPos) {
			return fmt.Errorf("update: chunk padding %s is toggled", name)
		}

		valenc.ChunkPos++
		valenc.BitPos = 0
	case alignPadType:
		if valenc.CurrentBitmask.Enabled(valenc.ChunkPos) {
			if err := writeUint32(valenc.Encoder, 0x00000000); err != nil {
				return err
			}
		}
		valenc.ChunkPos++
		valenc.BitPos = 0
	case bitPadType:
		valenc.BitPos++
	case bytePadType:
		if valenc.BitPos%8 != 0 {
			valenc.BitPos += (8 - valenc.BitPos%8)
		} else {
			valenc.BitPos += 8
		}
	default:
		quit = false
	}

	if quit {
		return nil
	}

	switch value.Kind() {
	case reflect.Uint16:
		if valenc.CurrentBitmask.Enabled(valenc.ChunkPos) {
			binary.LittleEndian.PutUint16(valenc.NextChunk[valenc.BitPos/8:], uint16(value.Uint()))
		}
		valenc.BitPos += 16
	case reflect.Bool:
		if valenc.CurrentBitmask.Enabled(valenc.ChunkPos) {
			fmt.Printf("%s 0x%08X\n", name, (1 << valenc.BitPos))
			if value.Bool() {
				valenc.NextChunk[valenc.BitPos/8] |= (1 << (valenc.BitPos % 8))
			} else {
				valenc.NextChunk[valenc.BitPos/8] &= ^(1 << (valenc.BitPos % 8))
			}
		}
		valenc.BitPos++
	case reflect.Uint8:
		if valenc.BitPos%8 != 0 {
			valenc.BitPos += (8 - valenc.BitPos%8)
		}
		if valenc.CurrentBitmask.Enabled(valenc.ChunkPos) {
			valenc.NextChunk[valenc.BitPos/8] = uint8(value.Uint())
		}
		valenc.BitPos += 8
	case reflect.Array:
		for x := 0; x < value.Len(); x++ {
			if err := valenc.EncodeValue(value.Index(x), fmt.Sprintf("%s[%d]", name, x), tag); err != nil {
				return err
			}
		}
	case reflect.Struct:
		for x := 0; x < value.NumField(); x++ {
			nm := name + "." + value.Type().Field(x).Name
			if err := valenc.EncodeValue(value.Field(x), nm, FieldTag(value.Type().Field(x).Tag.Get("update"))); err != nil {
				return err
			}
		}
	case reflect.Uint64:
		var bytes [8]byte
		binary.LittleEndian.PutUint64(bytes[:], value.Uint())

		if valenc.CurrentBitmask.Enabled(valenc.ChunkPos) {
			if _, err := valenc.Encoder.Write(bytes[0:4]); err != nil {
				return err
			}
		}

		if valenc.CurrentBitmask.Enabled(valenc.ChunkPos + 1) {
			if _, err := valenc.Encoder.Write(bytes[0:4]); err != nil {
				return err
			}
		}
		valenc.ChunkPos += 2
		valenc.BitPos = 0
	case reflect.Uint32:
		if valenc.CurrentBitmask.Enabled(valenc.ChunkPos) {
			if err := writeUint32(valenc.Encoder, uint32(value.Uint())); err != nil {
				return err
			}
		}
		valenc.ChunkPos++
		valenc.BitPos = 0
	case reflect.Int32:
		if valenc.CurrentBitmask.Enabled(valenc.ChunkPos) {
			if err := writeInt32(valenc.Encoder, int32(value.Int())); err != nil {
				return err
			}
		}
		valenc.ChunkPos++
		valenc.BitPos = 0
	case reflect.Float32:
		if valenc.CurrentBitmask.Enabled(valenc.ChunkPos) {
			if err := writeFloat32(valenc.Encoder, float32(value.Float())); err != nil {
				return err
			}
		}
		valenc.ChunkPos++
		valenc.BitPos = 0
	default:
		return fmt.Errorf("update: unhandled attempt to encode %s kind %s", value.Type(), value.Kind())
	}

	if valenc.BitPos == 32 {
		if valenc.CurrentBitmask.Enabled(valenc.ChunkPos) {
			valenc.Encoder.Write(valenc.NextChunk[:])
		}
		valenc.clearNextChunk()
		valenc.BitPos = 0
		valenc.ChunkPos++
	}

	return nil
}

func (valuesBlock *ValuesBlock) WriteData(e *Encoder, viewMask VisibilityFlags, create bool) error {
	valenc := &ValuesEncoder{
		Encoder:     e,
		Create:      create,
		ValuesBlock: valuesBlock,
		ViewMask:    viewMask,
	}

	if !valuesBlock.StorageDescriptor.IsValid() {
		return fmt.Errorf("update: cannot encode empty ValuesBlock.StorageDescriptor")
	}

	if valenc.Create {
		// TODO: in the future, all fields are included in the create block (without a bitmask)

		// All non-zero and non-private chunks will be included.
		valenc.CurrentBitmask = NewBitmask()
		if err := valenc.SetCreateBits(); err != nil {
			return err
		}
	} else {
		// All fields enabled in the change mask will be included, even zero and private ones.
		valenc.CurrentBitmask = valuesBlock.ChangeMask
	}

	// TODO: in the future, bitmasks are stored within the descriptor's structs.
	// valenc.CurrentMask will be an alias for these.

	// Write uint8 len + uint32[len]
	if err := WriteBitmask(valenc.CurrentBitmask, e.Descriptor, e); err != nil {
		return err
	}

	// Write uint32 blocks, the size of which is the number of true bits in the bitmask
	if err := valenc.EncodeValue(valuesBlock.StorageDescriptor.Elem(), valuesBlock.StorageDescriptor.Type().String(), ""); err != nil {
		return err
	}

	if valenc.BitPos > 0 {
		if valenc.CurrentBitmask.Enabled(valenc.ChunkPos) {
			valenc.Encoder.Write(valenc.NextChunk[:])
		}
		valenc.clearNextChunk()
		valenc.BitPos = 0
		valenc.ChunkPos++
	}

	return nil
}
