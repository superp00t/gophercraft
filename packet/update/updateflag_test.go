package update

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/superp00t/etc"
)

func TestUpdateFlags(t *testing.T) {
	plyrf := uint32(0x0001 | 0x0020 | 0x0040)

	buf := etc.NewBuffer()

	encodeUpdateFlags(5875, buf, UpdateFlagSelf|UpdateFlagLiving|UpdateFlagHasPosition)

	if ac := buf.ReadUint32(); ac != plyrf {
		t.Fatal(ac)
	}

	tbc := bytes.NewReader([]byte{0x71})

	tbcUf, err := decodeUpdateFlags(8606, tbc)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(tbcUf)
}
