package update

import (
	"fmt"
	"reflect"

	"github.com/superp00t/gophercraft/guid"
	"github.com/superp00t/gophercraft/vsn"
)

func nilValue() reflect.Value {
	var nilI interface{}
	return reflect.ValueOf(nilI)
}

func isNil(value reflect.Value) bool {
	if value.IsValid() == false {
		return true
	}

	if value.Kind() == reflect.Interface && value.IsNil() {
		return true
	}
	return false
}

func NewValuesBlock(build vsn.Build, mask guid.TypeMask) (*ValuesBlock, error) {
	descriptor, ok := Descriptors[build]
	if !ok {
		return nil, fmt.Errorf("update: can not find descriptor for %s", build)
	}

	storageDescriptorType := descriptor.ObjectDescriptors[mask]
	if storageDescriptorType == nil {
		return nil, fmt.Errorf("update: cannot find storage descriptor for type %s in descriptor %s", mask, build)
	}

	newStorageDescriptor := reflect.New(storageDescriptorType)

	vBlock := &ValuesBlock{
		TypeMask:          mask,
		Descriptor:        descriptor,
		ChangeMask:        NewBitmask(),
		StorageDescriptor: newStorageDescriptor,
	}

	typeUint, err := mask.Resolve(build)
	if err != nil {
		return nil, err
	}

	vBlock.SetUint32("Type", typeUint)

	return vBlock, nil
}

func (vb *ValuesBlock) ClearChanges() {
	vb.ChangeMask.Clear()
}

func (vb *ValuesBlock) ClearChangesAndUnlock() {
	vb.ClearChanges()
	vb.Unlock()
}

func (vb *ValuesBlock) findValueOffset(offset, bitOffset *uint32, value reflect.Value, currentKeys, targetKeys []interface{}) (reflect.Value, error) {
	// fmt.Println(currentKeys, "==", targetKeys)
	if reflect.DeepEqual(currentKeys, targetKeys) {
		return value, nil
	}

	quit := true

	switch value.Type() {
	case guidType:
		nxtChunk(offset, bitOffset)
		nxtChunk(offset, bitOffset)
	case bitPadType:
		nxtBit(offset, bitOffset)
	case bytePadType:
		nxtByte(offset, bitOffset)
	case chunkPadType:
		nxtChunk(offset, bitOffset)
	default:
		quit = false
	}

	if quit {
		return nilValue(), nil
	}

	switch value.Kind() {
	case reflect.Bool:
		nxtBit(offset, bitOffset)
	case reflect.Uint8:
		nxtByte(offset, bitOffset)
	case reflect.Uint32:
		nxtChunk(offset, bitOffset)
	case reflect.Int32:
		nxtChunk(offset, bitOffset)
	case reflect.Float32:
		nxtChunk(offset, bitOffset)
	case reflect.Array:
		for x := 0; x < value.Len(); x++ {
			val := value.Index(x)
			subVal, err := vb.findValueOffset(offset, bitOffset, val, append(currentKeys, x), targetKeys)
			if err != nil {
				return nilValue(), err
			}

			if !isNil(subVal) {
				return subVal, nil
			}
		}
		return nilValue(), nil
	case reflect.Struct:
		tp := value.Type()
		for x := 0; x < tp.NumField(); x++ {
			ftp := tp.Field(x)
			subVal, err := vb.findValueOffset(offset, bitOffset, value.Field(x), append(currentKeys, ftp.Name), targetKeys)
			if err != nil {
				return nilValue(), err
			}

			if !isNil(subVal) {
				return subVal, nil
			}
		}
		return nilValue(), nil
	default:
		return nilValue(), fmt.Errorf("update: unknown field during calculation to set bitmask %s", value.Kind())
	}

	return nilValue(), nil
}

func (vb *ValuesBlock) FindValueOffset(keys ...interface{}) (uint32, reflect.Value, error) {
	sdesc := vb.StorageDescriptor.Elem()
	var offset, bitOffset uint32
	for x := 0; x < sdesc.NumField(); x++ {
		dataStruct := sdesc.Field(x)
		value, err := vb.findValueOffset(&offset, &bitOffset, dataStruct, []interface{}{}, keys)
		if err != nil {
			return offset, value, err
		}

		if !isNil(value) {
			return offset, value, nil
		}
	}
	return 0, nilValue(), fmt.Errorf("could not find offset: %s", fmt.Sprintln(keys))
}

func indexValue(value reflect.Value, index interface{}) reflect.Value {
	switch i := index.(type) {
	case int:
		return value.Index(i)
	case string:
		return value.FieldByName(i)
	default:
		panic("unknown type")
	}
}

func (vb *ValuesBlock) SetStructArrayValue(glob string, index int, columnName string, value interface{}) {
	vb.Lock()
	offset, val, err := vb.FindValueOffset(glob, index, columnName)
	if err != nil {
		panic(err)
	}
	fmt.Println("found offset of", glob, index, offset, val.Addr().Pointer())
	val.Set(reflect.ValueOf(value))
	vb.ChangeMask.Set(offset, true)
	if val.Type() == guidType {
		vb.ChangeMask.Set(offset+1, true)
	}
	vb.Unlock()
}

// Ensure proper type inference by setting with these functions

func (vb *ValuesBlock) SetUint32(glob string, value uint32) {
	vb.Lock()
	offset, val, err := vb.FindValueOffset(glob)
	if err != nil {
		panic(err)
	}
	val.SetUint(uint64(value))
	vb.ChangeMask.Set(offset, true)
	vb.Unlock()
}

func (vb *ValuesBlock) SetGUID(glob string, value guid.GUID) {
	vb.Lock()
	offset, val, err := vb.FindValueOffset(glob)
	if err != nil {
		panic(err)
	}
	val.Set(reflect.ValueOf(value))
	vb.ChangeMask.Set(offset, true)
	vb.ChangeMask.Set(offset+1, true)
	vb.Unlock()
}

func (vb *ValuesBlock) SetFloat32(glob string, value float32) {
	vb.Lock()
	offset, val, err := vb.FindValueOffset(glob)
	if err != nil {
		panic(err)
	}
	val.SetFloat(float64(value))
	vb.ChangeMask.Set(offset, true)
	vb.Unlock()
}

func (vb *ValuesBlock) SetArrayValue(glob string, index int, value interface{}) {
	vb.Lock()
	offset, val, err := vb.FindValueOffset(glob, index)
	if err != nil {
		panic(err)
	}

	val.Set(reflect.ValueOf(value))

	vb.ChangeMask.Set(offset, true)
	if val.Type() == guidType {
		vb.ChangeMask.Set(offset+1, true)
	}

	// if val.Kind() != reflect.Array {
	// 	panic(val.Type().String() + " is not an array")
	// }

	// size := 1
	// if val.Index(0).Type() == guidType {
	// 	size = 2
	// }

	// actualOffset := offset + uint32(size*index)
	// for x := 0; x < size; x++ {
	// 	vb.ChangeMask.Set(actualOffset+uint32(x), true)
	// }

	// val.Index(index).Set(reflect.ValueOf(value))

	vb.Unlock()
}

func (vb *ValuesBlock) SetBit(glob string, value bool) {
	vb.Lock()
	offset, val, err := vb.FindValueOffset(glob)
	if err != nil {
		panic(err)
	}
	val.SetBool(value)
	vb.ChangeMask.Set(offset, true)
	vb.Unlock()
}

func (vb *ValuesBlock) SetByte(glob string, value uint8) {
	vb.Lock()
	offset, val, err := vb.FindValueOffset(glob)
	if err != nil {
		panic(err)
	}
	val.SetUint(uint64(value))
	vb.ChangeMask.Set(offset, true)
	vb.Unlock()
}

func (vb *ValuesBlock) SetInt32(glob string, value int32) {
	vb.Lock()
	offset, val, err := vb.FindValueOffset(glob)
	if err != nil {
		panic(err)
	}
	val.SetInt(int64(value))
	vb.ChangeMask.Set(offset, true)
	vb.Unlock()
}

func (vb *ValuesBlock) GetByte(keys ...interface{}) uint8 {
	vb.Lock()
	_, val, err := vb.FindValueOffset(keys...)
	if err != nil {
		panic(err)
	}
	bt := val.Uint()
	vb.Unlock()
	return uint8(bt)
}

func (vb *ValuesBlock) GetBit(keys ...interface{}) bool {
	vb.Lock()
	_, val, err := vb.FindValueOffset(keys...)
	if err != nil {
		panic(err)
	}
	bt := val.Bool()
	vb.Unlock()
	return bt
}

func (vb *ValuesBlock) SetGUIDArrayValue(glob string, index int, value guid.GUID) {
	vb.SetArrayValue(glob, index, value)
}

func (vb *ValuesBlock) SetFloat32ArrayValue(glob string, index int, value float32) {
	vb.SetArrayValue(glob, index, value)
}

func (vb *ValuesBlock) SetInt32ArrayValue(glob string, index int, value int32) {
	vb.SetArrayValue(glob, index, value)
}

func (vb *ValuesBlock) SetUint32ArrayValue(glob string, index int, value uint32) {
	vb.SetArrayValue(glob, index, value)
}

// Todo: Faster implementation would ignore the need to find offsets
func (vb *ValuesBlock) GetGUID(keys ...interface{}) guid.GUID {
	vb.Lock()

	_, val, err := vb.FindValueOffset(keys...)
	if err != nil {
		panic(err)
	}

	id := val.Interface().(guid.GUID)
	vb.Unlock()
	return id
}

func (vb *ValuesBlock) GetUint32(keys ...interface{}) uint32 {
	vb.Lock()

	_, val, err := vb.FindValueOffset(keys...)
	if err != nil {
		panic(err)
	}

	value := val.Interface().(uint32)
	vb.Unlock()
	return value
}

func (vb *ValuesBlock) GetFloat32(keys ...interface{}) float32 {
	vb.Lock()

	_, val, err := vb.FindValueOffset(keys...)
	if err != nil {
		panic(err)
	}

	value := val.Interface().(float32)
	vb.Unlock()
	return value
}

func (vb *ValuesBlock) Get(keys ...interface{}) reflect.Value {
	vb.Lock()

	_, val, err := vb.FindValueOffset(keys...)
	if err != nil {
		panic(err)
	}

	vb.Unlock()
	return val
}

func (vb *ValuesBlock) GetUint32Slice(keys ...interface{}) []uint32 {
	val := vb.Get(keys...)
	sli := val.Slice(0, val.Len())
	return sli.Interface().([]uint32)
}

// func parseString(runes []rune) (runes []rune, out string, err error) string {
// 	for len(runes) > 0 {
// 		next := runes[0]
// 		if nxt
// 	}
// }

// func ParseIndex(index string) ([]interface{}, error) {
// 	stream := []rune(index)

// 	for len(stream) > 0 {
// 		char := stream[0]
// 		stream = stream[1:]
// 		switch char {

// 		}
// 	}
// }
