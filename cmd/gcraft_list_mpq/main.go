package main

import (
	"fmt"
	"os"

	"github.com/superp00t/etc/yo"
	"github.com/superp00t/gophercraft/format/mpq"
	"github.com/superp00t/gophercraft/vsn"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("gcraft_list_mpq /path/to/my/game/directory")
		return
	}

	var fp = os.Args[1]

	build, err := vsn.DetectGame(fp)
	if err != nil {
		yo.Fatal(err)
	}

	yo.Ok(build, "detected")

	s, err := mpq.GetFiles(fp)
	if err != nil {
		yo.Fatal(err)
	}

	m, err := mpq.OpenPool(s)
	if err != nil {
		yo.Fatal(err)
	}

	for _, v := range m.ListFiles() {
		fmt.Println(v)
	}
}
