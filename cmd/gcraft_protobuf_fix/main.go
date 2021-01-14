package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/superp00t/etc"
	"github.com/superp00t/etc/yo"
)

func main() {
	yo.MainArgs([]string{"source", "out"}, "fix some small inconsistencies in the protobuf definitions and compiles them to go", _main)
	yo.Init()
}

func _main(args []string) {

	cohorts := make(map[string][]string)

	path := etc.ParseSystemPath(os.Getenv("GOPATH")).
		Concat("src", "github.com", "superp00t", "gophercraft", "bnet", "protos").
		Render()

	filepath.Walk(path, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if fi.IsDir() == false {
			if strings.HasSuffix(fi.Name(), ".proto") {
				data, err := ioutil.ReadFile(path)
				if err != nil {
					yo.Fatal(err)
				}

				sdata := strings.Split(string(data), "\n")

				out := []string{}

				for i, v := range sdata {
					if fi.Name() == "club_notification.proto" {
						v = strings.Replace(v, "import public", "import", -1)
					}

					if strings.HasPrefix(v, "package") {
						yo.Ok(i, v)
						s := strings.Split(v, " ")[1]
						s = strings.TrimRight(s, ";")
						cohorts[s] = append(cohorts[s], path)
						pkgPath := strings.Join(strings.Split(s, "."), "/")

						out = append(out, v)

						out = append(out, fmt.Sprintf(`option go_package = "github.com/superp00t/gophercraft/bnet/%s";`, pkgPath))
					} else {
						out = append(out, v)
					}
				}

				ioutil.WriteFile(path, []byte(strings.Join(out, "\n")), 0700)
			}
		}

		return nil
	})

	gcraftPath := os.Getenv("GOPATH") + "src/github.com/superp00t/gophercraft/bnet/"

	for k, v := range cohorts {
		fmt.Println("Compiling package", k)

		s := append(v, "--gcraft_out=plugins=bnet_rpc:"+os.Getenv("GOPATH")+"src/")

		c := exec.Command("protoc", append([]string{"-I" + gcraftPath + "protos"}, s...)...)
		c.Stderr = os.Stdout
		c.Stdout = os.Stdout
		if err := c.Run(); err != nil {
			yo.Fatal(err)
		}
	}
}

func hashServiceName(name string) uint32 {
	var hash uint32 = 0x811C9DC5
	for i := 0; i < len(name); i++ {
		hash ^= uint32(name[i])
		hash *= 0x1000193
	}

	return hash
}
