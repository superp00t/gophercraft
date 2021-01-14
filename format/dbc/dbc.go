//Package dbc implements a reflection-based decoder for multiple versions of the DBC/DB2 format
package dbc

import (
	"fmt"
	"reflect"

	"github.com/superp00t/gophercraft/vsn"

	"github.com/superp00t/etc"
)

//go:generate gcraft_stringer -type=FieldType
type FieldType int
type MagicType int

var magicMap = map[string]MagicType{
	"WDB2": WDB2,
	"WDB3": WDB2,
	"WDB4": WDB2,
	"WDB5": WDB5,
	"WDBC": WDBC,
}

const (
	// Format versions
	WDBC MagicType = 1
	WDB2 MagicType = 2
	WDB3 MagicType = 3
	WDB4 MagicType = 4
	WDB5 MagicType = 5
	WDB6 MagicType = 6
	WDC1 MagicType = 7
	WDC2 MagicType = 8

	HasOffsetMap    uint16 = 0x01
	HasSecondaryKey uint16 = 0x02
	HasNonOnlineIds uint16 = 0x04
	IsBitpacked     uint16 = 0x10
)

const (
	Uint8 FieldType = iota
	Uint16
	Uint32
	Uint64
	Int32
	Float
	String
	Array
	Slice
)

type FieldStruct struct {
	Size     int16
	Position uint16
}

// DBC contains DBC/DB2 file metadata.
type DBC struct {
	Magic           MagicType
	RecordCount     uint32
	FieldCount      uint32
	RecordSize      uint32
	StringBlockSize uint32

	// DB2 only
	TableHash            uint32
	Build                vsn.Build
	TimestampLastWritten uint32
	MinID, MaxID         uint32
	Locale               uint32
	CopyTableSize        uint32

	// DB5
	LayoutHash     uint32
	Flags          uint16
	IDIndex        uint16
	FieldStructure []FieldStruct
	stringRefs     map[uint32]*string

	buf *etc.Buffer
}

type Record struct {
	ID        uint32
	Reference uint32
}

func Parse(game vsn.Build, input []byte) (*DBC, error) {
	return Decode(game, etc.FromBytes(input))
}

func Open(game vsn.Build, path string) (*DBC, error) {
	f, err := etc.FileController(path)
	if err != nil {
		return nil, err
	}

	return Decode(game, f)
}

func Decode(game vsn.Build, i *etc.Buffer) (*DBC, error) {
	d := new(DBC)

	d.buf = i

	d.Magic = magicMap[i.ReadFixedString(4)]
	if d.Magic == 0 {
		return nil, fmt.Errorf("dbc: not a DBC file")
	}

	switch d.Magic {
	case WDBC:
		d.RecordCount = i.ReadUint32()
		d.FieldCount = i.ReadUint32()
		d.RecordSize = i.ReadUint32()
		d.StringBlockSize = i.ReadUint32()
		d.Build = game
	case WDB2:
		d.RecordCount = i.ReadUint32()
		d.FieldCount = i.ReadUint32()
		d.RecordSize = i.ReadUint32()
		d.StringBlockSize = i.ReadUint32()
		d.TableHash = i.ReadUint32()
		d.Build = vsn.Build(i.ReadUint32())
		d.TimestampLastWritten = i.ReadUint32()
		d.MinID = i.ReadUint32()
		d.MaxID = i.ReadUint32()
		d.Locale = i.ReadUint32()
		d.CopyTableSize = i.ReadUint32()
	case WDB3, WDB4:
		return nil, fmt.Errorf("dbc: DB3 and DB4 are deprecated")
	case WDB5:
		d.RecordCount = i.ReadUint32()
		d.FieldCount = i.ReadUint32()
		d.RecordSize = i.ReadUint32()
		d.StringBlockSize = i.ReadUint32()
		d.TableHash = i.ReadUint32()
		d.LayoutHash = i.ReadUint32()
		d.MinID = i.ReadUint32()
		d.MaxID = i.ReadUint32()
		d.Locale = i.ReadUint32()
		d.CopyTableSize = i.ReadUint32()
		d.Flags = i.ReadUint16()
		d.IDIndex = i.ReadUint16()

		if d.Flags&HasNonOnlineIds != 0 {
			d.IDIndex = 0
		}

		d.FieldStructure = make([]FieldStruct, d.FieldCount)
		for z := 0; z < int(d.FieldCount); z++ {
			fs := FieldStruct{}
			fs.Size = i.ReadInt16()
			fs.Position = i.ReadUint16()
			d.FieldStructure[z] = fs
		}
	case WDB6, WDC1, WDC2:
		return nil, fmt.Errorf("dbc: not supported yet")
	default:
		return nil, fmt.Errorf("dbc: not a DBC file")
	}

	d.stringRefs = make(map[uint32]*string)

	return d, nil
}

type _fieldType struct {
	Type      FieldType
	ArrayType *_fieldType
	Length    int
	tag       string
	opts      []tagOpt
	disabled  bool
}

func mt(t FieldType) *_fieldType {
	return &_fieldType{
		Type: t,
	}
}

func (d *DBC) getType(v reflect.Type) (*_fieldType, error) {
	rt := mt(0)
	switch v.Kind() {
	case reflect.String:
		rt = mt(String)
	case reflect.Float32:
		rt = mt(Float)
	case reflect.Int32:
		rt = mt(Int32)
	case reflect.Uint8:
		rt = mt(Uint8)
	case reflect.Uint16:
		rt = mt(Uint16)
	case reflect.Uint32:
		rt = mt(Uint32)
	case reflect.Uint64:
		rt = mt(Uint64)
	case reflect.Slice:
		rt = mt(Slice)
		var err error
		rt.ArrayType, err = d.getType(v.Elem())
		if err != nil {
			return nil, err
		}
	case reflect.Array:
		t := mt(Array)
		t.Length = v.Len()
		tp := v.Elem()
		var err error
		t.ArrayType, err = d.getType(tp)
		if err != nil {
			return nil, err
		}
		rt = t
	default:
		return nil, fmt.Errorf("Unknown type")
	}

	return rt, nil
}

func (d *DBC) ParseRecords(out interface{}) error {
	// val := reflect.Indirect(reflect.ValueOf(enttype))
	structType := reflect.TypeOf(out).Elem().Elem()
	dummyValue := reflect.Indirect(reflect.New(structType))
	structLen := structType.NumField()

	recTypes := make([]*_fieldType, structLen)

	for i := 0; i < structLen; i++ {
		tp, err := d.getType(dummyValue.Field(i).Type())
		if err != nil {
			return err
		}
		sfield := structType.Field(i)
		str, ok := sfield.Tag.Lookup("dbc")
		tp.tag = str
		if ok {
			tg := parseTag(str)
			tp.disabled, tp.opts = tg.getValidOpts(d.Build)
			if tp.disabled {
				fmt.Println(structType.Field(i).Name, "is disabled in", d.Build)
			}

			if tp.Type == Slice && tp.Length == 0 {
				for _, v := range tp.opts {
					if v.Type == lengthOpt {
						tp.Length = int(v.Len)
					}
				}
			}
		} else {
			if tp.Type == Slice {
				return fmt.Errorf("dbc: %s: supply (len:X) parameter to field tag if you want to use Go slices.", dummyValue.Type().Field(i).Name)
			}
		}
		recTypes[i] = tp
	}

	il := int(d.RecordCount)

	sli := reflect.ValueOf(out).Elem()
	sli.Set(reflect.MakeSlice(sli.Type(), il, il))

	if reflect.ValueOf(out).Elem().Len() < il {
		return fmt.Errorf("You must make a slice of size DBC.Record to hold the output")
	}

	for i := uint32(0); i < d.RecordCount; i++ {
		rcd := reflect.ValueOf(out).Elem().Index(int(i))

		for ri := 0; ri < len(recTypes); ri++ {
			rt := recTypes[ri]
			if rt.disabled {
				continue
			}

			fld := rcd.Field(ri)
			if !fld.CanSet() {
				panic("cannot field")
			}

			d.setField(fld, d.buf, rt)
		}
	}

	if d.Flags&HasOffsetMap == 0 {
		stringBlock := etc.MkBuffer(d.buf.ReadBytes(int(d.StringBlockSize)))
		for k, ptr := range d.stringRefs {
			stringBlock.SeekR(int64(k))
			str := stringBlock.ReadCString()
			if ptr != nil {
				*ptr = str
			}
		}
	}

	return nil
}

func (ft *_fieldType) isLoc() bool {
	for _, v := range ft.opts {
		if v.Type == locOpt {
			return true
		}
	}

	return false
}

func (d *DBC) setField(fld reflect.Value, buf *etc.Buffer, tp *_fieldType) {
	switch tp.Type {
	case String:
		if tp.isLoc() {
			ln := 0
			if d.Build.RemovedIn(13164) {
				ln = 9
			}

			if d.Build.AddedIn(6692) {
				ln = 17
			}

			lc := make([]uint32, ln)
			for x := 0; x < ln; x++ {
				lc[x] = buf.ReadUint32()
			}
			pptr := fld.Addr().Interface().(*string)
			d.stringRefs[lc[0]] = pptr
			return
		}
		if d.Flags&HasOffsetMap != 0 {
			panic("cannot read without offsetmap")
		} else {
			id := buf.ReadUint32()
			pptr := fld.Addr().Interface().(*string)
			d.stringRefs[id] = pptr
		}
	case Int32:
		fld.SetInt(int64(buf.ReadInt32()))
	case Float:
		fld.SetFloat(float64(buf.ReadFloat32()))
	case Uint8:
		fld.SetUint(uint64(buf.ReadByte()))
	case Uint16:
		fld.SetUint(uint64(buf.ReadUint16()))
	case Uint32:
		fld.SetUint(uint64(buf.ReadUint32()))
	case Uint64:
		fld.SetUint(uint64(buf.ReadUint64()))
	case Array:
		for ai := 0; ai < tp.Length; ai++ {
			d.setField(fld.Index(ai), buf, tp.ArrayType)
		}
	case Slice:
		slice := reflect.MakeSlice(fld.Type(), tp.Length, tp.Length)
		for ai := 0; ai < tp.Length; ai++ {
			d.setField(slice.Index(ai), buf, tp.ArrayType)
		}
		any := false
		for x := 0; x < slice.Len(); x++ {
			val := slice.Index(x)
			if !val.IsZero() {
				any = true
			}
		}
		if any {
			fld.Set(slice)
		}
	}
}
