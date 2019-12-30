package guid

import (
	"fmt"
	"testing"

	"github.com/superp00t/etc"
	"github.com/superp00t/etc/yo"
)

func TestGUID(t *testing.T) {
	var guid1 uint64 = 0x0000DEADC0DEBBBB

	g := Classic(guid1)

	if g.HighType() != Player {
		t.Fatal("mismatch")
	}

	if g.Classic() != guid1 {
		fmt.Printf("0x%16X\n", guid1)
		panic("data loss in encoded guid")
	}

	fmt.Println(g)

	buffer := etc.NewBuffer()
	g.EncodePacked(5875, buffer)
	yo.Spew(buffer.Bytes())

	gd, err := DecodePacked(5875, buffer)
	if err != nil {
		panic(err)
	}

	fmt.Println(g, gd)

	if gd != g {
		t.Fatal("Lossy codec", g, gd)
	}
}

// func testEncoding(t *testing.T, guid1 GUID) {
// 	fmt.Println("Before encoding: ", guid1)

// 	bts := guid1.EncodePacked()
// 	fmt.Println("Encoded as bytes:", bts)
// 	e2 := etc.MkBuffer(bts)
// 	// append some data, so we can be sure that the decoder works
// 	// even when the packed GUID data is followed immediately by other data
// 	e2.WriteUint32(0xFFFFFFFF)
// 	g2 := DecodePacked(e2)
// 	fmt.Println("Decoded from bytes:", g2)

// 	safeData := e2.ReadUint32()
// 	if safeData != 0xFFFFFFFF {
// 		t.Fatal("mismatch")
// 	}
// }

// func TestEncoding(t *testing.T) {
// 	for _, g := range []GUID{
// 		0x0000000000521BC0,
// 		0xDEADBEEF1337BADC,
// 	} {
// 		testEncoding(t, g)
// 	}
// }
