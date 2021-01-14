package blp

import (
	"fmt"
	"image/png"
	"os"
	"testing"

	"github.com/superp00t/etc"
)

func TestBLP(t *testing.T) {
	projectFolder := etc.Import("github.com/superp00t/gophercraft/format/blp/testdata")

	tFile, err := os.Open(projectFolder.Concat("rats.png").Render())
	if err != nil {
		t.Fatal(err)
	}

	img, err := png.Decode(tFile)
	if err != nil {
		t.Fatal(err)
	}

	tFile.Close()

	tFileOut, err := os.OpenFile(projectFolder.Concat("results", "rats.blp").Render(), os.O_CREATE|os.O_RDWR, 0700)
	if err != nil {
		t.Fatal(err)
	}

	if err := Encode(img, tFileOut, TextureUncompressed); err != nil {
		t.Fatal(err)
	}

	for testBLP := 1; testBLP <= 9; testBLP++ {
		srcFile, err := os.Open(projectFolder.Concat(fmt.Sprintf("test%d.blp", testBLP)).Render())
		if err != nil {
			t.Fatal(err)
		}

		fmt.Println("decoding", testBLP)

		texture, err := Decode(srcFile)
		if err != nil {
			t.Fatal(err)
		}

		srcFile.Close()
		output, err := os.OpenFile(projectFolder.Concat("results", fmt.Sprintf("test%d.png", testBLP)).Render(), os.O_CREATE|os.O_RDWR, 0700)
		if err != nil {
			t.Fatal(err)
		}

		if err := png.Encode(output, texture); err != nil {
			t.Fatal(err)
		}

		output.Close()
		// exec.Command("explorer.exe", path.Render())
	}
}
