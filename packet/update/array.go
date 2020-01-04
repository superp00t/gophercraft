package update

// PlayerQuestLog: ArrayValue{
//	{"Creator", }
// }

type ArrayCell struct {
	Key   *string
	Value interface{}
}

type ArrayData struct {
	Cols []string
	Rows [][]interface{}
}

type Array struct {
	c           *Class
	BlockOffset int64
	Len         int64
	Fields      []ArrayField
}

type ArrayField struct {
	Key string
	Len int64
	FieldType
	FieldFlags
}

func (ad *ArrayData) SetValue(column string, row int, value interface{}) {
	for idx, v := range ad.Cols {
		if column == v {
			if len(ad.Rows) <= row {
				diff := make([][]interface{}, (row-len(ad.Rows))+1)
				ad.Rows = append(ad.Rows, diff...)
			}

			ad.Rows[row][idx] = value
			return
		}
	}

	panic("unknown column " + column)
}

func (c *Class) Array(glob Global, ln int64) *Array {
	arr := &Array{
		c:   c,
		Len: ln,
	}

	ft := &ClassField{
		Global:    glob,
		FieldType: ArrayType,
		array:     arr,
	}

	c.addOffsets(ft)
	c.Fields = append(c.Fields, ft)
	return arr
}

func (arr *Array) Uint32Array(key string, ln int64, view FieldFlags) {
	arr.Fields = append(arr.Fields, ArrayField{
		key,
		ln,
		Uint32Array,
		view,
	})

	arr.BlockOffset += ln
}

func (arr *Array) Uint32(key string, view FieldFlags) {
	arr.Fields = append(arr.Fields, ArrayField{
		key,
		0,
		Uint32,
		view,
	})

	arr.BlockOffset++
}

func (arr *Array) Pad() {
	arr.Fields = append(arr.Fields, ArrayField{
		"",
		0,
		Pad,
		Private,
	})

	arr.BlockOffset++
}

func (arr *Array) GUID(key string, view FieldFlags) {
	arr.Fields = append(arr.Fields, ArrayField{
		key,
		0,
		GUID,
		view,
	})

	arr.BlockOffset += 2
}

func (arr *Array) End() {
	arr.c.BlockOffset += (arr.BlockOffset * arr.Len)
	arr.c.BitOffset = 0
}
