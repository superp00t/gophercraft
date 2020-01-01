package content

import (
	"fmt"

	"github.com/superp00t/gophercraft/vsn"
)

type Volume interface {
	Build() vsn.Build
	ListFiles() ([]string, error)
	ReadFile(at string) ([]byte, error)
	Close() error
}

func Open(path string) (Volume, error) {
	v, err := vsn.DetectGame(path)
	if err != nil {
		return nil, err
	}

	vt, path2, err := vsn.DetectVolumeLocation(path)
	if err != nil {
		return nil, err
	}

	switch vt {
	case vsn.NGDP:
		return nil, fmt.Errorf("NGDP nyi")
	case vsn.MPQ:
		return openMpq(v, path2)
	default:
		return nil, fmt.Errorf("unknown folder type")
	}
}
