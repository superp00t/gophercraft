//Package ccfg handles CASC and NGDP config text files
//this package shouldn't be used for general text encoding
package ccfg

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"
)

type Comment struct{}

var _comment = reflect.TypeOf(Comment{})

type Decoder struct {
	*bufio.Reader
}

func NewDecoder(rd io.Reader) *Decoder {
	return &Decoder{
		Reader: bufio.NewReader(rd),
	}
}

func (t *Decoder) decodeWords() ([]string, error) {
	line, err := t.ReadString('\n')
	if err == io.EOF {
		if len(line) == 0 {
			return nil, err
		}
	} else {
		if err != nil {
			return nil, err
		}
	}

	line = strings.TrimLeft(line, " ")

	line = strings.TrimRight(line, "\r\n")
	return strings.Split(line, " "), nil
}

func (t *Decoder) decodeSingleValue(word string, value reflect.Value) error {
	if word == "" {
		return nil
	}

	if isHexData(value) {
		data, err := hex.DecodeString(word)
		if err != nil {
			return err
		}
		if value.Kind() == reflect.Array {
			if len(data) != value.Len() {
				return fmt.Errorf("data is %d long, not %d", len(data), value.Len())
			}

			reflect.Copy(value, reflect.ValueOf(data).Slice(0, value.Len()))
		} else {
			value.SetBytes(data)
		}
		return nil
	}

	k := value.Kind()
	switch {
	case k >= reflect.Int && k <= reflect.Int64:
		i, err := strconv.ParseInt(word, 0, 64)
		if err != nil {
			return err
		}
		value.SetInt(i)
	case k >= reflect.Uint && k <= reflect.Uint64:
		u, err := strconv.ParseUint(word, 0, 64)
		if err != nil {
			return err
		}
		value.SetUint(u)
	case k == reflect.String:
		value.SetString(word)
	default:
		return fmt.Errorf("unhandled type %s", value.Type())
	}

	return nil
}

func isHexData(t reflect.Value) bool {
	arrayType := (t.Kind() == reflect.Array || t.Kind() == reflect.Slice)
	if !arrayType {
		return false
	}

	return t.Type().Elem().Kind() == reflect.Uint8
}

func (t *Decoder) decodeField(value reflect.Value) error {
	words, err := t.decodeWords()
	if err != nil {
		return err
	}

	return t.decodeWordsField(words, value)
}

func (t *Decoder) decodeWordsField(words []string, value reflect.Value) error {
	if isHexData(value) {
		return t.decodeSingleValue(words[0], value)
	}

	if value.Type().Kind() == reflect.Array {
		if len(words) > value.Len() {
			return fmt.Errorf("fixed array size exceeded")
		}

		for i, word := range words {
			if err := t.decodeSingleValue(word, value.Index(i)); err != nil {
				return err
			}
		}
		return nil
	}

	if value.Type().Kind() == reflect.Slice {
		value.Set(reflect.MakeSlice(value.Type(), len(words), len(words)))
		for i, word := range words {
			if err := t.decodeSingleValue(word, value.Index(i)); err != nil {
				return err
			}
		}
		return nil
	}

	if len(words) > 1 {
		return fmt.Errorf("multiple words for a single type")
	}

	return t.decodeSingleValue(words[0], value)
}

func findStructMember(_struct reflect.Type, fieldName string) (int, error) {
	for i := 0; i < _struct.NumField(); i++ {
		field := _struct.Field(i)
		nameInFile := field.Tag.Get("ccfg")

		if fieldName == nameInFile {
			return i, nil
		}

		if fieldName == field.Name {
			return i, nil
		}
	}

	return 0, fmt.Errorf("could not find struct member %s", fieldName)
}

func isNumber(k reflect.Kind) (bool, uint64) {
	switch k {
	case reflect.Int8:
		return true, 1
	case reflect.Int16:
		return true, 2
	case reflect.Int32:
		return true, 4
	case reflect.Int64:
		return true, 8
	case reflect.Uint8:
		return true, 1
	case reflect.Uint16:
		return true, 2
	case reflect.Uint32:
		return true, 4
	case reflect.Uint64:
		return true, 8
	case reflect.Float32:
		return true, 4
	case reflect.Float64:
		return true, 8
	default:
		return false, 0
	}
}

func (t *Decoder) Decode(v interface{}) error {
	_struct := reflect.ValueOf(v)
	if _struct.Kind() == reflect.Ptr {
		_struct = _struct.Elem()
	}

	isTable := false
	var indices []int

	// Read table format
	if _struct.Kind() == reflect.Slice {
		// not a struct
		table := _struct
		if _struct.Type().Elem().Kind() != reflect.Struct {
			return fmt.Errorf("invalid tabular format")
		}

		// Perform type check
		header, err := t.ReadString('\n')
		if err != nil {
			return err
		}

		header = strings.TrimRight(header, "\r\n")
		types := strings.Split(header, "|")

		isTable = true
		indices = make([]int, len(types))

		for i, typeString := range types {
			nameType := strings.SplitN(typeString, "!", 2)
			name := nameType[0]
			index, err := findStructMember(table.Type().Elem(), name)
			if err != nil {
				return err
			}
			field := table.Type().Elem().Field(index)
			typeTypes := strings.SplitN(nameType[1], ":", 2)
			typeSize, err := strconv.ParseUint(typeTypes[1], 0, 64)
			if err != nil {
				return err
			}
			t := strings.ToUpper(typeTypes[0])
			switch t {
			case "STRING":
				isStringArray := field.Type.Kind() == reflect.Slice && field.Type.Elem().Kind() == reflect.String
				if field.Type.Kind() != reflect.String && !isStringArray {
					return fmt.Errorf("%s is not a string", name)
				}
			case "DEC":
				is, length := isNumber(field.Type.Kind())
				if !is {
					return fmt.Errorf("%s is not a decimal", name)
				}

				if length != typeSize {
					return fmt.Errorf("file encoded with different typesize %d for %s", typeSize, name)
				}
			case "HEX":
				isHex := field.Type.Kind() == reflect.Array || field.Type.Kind() == reflect.Slice
				if isHex {
					isHex = field.Type.Elem().Kind() == reflect.Uint8
				}

				if !isHex {
					return fmt.Errorf("field %s is not a byte array", name)
				}
			default:
				return fmt.Errorf("unknown field type %s", typeTypes[0])
			}
			indices[i] = index
		}
	} else {
		if _struct.Kind() != reflect.Struct {
			return fmt.Errorf("cannot decode non-struct type %s", _struct.Type())
		}
	}

	for {
		linePrefix, err := t.Peek(1)
		if err == io.EOF {
			return nil
		}

		if err != nil {
			return err
		}

		if linePrefix[0] == '#' {
			t.ReadString('\n')
			continue
		}

		if isTable {
			line, err := t.ReadString('\n')
			if err == io.EOF {
				if len(line) == 0 {
					return io.EOF
				}
			} else {
				if err != nil {
					return err
				}
			}

			line = strings.TrimRight(line, "\r\n")
			fields := strings.Split(line, "|")

			if len(fields) != len(indices) {
				return fmt.Errorf("field length mismatch")
			}

			value := reflect.New(_struct.Type().Elem()).Elem()

			for i, field := range fields {
				sField := value.Field(indices[i])
				words := strings.Split(field, " ")
				err = t.decodeWordsField(words, sField)
				if err != nil {
					return err
				}
			}

			_struct.Set(reflect.Append(_struct, value))
			continue
		}

		fieldName, err := t.ReadString('=')
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		// trim =
		fieldName = fieldName[:len(fieldName)-1]
		fieldName = strings.TrimSpace(fieldName)

		fieldIndex, err := findStructMember(_struct.Type(), fieldName)
		if err != nil {
			return err
		}

		err = t.decodeField(_struct.Field(fieldIndex))
		if err != nil {
			return err
		}
	}
}
