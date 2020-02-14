package csv

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strings"

	"github.com/superp00t/etc"
)

type Writer struct {
	closer      io.WriteCloser
	columnNames []string
	writer      *csv.Writer
	init        bool
}

func NewWriter(wrc io.WriteCloser) *Writer {
	wr := new(Writer)
	wr.writer = csv.NewWriter(wrc)
	wr.closer = wrc
	return wr
}

func escapeString(in string) string {
	b, err := json.Marshal(in)
	if err != nil {
		panic(err)
	}
	// // Trim the beginning and trailing " character
	// return string(b[1 : len(b)-1])
	return string(b)
}

func encodeStruct(cell reflect.Value) string {
	str := etc.NewBuffer()
	nmField := cell.NumField()
	for i := 0; i < nmField; i++ {
		comma := false
		if i != nmField-1 {
			comma = true
		}
		fieldName := cell.Type().Field(i).Name
		str.Write([]byte(fieldName))
		str.WriteRune(':')
		str.Write([]byte(encodeCell(cell.Field(i))))
		if comma {
			str.WriteRune(';')
		}
	}

	return str.ToString()
}

func encodeCell(cell reflect.Value) string {
	switch cell.Kind() {
	case reflect.Uint8:
		return fmt.Sprintf("%d", cell.Interface())
	case reflect.Uint16:
		return fmt.Sprintf("%d", cell.Interface())
	case reflect.Uint32:
		return fmt.Sprintf("%d", cell.Interface())
	case reflect.Uint64:
		return fmt.Sprintf("%d", cell.Interface())
	case reflect.Int8:
		return fmt.Sprintf("%d", cell.Interface())
	case reflect.Int16:
		return fmt.Sprintf("%d", cell.Interface())
	case reflect.Int32:
		return fmt.Sprintf("%d", cell.Interface())
	case reflect.Int64:
		return fmt.Sprintf("%d", cell.Interface())
	case reflect.Float32, reflect.Float64:
		str, _ := json.Marshal(cell.Interface())
		return string(str)
	case reflect.String:
		return cell.String()
	case reflect.Bool:
		if cell.Bool() {
			return "true"
		}
		return "false"
	case reflect.Ptr:
		return encodeCell(cell.Elem())
	case reflect.Slice:
		strSlice := make([]string, cell.Len())
		for i := range strSlice {
			strSlice[i] = encodeCell(cell.Index(i))
		}
		return strings.Join(strSlice, ",")
	case reflect.Struct:
		return encodeStruct(cell)
	default:
		panic(cell.Kind())
	}
}

func encodeRecord(valueType reflect.Type, enabledFields []string, value reflect.Value) []string {
	field := make([]string, len(enabledFields))

	for idx, name := range enabledFields {
		fld := value.FieldByName(name)
		field[idx] = encodeCell(fld)
	}

	return field
}

func (wr *Writer) AddRecord(v interface{}) error {
	value := reflect.ValueOf(v)

	if value.Kind() == reflect.Ptr {
		value = value.Elem()
	}

	valueType := value.Type()

	if !wr.init {
		wr.columnNames = []string{}

		for c := 0; c < valueType.NumField(); c++ {
			field := valueType.Field(c)
			csv := field.Tag.Get("csv")
			if csv != "-" {
				wr.columnNames = append(wr.columnNames, field.Name)
			}
		}

		if err := wr.writer.Write(wr.columnNames); err != nil {
			return err
		}

		wr.init = true
	}

	return wr.writer.Write(encodeRecord(valueType, wr.columnNames, value))
}

func (wr *Writer) Close() error {
	wr.writer.Flush()
	return wr.closer.Close()
}
