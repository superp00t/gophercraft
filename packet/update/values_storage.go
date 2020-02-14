package update

import (
	"fmt"

	"github.com/superp00t/etc/yo"
	"github.com/superp00t/gophercraft/guid"
)

func NewValuesBlock() *ValuesBlock {
	return &ValuesBlock{
		Values:  make(map[Global]interface{}),
		Changes: make(map[Global]bool),
	}
}

func (vb *ValuesBlock) ModifyAndLock(attributes map[Global]interface{}) {
	vb.Lock()
	for glob, attr := range attributes {
		vb.Values[glob] = attr
		vb.Changes[glob] = true
	}
}

func (vb *ValuesBlock) ClearChanges() {
	for k := range vb.Changes {
		delete(vb.Changes, k)
	}
}

func (vb *ValuesBlock) ClearChangesAndUnlock() {
	vb.ClearChanges()
	vb.Unlock()
}

func (vb *ValuesBlock) Set(glob Global, value interface{}) {
	vb.Lock()
	vb.Values[glob] = value
	vb.Changes[glob] = true
	vb.Unlock()
}

func InitArrayData(version uint32, glob Global) *ArrayData {
	var cols []string
	for _, v := range Descriptors[version].Classes {
		for _, cf := range v.Fields {
			if cf.Global == glob {
				for _, v := range cf.array.Fields {
					if v.FieldType != Pad {
						cols = append(cols, v.Key)
					}
				}
				goto end
			}
		}
	}

end:

	return &ArrayData{
		cols,
		nil,
	}
}

func (vb *ValuesBlock) SetArrayValue(glob Global, index int, columnName string, value interface{}) {
	vb.Lock()
	arrayField := vb.Values[glob].(*ArrayData)
	if arrayField == nil {
		panic("SetArrayValue on nil")
	}

	arrayField.SetValue(columnName, index, value)
	vb.Changes[glob] = true
	vb.Unlock()
}

// Ensure proper type inference by setting with these functions

func (vb *ValuesBlock) SetTypeMask(version uint32, tm guid.TypeMask) {
	data, err := tm.Resolve(version)
	if err != nil {
		panic(err)
	}
	vb.SetUint32Value(ObjectType, data)
}

func (vb *ValuesBlock) SetUint32Value(glob Global, value uint32) {
	vb.Set(glob, value)
}

func (vb *ValuesBlock) SetGUIDValue(glob Global, value guid.GUID) {
	vb.Set(glob, value)
}

func (vb *ValuesBlock) SetFloat32Value(glob Global, value float32) {
	vb.Set(glob, value)
}

func (vb *ValuesBlock) SetUint32Array(glob Global, values ...interface{}) {
	ptr := make([]*uint32, len(values))

	for idx, v := range values {
		if v != nil {
			data := uint32(0)
			switch vt := v.(type) {
			case uint32:
				data = vt
			case int:
				data = uint32(vt)
			}

			ptr[idx] = &data
		}
	}

	vb.Values[glob] = ptr
}

func (vb *ValuesBlock) SetByteValue(glob Global, value uint8) {
	vb.Set(glob, value)
}

func (vb *ValuesBlock) SetInt32Value(glob Global, value int32) {
	vb.Set(glob, value)
}

func (vb *ValuesBlock) Get(glob Global) (interface{}, error) {
	dat, ok := vb.Values[glob]
	if !ok {
		return nil, fmt.Errorf("update: Global %s has not been entered", glob)
	}

	return dat, nil
}

func (vb *ValuesBlock) GetByteValue(glob Global) uint8 {
	dat, err := vb.Get(glob)
	if err != nil {
		return 0
	}

	return dat.(uint8)
}

func (vb *ValuesBlock) GetUint32Value(glob Global) uint32 {
	dat, err := vb.Get(glob)
	if err != nil {
		return 0
	}

	return dat.(uint32)
}

func (vb *ValuesBlock) GetGUIDValue(glob Global) guid.GUID {
	dat, err := vb.Get(glob)
	if err != nil {
		return guid.Nil
	}

	return dat.(guid.GUID)
}

func (vb *ValuesBlock) SetGUIDArrayValue(glob Global, index int, value guid.GUID) {
	yo.Warn("Setting...", glob, index, value)
	vb.Lock()
	sli := vb.Values[glob].([]*guid.GUID)
	sli[index] = &value
	vb.Changes[glob] = true
	vb.Unlock()
}

func (vb *ValuesBlock) SetFloat32ArrayValue(glob Global, index int, value float32) {
	vb.Lock()
	sli := vb.Values[glob].([]*float32)
	sli[index] = &value
	vb.Changes[glob] = true
	vb.Unlock()
}

func (vb *ValuesBlock) SetInt32ArrayValue(glob Global, index int, value int32) {
	vb.Lock()
	sli := vb.Values[glob].([]*int32)
	sli[index] = &value
	vb.Changes[glob] = true
	vb.Unlock()
}

func (vb *ValuesBlock) SetUint32ArrayValue(glob Global, index int, value uint32) {
	vb.Lock()
	sli := vb.Values[glob].([]*uint32)
	sli[index] = &value
	vb.Changes[glob] = true
	vb.Unlock()
}

func (vb *ValuesBlock) GetGUIDArrayValue(glob Global, index int) guid.GUID {
	dat, err := vb.Get(glob)
	if err != nil {
		panic(err)
	}

	if dat == nil {
		panic(glob.String() + " is nil")
	}

	sli := dat.([]*guid.GUID)
	if sli[index] == nil {
		return guid.Nil
	}

	return *sli[index]
}
