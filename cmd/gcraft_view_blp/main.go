package main

import (
	"bytes"
	"image/png"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/superp00t/etc"

	"github.com/superp00t/gophercraft/format/blp"
)

func viewBLP(path string) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}
	texture, err := blp.DecodeBLP(data)
	if err != nil {
		panic(err)
	}

	img := texture.Mipmap(0)
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
