package main

import (
	"os"
	"runtime"
	"strings"

	"github.com/mitchellh/go-ps"
	"github.com/superp00t/etc"
	"github.com/superp00t/gophercraft/vsn"
)

var wizGameDir = ""

// cmd.exe doesn't support emojis :(
func doItWithStyle() bool {
	switch runtime.GOOS {
	case "windows":
		ppid := os.Getppid()
		proc, err := ps.FindProcess(ppid)
		if err != nil {
			return false
		}

		if strings.HasSuffix(proc.Executable(), "explorer.exe") {
			return false
		}

		if strings.HasSuffix(proc.Executable(), "cmd.exe") {
			return false
		}

		return true
	default:
		return true
	}

}

func main() {
	if doItWithStyle() {
		wiz = "🧙🏿"
	} else {
		wiz = "[Wiz] "
	}

	if len(os.Args) > 1 {
		wizGameDir = os.Args[1]
	}

	vsn.PrintBanner()
	wizStart()
}

func getFolder() etc.Path {
	return etc.LocalDirectory().Concat("Gophercraft")
}

// func getAuth() *config.Auth {
// 	authLoc := getFolder().Concat("Auth")

// 	if !authLoc.IsExtant() {
// 		return nil
// 	}

// 	a, err := config.LoadAuth(authLoc.Render())
// 	if err != nil {
// 		yo.Fatal(err)
// 	}

// 	return a
// }
