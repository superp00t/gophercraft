package update_test

import (
	"compress/zlib"
	"fmt"
	"io"
	"io/ioutil"
	"strconv"
	"strings"
	"testing"

	"github.com/davecgh/go-spew/spew"

	"github.com/superp00t/etc/yo"
	"github.com/superp00t/gophercraft/packet/update"

	"github.com/superp00t/etc"
)

func TestDescriptor(t *testing.T) {
	desc := update.Descriptors[5875]

	cpp, err := desc.GenerateCPP()
	if err != nil {
		panic(err)
	}

	fmt.Println(cpp)
}

type capture struct {
	Version     uint32
	Description string
	Compression bool
	Data        []byte
}

// Check the ability of this package to successfully parse known-good packet captures.
func TestUnmarshal(t *testing.T) {
	spew.Config.SortKeys = true

	var captures []capture

	captureDir := etc.Import("github.com/superp00t/gophercraft/packet/update/testdata")

	list, err := ioutil.ReadDir(captureDir.Render())
	if err != nil {
		t.Fatal(err)
	}

	for _, v := range list {
		if !v.IsDir() {
			if strings.HasSuffix(v.Name(), ".bin") {
				elements := strings.Split(v.Name(), ".")
				vsn, err := strconv.ParseInt(elements[0], 0, 64)
				if err != nil {
					t.Fatal(err)
				}

				data, err := captureDir.Concat(v.Name()).ReadAll()
				if err != nil {
					t.Fatal(err)
				}

				cap := capture{
					Version:     uint32(vsn),
					Description: elements[1],
					Compression: elements[2] == "compressed",
					Data:        data,
				}

				captures = append(captures, cap)
			}
		}
	}

	for _, v := range captures {
		data := v.Data

		if v.Compression {
			dataBuffer := etc.FromBytes(data)
			decompressedSize := dataBuffer.ReadUint32()
			if decompressedSize > 2e+8 {
				t.Fatal("decompressed size is too big")
			}

			z, err := zlib.NewReader(dataBuffer)
			if err != nil {
				t.Fatal("zlib", err)
			}

			out := make([]byte, decompressedSize)
			_, err = z.Read(out)
			if err != nil && err != io.EOF {
				t.Fatal(err)
			}

			data = out
		}

		upd, err := update.Unmarshal(v.Version, data)
		if err != nil {
			t.Fatal(err)
		}

		yo.Spew(upd)
	}
}
