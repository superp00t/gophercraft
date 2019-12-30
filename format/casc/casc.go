/*Package casc [soon™] implements a decoder for the CASC format
 */
package casc

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/superp00t/etc"
)

type CASC struct {
	Path etc.Path
}

func (c *CASC) Subpath(path []string) string {
	return c.BasePath + filepath.Join(path...)
}

func (c *CASC) GetFile(path []string) []byte {
	sp := c.Subpath(path)
	if !exists(sp) {
		panic("could not find " + sp)
	}

	b, _ := ioutil.ReadFile(sp)
	return b
}

func exists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		return false
	}

	return true
}

func OpenCASC(path string) (*CASC, error) {
	p := new(CASC)

	if !exists(path) {
		return nil, os.ErrNotExist
	}

	p.BasePath = path

	return nil, nil
}
