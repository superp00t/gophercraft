package text

import (
	"fmt"
	"io"
	"reflect"
	"sort"
	"strconv"
	"strings"
)

type Encoder struct {
	out    io.Writer
	Indent string
}

func NewEncoder(out io.Writer) *Encoder {
	return &Encoder{
		out:    out,
		Indent: "\t",
	}
}

func (encoder *Encoder) writeIndentation(depth int) {
	for x := 0; x < depth; x++ {
		if _, err := encoder.out.Write([]byte(encoder.Indent)); err != nil {
			return
		}
	}
}

func (encoder *Encoder) Encode(i interface{}) error {
	value := reflect.ValueOf(i)
	if value.Kind() == reflect.Ptr {
		value = value.Elem()
	}
	return encoder.encodeValue(0, value)
}

func (encoder *Encoder) encodeString(str string) error {
	if (strings.Contains(str, " ") ||
		strings.Contains(str, "\n") ||
		strings.Contains(str, "\t") ||
		strings.Contains(str, "\r") ||
		strings.Contains(str, "'") ||
		strings.Contains(str, "\"")) == false {
		_, err := encoder.out.Write([]byte(str))
		return err
	}

	str = strings.Replace(str, "\n", "\\n", -1)
	str = strings.Replace(str, "\t", "\\t", -1)
	str = strings.Replace(str, "\r", "\\r", -1)
	str = strings.Replace(str, "\"", "\\\"", -1)

	_, err := encoder.out.Write([]byte("\"" + str + "\""))
	return err
}

func isCurly(field reflect.Value) bool {
	return field.Kind() == reflect.Struct || field.Kind() == reflect.Array || field.Kind() == reflect.Slice || field.Kind() == reflect.Map
}

func (encoder *Encoder) encodeValue(depth int, value reflect.Value) error {
	encoder.writeIndentation(depth)

	switch value.Kind() {
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if _, err := encoder.out.Write([]byte(strconv.FormatUint(value.Uint(), 10))); err != nil {
			return err
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if _, err := encoder.out.Write([]byte(strconv.FormatInt(value.Int(), 10))); err != nil {
			return err
		}
	case reflect.Float32, reflect.Float64:
		if _, err := encoder.out.Write([]byte(strconv.FormatFloat(value.Float(), 'f', -1, bitSize(value.Kind())))); err != nil {
			return err
		}
	case reflect.Bool:
		if _, err := encoder.out.Write([]byte(strconv.FormatBool(value.Bool()))); err != nil {
			return err
		}
	case reflect.String:
		if err := encoder.encodeString(value.String()); err != nil {
			return err
		}
	case reflect.Slice, reflect.Array:
		encoder.out.Write([]byte("{\n"))
		for x := 0; x < value.Len(); x++ {
			if err := encoder.encodeValue(depth+1, value.Index(x)); err != nil {
				return err
			}

			if !isCurly(value.Index(x)) {
				encoder.out.Write([]byte("\n"))
			}
		}
		for x := 0; x < depth; x++ {
			if _, err := encoder.out.Write([]byte(encoder.Indent)); err != nil {
				return err
			}
		}
		encoder.out.Write([]byte("}\n"))
	case reflect.Struct:
		encoder.out.Write([]byte("{\n"))

		for x := 0; x < value.NumField(); x++ {
			field := value.Field(x)

			if field.IsZero() {
				continue
			}

			encoder.writeIndentation(depth + 1)
			encoder.out.Write([]byte(value.Type().Field(x).Name))

			if isCurly(field) {
				encoder.out.Write([]byte("\n"))
				if err := encoder.encodeValue(depth+1, field); err != nil {
					return err
				}
			} else {
				encoder.out.Write([]byte(" "))
				if err := encoder.encodeValue(0, field); err != nil {
					return err
				}
				encoder.out.Write([]byte("\n"))
			}
		}

		encoder.writeIndentation(depth)

		encoder.out.Write([]byte("}\n"))
	case reflect.Map:
		encoder.out.Write([]byte("{\n"))
		mKeys := value.MapKeys()
		sortValues(mKeys)

		for _, key := range mKeys {
			if err := encoder.encodeValue(depth+1, key); err != nil {
				return err
			}

			field := value.MapIndex(key)

			if isCurly(field) {
				encoder.out.Write([]byte("\n"))
				if err := encoder.encodeValue(depth+1, field); err != nil {
					return err
				}
			} else {
				encoder.out.Write([]byte(" "))
				if err := encoder.encodeValue(0, field); err != nil {
					return err
				}
				encoder.out.Write([]byte("\n"))
			}
		}
		encoder.writeIndentation(depth)
		encoder.out.Write([]byte("}\n"))
	default:
		return fmt.Errorf("unknown kind %s", value.Kind())
	}

	return nil
}

type valueSorter []reflect.Value

func (vs valueSorter) Len() int {
	return len(vs)
}

func (vs valueSorter) Less(i, j int) bool {
	switch vs[0].Kind() {
	case reflect.String:
		return vs[i].String() < vs[j].String()
	case reflect.Uint:
		return vs[i].Uint() < vs[j].Uint()
	case reflect.Int:
		return vs[i].Int() < vs[j].Int()
	case reflect.Float32:
		return vs[i].Float() < vs[j].Float()
	default:
		panic("unknown kind " + vs[0].Kind().String())
	}
}

func (vs valueSorter) Swap(i, j int) {
	_i := vs[i]
	_j := vs[j]
	vs[i] = _j
	vs[j] = _i
}

func sortValues(values []reflect.Value) {
	if len(values) == 0 {
		return
	}

	sorter := valueSorter(values)
	sort.Sort(sorter)
}
