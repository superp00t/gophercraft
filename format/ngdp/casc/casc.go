/*Package casc [soon™] implements the Content-Addressable Storage Container (CASC) format.
 */
package casc

import "io/ioutil"

type FileID uint64

type Archive struct {
	Flags
	Path string
}

// Open a CASC archive, assuming it already exists.
func Open(path string) (*Archive, error) {
	archive := &Archive{}
	archive.Flags

	list, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}

}

type File struct{}

func (a *Archive) OpenFile(fileID FileID) (*File, error) {

}
