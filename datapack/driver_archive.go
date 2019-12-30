package datapack

import (
	"archive/zip"
	"fmt"
	"io"
)

var (
	ErrFileNotFound = fmt.Errorf("update: file not found")
)

// archive implements Driver

// .zip
type archive struct {
	Path string
	zr   *zip.ReadCloser
}

func (a *archive) Init(at string) (Opts, error) {
	var err error
	a.Path = at
	a.zr, err = zip.OpenReader(at)
	if err != nil {
		return 0, err
	}

	return Read, nil
}

func (a *archive) WriteFile(path string) (WriteFile, error) {
	return nil, fmt.Errorf("datapack: ZIP archives are read-only. You can extract this zip to a folder if you need to make a modification!")
}

func (a *archive) ReadFile(path string) (File, error) {
	for _, v := range a.zr.File {
		if path == v.Name {
			return v.Open()
		}
	}

	return nil, ErrFileNotFound
}

func (a *archive) Close() error {
	// nothing to do here, really
	return nil
}

func (f *archive) List() []string {
	var s []string

	for _, v := range f.zr.File {
		s = append(s, v.Name)
	}

	return s
}

func addFileToZip(zipWriter *zip.Writer, filename string, rdr io.Reader) error {
	header := &zip.FileHeader{}
	// Using FileInfoHeader() above only uses the basename of the file. If we want
	// to preserve the folder structure we can overwrite this with the full path.
	header.Name = filename

	// Change to deflate to gain better compression
	// see http://golang.org/pkg/archive/zip/#pkg-constants
	header.Method = zip.Deflate

	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}
	_, err = io.Copy(writer, rdr)
	return err
}
