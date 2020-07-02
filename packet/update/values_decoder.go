package update

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"reflect"

	"github.com/superp00t/etc/yo"
	"github.com/superp00t/gophercraft/guid"
)

var (
	ErrNotAStruct = errors.New("update: not a struct")

	// types
	guidType     = reflect.TypeOf(guid.GUID{})
	bitPadType   = reflect.TypeOf(BitPad{})
	bytePadType  = reflect.TypeOf(BytePad{})
	chunkPadType = reflect.TypeOf(ChunkPad{})
)

type ValuesDecoder struct {
	Decoder          *Decoder
	CurrentBitmask   *Bitmask
	CurrentChunk     [4]byte
	ChunkPos, BitPos uint32
}

// Fwd the decoder's chunk offset forward by n chunks
func (valdec *ValuesDecoder) Fwd(n uint32) {
	valdec.ChunkPos += n
	valdec.BitPos = 0
}

// FwdBits move the decoder's bit offset by n bits
func (valdec *ValuesDecoder) FwdBits(n uint32) {
	valdec.BitPos += n
	if valdec.BitPos >= 32 {
		valdec.BitPos = 0
		valdec.ChunkPos++
	}
}

// FwdBits move the decoder's byte offset by n bytes
func (valdec *ValuesDecoder) FwdBytes(n uint32) {
	if n > 4 {
		panic("use Fwd instead")
	}

	if valdec.BitPos%8 != 0 {
		valdec.BitPos += 8 - (valdec.BitPos % 8)
	} else {
		valdec.BitPos += 8
	}

	if valdec.BitPos == 32 {
		valdec.BitPos = 0
		valdec.ChunkPos++
	}

	if valdec.BitPos > 32 {
		panic("how did this happen?")
	}
}

func (valdec *ValuesDecoder) ReadGUID(to reflect.Value) error {
	if valdec.Decoder.Descriptor.DescriptorOptions&DescriptorOptionClassicGUIDs != 0 {
		// Use legacy format
		var guidData [8]byte

		// Get existing GUID data for this field. It doesn't matter if it's empty.
		currentGUID := to.Interface().(guid.GUID)

		binary.LittleEndian.PutUint64(guidData[:], currentGUID.Classic())

		// It is common practice to omit the high GUID chunk for players, as the legacy GUID format for players is zero.
		// The client already has this value as zero upon creation, so it makes sense to omit it.

		hiEnabled := valdec.CurrentBitmask.Enabled(valdec.ChunkPos)
		loEnabled := valdec.CurrentBitmask.Enabled(valdec.ChunkPos + 1)

		if hiEnabled {
			valdec.Decoder.Reader.Read(guidData[0:4])
		}

		if loEnabled {
			valdec.Decoder.Reader.Read(guidData[4:8])
		}

		newGUID := guid.Classic(binary.LittleEndian.Uint64(guidData[:]))
		to.Set(reflect.ValueOf(newGUID))
		valdec.Fwd(2)
	} else {
		// use modern GUID format.
		panic("nyi: modern GUID format")
	}
	return nil
}

func (valdec *ValuesDecoder) ReadUint32(to reflect.Value) error {
	if valdec.CurrentBitmask.Enabled(valdec.ChunkPos) {
		var chunk [4]byte
		valdec.Decoder.Reader.Read(chunk[:])
		to.Set(reflect.ValueOf(binary.LittleEndian.Uint32(chunk[:])))
	}

	valdec.Fwd(1)
	return nil
}

func (valdec *ValuesDecoder) ReadInt32(to reflect.Value) error {
	if valdec.CurrentBitmask.Enabled(valdec.ChunkPos) {
		var chunk [4]byte
		valdec.Decoder.Reader.Read(chunk[:])
		to.Set(reflect.ValueOf(int32(binary.LittleEndian.Uint32(chunk[:]))))
	}

	valdec.Fwd(1)
	return nil
}

func (valdec *ValuesDecoder) ReadFloat32(to reflect.Value) error {
	if valdec.CurrentBitmask.Enabled(valdec.ChunkPos) {
		var chunk [4]byte
		valdec.Decoder.Reader.Read(chunk[:])
		to.Set(reflect.ValueOf(float32(math.Float32frombits(binary.LittleEndian.Uint32(chunk[:])))))
	}

	valdec.Fwd(1)
	return nil
}

func (valdec *ValuesDecoder) BytePos() int {
	return int(valdec.BitPos / 8)
}

func (valdec *ValuesDecoder) FillCurrentChunk() {
	if valdec.BitPos == 0 && valdec.CurrentBitmask.Enabled(valdec.ChunkPos) {
		if _, err := valdec.Decoder.Reader.Read(valdec.CurrentChunk[:]); err != nil {
			panic(err)
		}
	}
}

func (valdec *ValuesDecoder) ReadByte(to reflect.Value) error {
	// TODO: In the new protocol, bytes are considered fields unto themselves.
	if valdec.CurrentBitmask.Enabled(valdec.ChunkPos) {
		valdec.FillCurrentChunk()

		to.Set(reflect.ValueOf(valdec.CurrentChunk[valdec.BytePos()]))
	}

	valdec.FwdBytes(1)
	return nil
}

func (valdec *ValuesDecoder) ReadBit(to reflect.Value) error {
	if valdec.CurrentBitmask.Enabled(valdec.ChunkPos) {
		valdec.FillCurrentChunk()

		bit8offset := valdec.BitPos % 8
		isToggled := valdec.CurrentChunk[valdec.BytePos()]&(1<<bit8offset) != 0
		yo.Spew(valdec.CurrentChunk)
		to.SetBool(isToggled)
	}

	valdec.FwdBits(1)
	return nil
}

func isSubChunkType(value reflect.Value) bool {
	if value.Kind() == reflect.Bool {
		return true
	}

	if value.Kind() == reflect.Uint8 {
		return true
	}

	if value.Type() == bytePadType {
		return true
	}

	if value.Type() == bitPadType {
		return true
	}

	return false
}

func isUint8Type(value reflect.Value) bool {
	if value.Type() == bytePadType {
		return true
	}

	if value.Kind() == reflect.Uint8 {
		return true
	}

	return false
}

func (decoder *Decoder) DecodeValuesBlockData(valuesBlock *ValuesBlock) error {
	var err error
	valuesBlock.ChangeMask, err = ReadBitmask(decoder.Descriptor, decoder.Reader)
	if err != nil {
		return err
	}

	fmt.Println(valuesBlock.ChangeMask)

	valDec := new(ValuesDecoder)
	valDec.Decoder = decoder
	valDec.CurrentBitmask = valuesBlock.ChangeMask

	// A storage struct has not yet been set.
	if !valuesBlock.StorageDescriptor.IsValid() {
		// All objects have an ObjectData field. This contains data necessary to parse the rest of the update stream.
		objectType := decoder.Descriptor.ObjectDescriptors[guid.TypeMaskObject]
		// Create an ObjectData instance.
		objectInstance := reflect.New(objectType)
		if err := valDec.Decode(objectInstance, "ObjectData"); err != nil {
			return err
		}

		var name string

		// Figure out what type this object has.
		typeMask, err := guid.ResolveTypeMask(valDec.Decoder.Build, uint32(objectInstance.Elem().FieldByName("Type").Uint()))
		if err != nil {
			return err
		}

		objectDescriptor := valDec.Decoder.Descriptor.ObjectDescriptors[typeMask]
		if objectDescriptor == nil {
			yo.Spew(objectInstance.Interface())
			return fmt.Errorf("update: no object descriptor for the received typemask: %s", typeMask)
		}

		if name == "" {
			name = objectDescriptor.Name()
		}

		// Now we are building the rest of the structure.
		fullObject := reflect.New(objectDescriptor)
		fullObject.Elem().FieldByName("ObjectData").Set(objectInstance.Elem())

		for x := 1; x < objectDescriptor.NumField(); x++ {
			if err := valDec.Decode(
				fullObject.Elem().Field(x),
				name+"."+fullObject.Elem().Type().Field(x).Name); err != nil {
				return err
			}
		}

		valuesBlock.StorageDescriptor = fullObject

		return nil
	}

	return valDec.Decode(
		valuesBlock.StorageDescriptor,
		"")
}

func (valdec *ValuesDecoder) Decode(field reflect.Value, name string) error {
	if field.Kind() == reflect.Ptr {
		field = field.Elem()
	}

	// Are we in the middle of reading a sub-chunk value?
	if valdec.BitPos != 0 {
		// If we're halfway through reading a byte,
		// Move the reader to the next byte.
		if valdec.BitPos%8 != 0 && field.Kind() == reflect.Uint8 {
			valdec.FwdBytes(1)
		}

		// can this type be resolved smaller than a chunk? Some chunks can contain multiple update fields. (in the format of bits (bool) and bytes (byte/uint8))
		if !isSubChunkType(field) {
			// reset the bit offset, and advance 1 chunk forward. This ensures proper alignment after a sub-chunk read.
			valdec.Fwd(1)

			if field.Type() == chunkPadType {
				return nil
			}
		}
	}

	if field.Type() == chunkPadType {
		if valdec.CurrentBitmask.Enabled(valdec.ChunkPos) {
			return fmt.Errorf("update: padding chunk for %s was enabled in Bitmask: check the descriptor and give this a proper type.", name)
		}
		valdec.Fwd(1)
		return nil
	}

	// This field is simply telling us to move on to the next bit offset.
	if field.Type() == bitPadType {
		valdec.FillCurrentChunk()
		valdec.FwdBits(1)
		return nil
	}

	if field.Type() == bytePadType {
		valdec.FillCurrentChunk()
		valdec.FwdBytes(1)
		return nil
	}

	if field.Type() == guidType {
		return valdec.ReadGUID(field)
	}

	if field.Kind() == reflect.Struct {
		for f := 0; f < field.NumField(); f++ {
			sfield := field.Field(f)
			if err := valdec.Decode(sfield, name+"."+field.Type().Field(f).Name); err != nil {
				return err
			}
		}

		return nil
	}

	switch field.Kind() {
	case reflect.Int32:
		return valdec.ReadInt32(field)
	case reflect.Uint32:
		return valdec.ReadUint32(field)
	case reflect.Float32:
		return valdec.ReadFloat32(field)
	case reflect.Array:
		for i := 0; i < field.Len(); i++ {
			if err := valdec.Decode(field.Index(i), fmt.Sprintf("%s[%d]", name, i)); err != nil {
				return err
			}
		}
		return nil
	case reflect.Bool:
		return valdec.ReadBit(field)
	case reflect.Uint8:
		return valdec.ReadByte(field)
	default:
		panic(fmt.Errorf("update: unrecognized type in %s: %s", name, field.Type()))
	}

	panic("unreachable")
}
