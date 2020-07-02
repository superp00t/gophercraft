package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/superp00t/etc"

	"github.com/superp00t/etc/yo"
	"github.com/superp00t/gophercraft/packet/update"
)

var (
	file *os.File
)

func arrayType(field *update.ClassField) string {
	out := etc.NewBuffer()

	arr := field.Array()
	fmt.Fprintf(out, "[%d]struct {\n", arr.Len)
	for x := 0; x < int(len(arr.Fields)); x++ {
		fmt.Fprintf(out, "  %s", arr.Fields[x].Key)
		cf := &update.ClassField{
			FieldType: arr.Fields[x].FieldType,
			SliceSize: arr.Fields[x].Len,
			Flags:     arr.Fields[x].FieldFlags,
		}
		fmt.Fprintf(out, " %s%s\n", fieldType(cf), fieldTag(cf))
	}
	fmt.Fprintf(out, "}\n")
	return out.ToString()
}

func fieldType(field *update.ClassField) string {
	switch field.FieldType {
	case update.Uint32:
		return "uint32"
	case update.Float32:
		return "float32"
	case update.Uint8:
		return "uint8"
	case update.Uint32Array:
		return fmt.Sprintf("[%d]uint32", field.SliceSize)
	case update.Float32Array:
		return fmt.Sprintf("[%d]float32", field.SliceSize)
	case update.Bit:
		return fmt.Sprintf("bool")
	case update.GUID:
		return "guid.GUID"
	case update.GUIDArray:
		return fmt.Sprintf("[%d]guid.GUID", field.SliceSize)
	case update.Int32:
		return "int32"
	case update.Int32Array:
		return fmt.Sprintf("[%d]int32", field.SliceSize)
	case update.ArrayType:
		return arrayType(field)
	case update.Pad:
		return "uint32"
	}
	return "error"
}

func fieldTag(field *update.ClassField) string {
	tag := []string{}

	if field.Flags&update.Public != 0 {
		tag = append(tag, "public")
	}

	if field.Flags&update.Private != 0 {
		tag = append(tag, "private")
	}

	if field.Flags&update.Party != 0 {
		tag = append(tag, "party")
	}

	if len(tag) > 0 {
		return " `update:\"" + strings.Join(tag, ",") + "\"`"
	}

	return ""
}
func generateDescriptor(build uint32) {
	desc := update.Descriptors[build]

	for _, class := range desc.Classes {
		fmt.Fprintf(file, "type %sData struct {\n", class.Name)
		for _, field := range class.Fields {
			fmt.Fprintf(file, "\t%s %s%s\n", field.Global, fieldType(field), fieldTag(field))
		}
		fmt.Fprintf(file, "}\n\n")
	}

	for _, class := range desc.Classes {
		var extends *update.Class = class.Extends
		if extends == nil {
			continue
		}
		var extendList = []string{class.Name}

		// Player
		fmt.Fprintf(file, "type %sDescriptor struct {\n", class.Name)

		fmt.Println("Ok ", class.Name)
		for extends != nil {
			// {} -> {Unit}
			yo.Spew(extendList)
			fmt.Println("prepending", extends.Name)
			// {Unit} -> {Object, Unit}
			extendList = append([]string{extends.Name}, extendList...)
			// Extends = Unit, Object
			extends = extends.Extends
			// ExtendList = {Unit}, {Object, Unit}
		}
		fmt.Println("Ok done")
		yo.Spew(extendList)

		for _, exClass := range extendList {
			fmt.Fprintf(file, "\t%sData\n", exClass)
		}

		if len(extendList) != 0 {
			fmt.Fprintf(file, "}\n")
		}
	}
}

func main() {
	var err error

	const path = "descriptors.go"

	os.Remove(path)

	file, err = os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0700)
	if err != nil {
		panic(err)
	}

	fmt.Fprintln(file, "package update")

	generateDescriptor(5875)

	c := exec.Command("gofmt", "-w", path)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	c.Run()

	file.Close()

	b, _ := ioutil.ReadFile(path)
	fmt.Println(string(b))
}
