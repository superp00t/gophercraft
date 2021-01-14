package terrain

import (
	"io"
	"os"

	"github.com/superp00t/etc"
)

type Source interface {
	ReadFile(at string) (io.ReadCloser, error)
	Exists(at string) bool
}

type Dir struct {
	Location string
}

func (d *Dir) path(at string) string {
	path := etc.ParseSystemPath(d.Location)
	sub := etc.ParseUnixPath(at)
	pth := path.GetSub(sub).Render()
	// fmt.Println(pth)
	return pth
}

func (d *Dir) ReadFile(at string) (io.ReadCloser, error) {
	return os.Open(d.path(at))
}

func (d *Dir) Exists(at string) bool {
	_, err := os.Stat(d.path(at))
	return err == nil
}
