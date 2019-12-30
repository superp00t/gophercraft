package csv

import (
	"encoding/csv"
	"fmt"
	"io"
	"reflect"
	"strings"
)

type Writer struct {
	closer io.WriteCloser
	writer *csv.Writer
	init   bool
}

func NewWriter(wrc io.WriteCloser) *Writer {
	wr := new(Writer)
	wr.writer = csv.NewWriter(wrc)
	wr.closer = wrc
	return wr
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
	case reflect.Float32:
		return fmt.Sprintf("%f", cell.Interface())
	case reflect.Float64:
		return fmt.Sprintf("%f", cell.Interface())
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
	default:
		panic(cell.Kind())
	}
}

func encodeRecord(valueType reflect.Type, value reflect.Value) []string {
	field := make([]string, valueType.NumField())

	for c := 0; c < valueType.NumField(); c++ {
		field[c] = encodeCell(value.Field(c))
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
		columnNames := []string{}

		for c := 0; c < valueType.NumField(); c++ {
			columnNames = append(columnNames, valueType.Field(c).Name)
		}

		if err := wr.writer.Write(columnNames); err != nil {
			return err
		}

		wr.init = true
	}

	return wr.writer.Write(encodeRecord(valueType, value))
}

func (wr *Writer) Close() error {
	wr.writer.Flush()
	return wr.closer.Close()
}
