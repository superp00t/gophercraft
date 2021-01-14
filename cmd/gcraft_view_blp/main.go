package main

import (
	"bytes"
	"image/png"
	"os"
	"os/exec"

	"github.com/superp00t/etc"

	"github.com/superp00t/gophercraft/format/blp"
)

func viewBLP(path string) {
	data, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer data.Close()
	img, err := blp.Decode(data)
	if err != nil {
		panic(err)
	}

	out := new(bytes.Buffer)
	png.Encode(out, img)

	pth := etc.LocalDirectory().Concat("blpviewer.png")
	pth.WriteAll(out.Bytes())

	exec.Command("explorer.exe", pth.Render()).Run()
}

func main() {
	if len(os.Args) < 2 {
		return
	}

	blpPath := os.Args[1]

	viewBLP(blpPath)
}
