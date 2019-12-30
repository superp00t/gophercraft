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
	FieldType
	FieldFlags
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

func (arr *Array) Uint32(key string, view FieldFlags) {
	arr.Fields = append(arr.Fields, ArrayField{
		key,
		Uint32,
		view,
	})

	arr.BlockOffset++
}

func (arr *Array) Pad() {
	arr.Fields = append(arr.Fields, ArrayField{
		"",
		Pad,
		Private,
	})

	arr.BlockOffset++
}

func (arr *Array) GUID(key string, view FieldFlags) {
	arr.Fields = append(arr.Fields, ArrayField{
		key,
		GUID,
		view,
	})

	arr.BlockOffset += 2
}

func (arr *Array) End() {
	arr.c.BlockOffset += (arr.BlockOffset * arr.Len)
	arr.c.BitOffset = 0
}
