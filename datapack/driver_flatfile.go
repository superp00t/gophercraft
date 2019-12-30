package datapack

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/superp00t/etc"
)

type File io.ReadCloser

// flatFile implements Driver
type flatFile struct {
	Base etc.Path
}

func (f *flatFile) Init(at string) (Opts, error) {
	f.Base = etc.ParseSystemPath(at)
	if f.Base.IsExtant() == false || f.Base.IsDirectory() == false {
		return None, fmt.Errorf("datapack: source does not exist or is is not a directory: check if you're using the correct driver")
	}

	return Read | Write, nil
}

func (f *flatFile) ReadFile(path string) (File, error) {
	pth := strings.Split(path, "/")
	file, err := os.OpenFile(f.Base.Concat(pth...).Render(), os.O_RDWR, 0700)
	if err != nil {
		return nil, err
	}

	return file, nil
}

func (f *flatFile) WriteFile(path string) (WriteFile, error) {
	pth := strings.Split(path, "/")

	ffile := f.Base.Concat(pth...)

	basePath := ffile[:len(ffile)-1]
	basePath.MakeDir()

	file, err := os.OpenFile(ffile.Render(), os.O_RDWR|os.O_CREATE, 0700)
	if err != nil {
		return nil, err
	}

	return file, nil
}

func (f *flatFile) Close() error {
	// nothing to do here, really
	return nil
}

func (f *flatFile) List() []string {
	var files []string

	err := filepath.Walk(f.Base.Render(), func(path string, info os.FileInfo, err error) error {
		if info.IsDir() == false {
			path = strings.Replace(path, f.Base.Render(), "", 1)
			path = strings.TrimLeft(path, "\\/")
			files = append(files, path)
		}
		return nil
	})

	if err != nil {
		panic(err)
	}

	return files
}
