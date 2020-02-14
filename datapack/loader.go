package datapack

import (
	"fmt"
	"io"
	"reflect"
	"strings"

	"github.com/superp00t/gophercraft/datapack/csv"
)

type Loader struct {
	Volumes []*Pack
}

func (ld *Loader) Exists(path string) bool {
	seg := strings.SplitN(path, ":", 2)
	for _, pack := range ld.Volumes {
		if pack.Name == seg[0] {
			return pack.Exists(seg[1])
		}
	}

	return false
}

func (ld *Loader) Open(path string) (io.ReadCloser, error) {
	seg := strings.SplitN(path, ":", 2)
	for _, pack := range ld.Volumes {
		if pack.Name == seg[0] {
			if pack.Exists(seg[1]) {
				return pack.ReadFile(seg[1])
			}
		}
	}

	return nil, fmt.Errorf("file %s not found", path)
}

func (ld *Loader) ReadAll(path string, slicePtr interface{}) int {
	read := 0

	valuePtr := reflect.ValueOf(slicePtr)
	if valuePtr.Kind() != reflect.Ptr {
		panic("expected pointer to slice")
	}

	value := valuePtr.Elem()

	valueType := value.Type()

	if valueType.Kind() != reflect.Slice {
		panic("expected slice")
	}

	elemType := valueType.Elem()

	if elemType.Kind() != reflect.Struct {
		panic("expected slice of structs")
	}

	for _, pack := range ld.Volumes {
		if pack.Exists(path) {
			fi, err := pack.ReadFile(path)
			if err != nil {
				panic(err)
			}

			defer fi.Close()

			rd, err := csv.NewScanner(fi)
			if err != nil {
				panic(err)
			}

			for {
				newMade := reflect.New(elemType)

				err := rd.Scan(newMade.Interface())
				if err == io.EOF {
					break
				}

				if err != nil {
					panic(err)
				}

				value.Set(reflect.Append(value, newMade.Elem()))
				read++
			}
		}
	}

	return read
}

func (ld *Loader) Close() {
	for _, v := range ld.Volumes {
		if err := v.Close(); err != nil {
			panic(err)
		}
	}

	ld.Volumes = nil
}
