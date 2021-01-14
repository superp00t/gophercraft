package datapack

import (
	"fmt"
	"io"
	"reflect"
	"strings"

	"github.com/superp00t/gophercraft/datapack/text"
)

type Loader struct {
	Volumes []*Pack
}

func (ld *Loader) Exists(path string) bool {
	colon := strings.Index(path, ":")
	if colon == -1 {
		for _, pack := range ld.Volumes {
			if pack.Exists(path) {
				fmt.Println(path, "does exist in", pack.Name)
				return true
			}
			fmt.Println(path, "doesn't exist in", pack.Name)
		}
	} else {
		pack := path[:colon]
		filePath := path[colon:]

		for _, dpack := range ld.Volumes {
			if dpack.Name == pack {
				if dpack.Exists(filePath) {
					return true
				}
			}
		}
	}

	return false
}

// func (ld *Loader) Open(path string) (io.ReadCloser, error) {
// 	seg := strings.SplitN(path, ":", 2)
// 	for _, pack := range ld.Volumes {
// 		if pack.Name == seg[0] {
// 			if pack.Exists(seg[1]) {
// 				return pack.ReadFile(seg[1])
// 			}
// 		}
// 	}

// 	return nil, fmt.Errorf("file %s not found", path)
// }

func (ld *Loader) ReadFile(path string) (io.ReadCloser, error) {
	colon := strings.IndexByte(path, ':')
	if colon == -1 {
		for _, pack := range ld.Volumes {
			if pack.Exists(path) {
				return pack.ReadFile(path)
			}
		}
	} else {
		pack := path[:colon]
		filePath := path[colon:]

		for _, dpack := range ld.Volumes {
			if dpack.Name == pack {
				if dpack.Exists(filePath) {
					return dpack.ReadFile(filePath)
				}
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

	if elemType.Kind() != reflect.Struct && elemType.Kind() != reflect.Map {
		panic(fmt.Sprintf("expected slice of structs: got %s", elemType.Kind()))
	}

	for _, pack := range ld.Volumes {
		if pack.Exists(path) {
			textFile, err := pack.ReadFile(path)
			if err != nil {
				panic(err)
			}

			decoder := text.NewDecoder(textFile)

			fmt.Println("open", path)

			defer textFile.Close()

			for {
				newMade := reflect.New(elemType)

				err = decoder.Decode(newMade.Interface())
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

func (ld *Loader) List() []string {
	var names []string

	for _, volume := range ld.Volumes {
		for _, volumeListing := range volume.List() {
			found := false
			for _, name := range names {
				if name == volumeListing {
					found = true
				}
			}
			if !found {
				names = append(names, volumeListing)
			}
		}
	}

	return names
}
