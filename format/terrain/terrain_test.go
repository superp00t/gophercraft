package terrain

import (
	"fmt"
	"testing"
)

func TestMapReader(t *testing.T) {
	src := &Dir{"E:\\Gaymes\\Work\\"}

	reader, err := NewMapReader(src, 5875, "Azeroth")
	if err != nil {
		t.Fatal(err)
	}

	cnk, err := reader.GetChunkByPos(-9460.25, 63.0612)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println("Area id", cnk.AreaID)

	// for i, v := range reader.Tiles {
	// 	fmt.Println(i/64, i%64, v.Flags)
	// }

	// tile, err := reader.ReadTile(7, 60)
	// if err != nil {
	// 	t.Fatal(err)
	// }

	// for x := 0;

	// fmt.Println(tile.ChunkData[3].AreaID)
}
