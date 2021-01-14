package datapack

import (
	"archive/zip"
	"fmt"
	"io"
	"strings"
)

var (
	ErrFileNotFound = fmt.Errorf("datapack: file not found")
)

// archive implements Driver

// .zip
type archive struct {
	Path   string
	Prefix string
	zr     *zip.ReadCloser
}

func (a *archive) Init(at string) (Opts, error) {
	var err error
	a.Path = at
	a.zr, err = zip.OpenReader(at)
	if err != nil {
		return 0, err
	}

	fl, err := a.ReadFile("Pack.txt")
	if err != nil {
		for _, file := range a.zr.File {
			if file.FileInfo().IsDir() {
				// Top level directory
				if strings.HasSuffix(file.Name, "/") && strings.Count(file.Name, "/") == 1 {
					a.Prefix = file.Name
					break
				}
			}
		}
	} else {
		fl.Close()
	}

	return Read, nil
}

func (a *archive) WriteFile(path string) (WriteFile, error) {
	return nil, fmt.Errorf("datapack: ZIP archives are read-only. You can extract this zip to a folder if you need to make a modification!")
}

func (a *archive) ReadFile(path string) (File, error) {
	for _, v := range a.zr.File {
		if a.Prefix+path == v.Name {
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
		if v.Name != f.Prefix {
			s = append(s, strings.TrimPrefix(v.Name, f.Prefix))
		}
	}

	return s
}

func addFileToZip(zipWriter *zip.Writer, filename string, rdr io.Reader) error {
	header := &zip.FileHeader{}

	header.Name = filename
	header.Method = zip.Deflate

	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}
	_, err = io.Copy(writer, rdr)
	return err
}
