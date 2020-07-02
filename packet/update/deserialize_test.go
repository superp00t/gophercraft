package update_test

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"io"
	"io/ioutil"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/davecgh/go-spew/spew"

	"github.com/superp00t/etc/yo"
	"github.com/superp00t/gophercraft/guid"
	"github.com/superp00t/gophercraft/packet/update"
	_ "github.com/superp00t/gophercraft/packet/update/descriptorsupport"
	"github.com/superp00t/gophercraft/vsn"

	"github.com/superp00t/etc"
)

func TestEncodeDecode(t *testing.T) {
	writer := new(bytes.Buffer)

	id := guid.RealmSpecific(guid.Player, 0, 69420)

	enc, err := update.NewEncoder(5875, writer, 1)
	if err != nil {
		t.Fatal(err)
	}

	vblock, err := update.NewValuesBlock(5875, guid.TypeMaskObject|guid.TypeMaskUnit|guid.TypeMaskPlayer)
	if err != nil {
		t.Fatal(err)
	}

	vblock.SetGUID("GUID", id)
	vblock.SetBit("PlayerControlled", true)
	vblock.SetBit("DetectStealth", true)
	vblock.SetGUID("Target", id)

	vblock.SetStructArrayValue("VisibleItems", 3, "Entry", uint32(50))

	yo.Spew(vblock.StorageDescriptor.Elem().FieldByName("PlayerData").FieldByName("VisibleItems").Index(3).FieldByName("Entry").Interface())

	if ent := vblock.GetUint32("VisibleItems", 3, "Entry"); ent != 50 {
		t.Fatal("failed to set")
	}

	enc.AddBlock(id, &update.CreateBlock{
		BlockType:     update.SpawnObject,
		ObjectType:    guid.TypePlayer,
		ValuesBlock:   vblock,
		MovementBlock: &update.MovementBlock{},
	}, update.Owner)

	yo.Spew(writer.Bytes())

	reader := bytes.NewReader(writer.Bytes())
	decoder, err := update.NewDecoder(5875, reader)
	if err != nil {
		t.Fatal(err)
	}

	for decoder.MoreBlocks() {
		bt, err := decoder.NextBlock()
		if err != nil {
			t.Fatal(err)
		}

		switch bt {
		case update.CreateObject, update.SpawnObject:
			id, err := decoder.DecodeGUID()
			if err != nil {
				t.Fatal(err)
			}

			fmt.Println("Object", id, "created")

			createBlock, err := decoder.DecodeCreateBlock()
			if err != nil {
				t.Fatal(err)
			}

			etc.LocalDirectory().Concat("decoded.txt").WriteAll([]byte(spew.Sdump(createBlock.ValuesBlock.StorageDescriptor.Interface())))
		default:
			fmt.Errorf("unhandled blocktype: %s", bt)
		}
	}
}

func TestFwd(t *testing.T) {
	vd := &update.ValuesDecoder{}
	vd.BitPos = 7
	vd.FwdBytes(1)
	if vd.BitPos != 8 {
		t.Fatal(vd.BitPos)
	}

	vd.FwdBytes(1)
	if vd.BitPos != 16 {
		t.Fatal(vd.BitPos)
	}

	vd.FwdBits(3)
	vd.FwdBytes(1)
	if vd.BitPos != 24 {
		t.Fatal(vd.BitPos)
	}
}

func TestBitmask(t *testing.T) {
	bm := &update.Bitmask{0, 0}
	bm.Set(86, true)
	if !bm.Enabled(86) {
		t.Fatal("86 was not set")
	}

	buffer := bytes.NewBuffer(nil)
	if err := update.WriteBitmask(bm, update.Descriptors[5875], buffer); err != nil {
		t.Fatal(err)
	}

	yo.Spew(buffer.Bytes())

	buffer = bytes.NewBuffer(buffer.Bytes())

	yo.Spew(buffer.Bytes())

	bm2, err := update.ReadBitmask(update.Descriptors[5875], buffer)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(*bm, *bm2) {
		fmt.Println(bm)
		fmt.Println(bm2)
		t.Fatal("mismatch")
	}
}

func TestDescriptor(t *testing.T) {
	// desc := update.Descriptors[5875]

	// cpp, err := desc.GenerateCPP()
	// if err != nil {
	// 	panic(err)
	// }

	// fmt.Println(cpp)
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

		reader := bytes.NewReader(data)
		decoder, err := update.NewDecoder(vsn.Build(v.Version), reader)
		if err != nil {
			t.Fatal(err)
		}

		for decoder.MoreBlocks() {
			bt, err := decoder.NextBlock()
			if err != nil {
				t.Fatal(err)
			}

			switch bt {
			case update.CreateObject, update.SpawnObject:
				id, err := decoder.DecodeGUID()
				if err != nil {
					t.Fatal(err)
				}

				fmt.Println("Object", id, "created")

				createBlock, err := decoder.DecodeCreateBlock()
				if err != nil {
					t.Fatal(err)
				}

				yo.Spew(createBlock)
			default:
				fmt.Errorf("unhandled blocktype: %s", bt)
			}
		}
	}
}
