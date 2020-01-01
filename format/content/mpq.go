package content

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/superp00t/gophercraft/format/mpq"
	"github.com/superp00t/gophercraft/vsn"
)

type mpqp struct {
	build vsn.Build
	pool  *mpq.Pool
}

func (m *mpqp) Build() vsn.Build {
	return m.build
}

func openMpq(version vsn.Build, path string) (*mpqp, error) {
	m := &mpqp{build: version}

	var names []string

	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err == nil {
			if strings.HasSuffix(path, ".MPQ") {
				names = append(names, path)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	m.pool, err = mpq.OpenPool(names)
	if err != nil {
		return nil, err
	}

	return m, nil
}

func (m *mpqp) ListFiles() ([]string, error) {
	fl := m.pool.ListFiles()
	return fl, nil
}

func (m *mpqp) ReadFile(at string) ([]byte, error) {
	fl, err := m.pool.OpenFile(at)
	if err != nil {
		return nil, err
	}

	data, err := fl.ReadBlock()
	fl.Close()

	return data, err
}

func (m *mpqp) Close() error {
	m.pool = nil
	return nil
}
