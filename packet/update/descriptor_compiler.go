package update

import (
	"fmt"
)

//go:generate gcraft_stringer -type=FieldType
type FieldType uint32

const (
	Uint32 FieldType = iota
	Uint32Array
	Int32
	Int32Array
	Float32
	Float32Array
	GUID
	GUIDArray
	ArrayType
	Uint8
	Bit
	Pad

	Public FieldFlags = 0

	Private FieldFlags = 1 << iota
	Party
)

type ClassField struct {
	Global
	FieldType

	class       *Class
	array       *Array
	SliceSize   int64
	BitOffset   uint8
	BlockOffset int64
	Flags       FieldFlags
}

type Class struct {
	dc          *DescriptorCompiler
	Name        string
	Extends     *Class
	BaseOffset  int64
	BlockOffset int64 // block refers to 32 bits
	BitOffset   uint8 // multiple values can be stored inside a 32-bit block, this will increment by 8 in the case of "bytes" fields
	Fields      []*ClassField
}

type DescriptorCompiler struct {
	Version uint32
	Classes []*Class
}

func NewDescriptorCompiler(vsn uint32) *DescriptorCompiler {
	return &DescriptorCompiler{
		Version: vsn,
	}
}

func (dc *DescriptorCompiler) ObjectBase() *Class {
	c := &Class{
		dc:   dc,
		Name: "Object",
	}
	dc.Classes = append(dc.Classes, c)
	return c
}

func (c *Class) Extend(name string) *Class {
	c2 := &Class{
		dc:         c.dc,
		Name:       name,
		BaseOffset: c.BaseOffset + c.BlockOffset,
		Extends:    c,
	}

	c.dc.Classes = append(c.dc.Classes, c2)
	return c2
}

// types

func (c *Class) Uint32(glob Global, view FieldFlags) {
	c.bitReset()
	f := &ClassField{
		Global:    glob,
		FieldType: Uint32,
		Flags:     view,
	}
	c.addOffsets(f)
	c.advanceBy(1)
	c.Fields = append(c.Fields, f)
}

func (c *Class) Int32(glob Global, view FieldFlags) {
	c.bitReset()
	f := &ClassField{
		Global:    glob,
		FieldType: Int32,
		Flags:     view,
	}
	c.addOffsets(f)
	c.advanceBy(1)
	c.Fields = append(c.Fields, f)
}

func (c *Class) Float32(glob Global, view FieldFlags) {
	c.bitReset()
	f := &ClassField{
		Global:    glob,
		FieldType: Float32,
		Flags:     view,
	}
	c.addOffsets(f)
	c.advanceBy(1)
	c.Fields = append(c.Fields, f)
}

func (c *Class) GUID(glob Global, view FieldFlags) {
	c.bitReset()
	f := &ClassField{
		Global:    glob,
		FieldType: GUID,
		Flags:     view,
	}
	c.addOffsets(f)
	c.advanceBy(2)
	c.Fields = append(c.Fields, f)
}

func (c *Class) Bit(glob Global, view FieldFlags) {
	f := &ClassField{
		Global:    glob,
		FieldType: Bit,
		Flags:     view,
	}
	c.addOffsets(f)
	c.Fields = append(c.Fields, f)
	c.BitOffset++
	if c.BitOffset == 32 {
		c.BlockOffset++
		c.BitOffset = 0
	}
}

func (c *Class) GUIDArray(glob Global, size int64, view FieldFlags) {
	c.bitReset()
	f := &ClassField{
		Global:    glob,
		FieldType: GUIDArray,
		SliceSize: size,
		Flags:     view,
	}
	c.addOffsets(f)
	c.advanceBy(size * 2)
	c.Fields = append(c.Fields, f)
}

func (c *Class) advanceBy(i int64) {
	c.BlockOffset += i
	c.BitOffset = 0
}

func (c *Class) bitReset() {
	if c.BitOffset > 0 {
		c.BlockOffset++
		c.BitOffset = 0
	}
}

func (c *Class) Uint32Array(glob Global, size int64, view FieldFlags) {
	c.bitReset()
	f := &ClassField{
		Global:    glob,
		FieldType: Uint32Array,
		SliceSize: size,
		Flags:     view,
	}
	c.addOffsets(f)
	c.advanceBy(size)
	c.Fields = append(c.Fields, f)
}

func (c *Class) Int32Array(glob Global, size int64, view FieldFlags) {
	c.bitReset()
	f := &ClassField{
		Global:    glob,
		FieldType: Int32Array,
		SliceSize: size,
		Flags:     view,
	}
	c.addOffsets(f)
	c.advanceBy(size)
	c.Fields = append(c.Fields, f)
}

func (c *Class) Float32Array(glob Global, size int64, view FieldFlags) {
	c.bitReset()
	f := &ClassField{
		Global:    glob,
		FieldType: Float32Array,
		SliceSize: size,
		Flags:     view,
	}
	c.addOffsets(f)
	c.advanceBy(size)
	c.Fields = append(c.Fields, f)
}

func (c *Class) addOffsets(ptr *ClassField) {
	ptr.BitOffset = c.BitOffset
	ptr.BlockOffset = c.BlockOffset
	ptr.class = c
}

func (c *Class) Byte(glob Global, view FieldFlags) {
	f := &ClassField{
		Global:    glob,
		FieldType: Uint8,
		Flags:     view,
	}

	c.addOffsets(f)

	c.BitOffset += 8

	if c.BitOffset == 32 {
		c.BitOffset = 0
		c.BlockOffset++
	}

	c.Fields = append(c.Fields, f)
}

func (c *Class) ExtractAllClassFields() []*ClassField {
	topFields := c.Fields
	fieldSlice := topFields

	curClass := c

	for curClass.Extends != nil {
		curClass = curClass.Extends
		fieldSlice = append(curClass.Fields, fieldSlice...)
	}

	return fieldSlice
}

func queryFieldByAbsoluteOffset(cf []*ClassField, u uint32) (*ClassField, error) {
	for _, v := range cf {
		if int64(u) == v.class.BaseOffset+v.BlockOffset {
			return v, nil
		}
	}

	return nil, fmt.Errorf("update: could not find field for offset 0x%08X", u)
}

func (c *DescriptorCompiler) GetClass(className string) (*Class, error) {
	for _, v := range c.Classes {
		if v.Name == className {
			return v, nil
		}
	}

	return nil, fmt.Errorf("update: could not find class named %s", className)
}

func (c *Class) EndOffset() uint32 {
	return uint32(c.BaseOffset + c.BlockOffset)
}

func (c *Class) Pad() {
	var pad ClassField
	c.addOffsets(&pad)
	pad.FieldType = Pad
	c.BlockOffset++
	c.Fields = append(c.Fields, &pad)
	c.BitOffset = 0
}

func (c *ClassField) CppString() string {
	return fmt.Sprintf("%s %s", c.FieldType.String(), c.Global.String())
}

func (c *ClassField) AbsBlockOffset() int64 {
	return c.class.BaseOffset + c.BlockOffset
}
