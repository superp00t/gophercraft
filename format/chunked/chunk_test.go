package chunked

import (
	"fmt"
	"os"
	"testing"
)

func TestChunkedReader(t *testing.T) {
	file, err := os.Open("Work\\World\\Maps\\Azeroth\\Azeroth.wdt")
	if err != nil {
		panic(err)
	}

	reader := &Reader{file}

	for {
		id, data, err := reader.ReadChunk()
		fmt.Println(id)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println(len(data))
	}
}
