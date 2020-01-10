package csv

import (
	"encoding/csv"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"
)

type Scanner struct {
	closer      io.ReadCloser
	reader      *csv.Reader
	columnNames []string
	init        bool
}

func NewScanner(in io.ReadCloser) (*Scanner, error) {
	rdr := &Scanner{}
	rdr.closer = in
	rdr.reader = csv.NewReader(in)
	rdr.reader.Comment = '#'

	return rdr, nil
}

func (rd *Scanner) Scan(v interface{}) error {
	value := reflect.ValueOf(v)

	if value.Kind() == reflect.Ptr {
		value = value.Elem()
	}

	valueType := value.Type()

	if !rd.init {
		var err error

		// we still haven't read the header record yet.
		// NOTE: As Gophercraft continues in its development, new fields will be added.
		// it should be acceptable to omit certain fields from your CSV, however not to include unknown fields.
		rd.columnNames, err = rd.reader.Read()
		if err != nil {
			return err
		}

		for _, ct := range rd.columnNames {
			_, found := valueType.FieldByName(ct)
			if !found {
				return fmt.Errorf("could not find field %s in field %s", ct, valueType.String())
			}
		}

		rd.init = true
	}

	switch value.Kind() {
	case reflect.Struct:
	default:
		return fmt.Errorf("csv: cannot scan to non-struct type")
	}

	rec, err := rd.reader.Read()
	if err != nil {
		return err
	}

	if len(rec) != len(rd.columnNames) {
		return fmt.Errorf("record length mismatch in line: %s", strings.Join(rec, ","))
	}

	for i, cname := range rd.columnNames {
		recd := rec[i]

		field := value.FieldByName(cname)

		err := parseField(field, recd)

		if err != nil {
			return err
		}
	}

	return nil
}

func (rd *Scanner) Close() error {
	return rd.closer.Close()
}

func parseUint(rec reflect.Value, bitSize int, value string) error {
	i, err := strconv.ParseUint(value, 0, bitSize)
	if err != nil {
		return err
	}

	rec.SetUint(i)
	return nil
}

func parseInt(rec reflect.Value, bitSize int, value string) error {
	i, err := strconv.ParseInt(value, 0, bitSize)
	if err != nil {
		return err
	}

	rec.SetInt(i)
	return nil
}

func parseFloat(rec reflect.Value, bitSize int, value string) error {
	f, err := strconv.ParseFloat(value, bitSize)
	if err != nil {
		return err
	}

	rec.SetFloat(f)
	return nil
}

func parseStruct(strct reflect.Value, value string) error {
	split := strings.Split(value, ";")

	for _, spl := range split {
		keys := strings.Split(spl, ":")
		if len(keys) != 2 {
			return fmt.Errorf("invalid number of elements")
		}

		_, ok := strct.Type().FieldByName(keys[0])
		if !ok {
			return fmt.Errorf("no field by name %d", value)
		}

		if err := parseField(strct.FieldByName(keys[0]), keys[1]); err != nil {
			return err
		}
	}

	return nil
}

func parseField(field reflect.Value, recd string) error {
	var err error
	switch field.Kind() {
	case reflect.Uint8:
		err = parseUint(field, 8, recd)
	case reflect.Uint16:
		err = parseUint(field, 16, recd)
	case reflect.Uint32:
		err = parseUint(field, 32, recd)
	case reflect.Uint64:
		err = parseUint(field, 64, recd)
	case reflect.Int8:
		err = parseInt(field, 8, recd)
	case reflect.Int16:
		err = parseInt(field, 16, recd)
	case reflect.Int32:
		err = parseInt(field, 32, recd)
	case reflect.Int64:
		err = parseInt(field, 64, recd)
	case reflect.String:
		field.SetString(recd)
	case reflect.Float32:
		err = parseFloat(field, 32, recd)
	case reflect.Float64:
		err = parseFloat(field, 64, recd)
	case reflect.Bool:
		field.SetBool(recd == "true")
	case reflect.Slice:
		strs := strings.Split(recd, ",")

		sli := reflect.MakeSlice(field.Type(), len(strs), len(strs))

		for i, v := range strs {
			err = parseField(sli.Index(i), v)

			if err != nil {
				break
			}
		}

		if err == nil {
			field.Set(sli)
		}
	case reflect.Struct:
		err = parseStruct(field, recd)
	default:
		err = fmt.Errorf("unhandled field kind %s", field.Kind().String())
	}

	return err
}
