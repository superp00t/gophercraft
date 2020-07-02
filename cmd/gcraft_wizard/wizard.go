package main

import (
	"crypto/tls"
	"errors"
	"fmt"
	"os"

	"github.com/AlecAivazis/survey"
	"github.com/superp00t/etc"
	"github.com/superp00t/gophercraft/datapack"
	"github.com/superp00t/gophercraft/format/content"
	"github.com/superp00t/gophercraft/gcore/config"
	"github.com/superp00t/gophercraft/gcore/sys"
)

var (
	wizGamePool  content.Volume
	wizWorldID   uint64
	wizWorldPath string
	wizPack      *datapack.Pack
)

// func wizConfirm(str string) bool {
// 	fmt.Println(str, "[y/n]")
// 	var input [1]byte
// 	for {
// 		if _, err := os.Stdin.Read(input[:]); err != nil {
// 			return false
// 		}

// 		if input[0] == 'y' {
// 			return true
// 		}

// 		if input[0] == 'n' {
// 			return false
// 		}
// 	}
// }

// func wizClearScreen() {
// 	switch runtime.GOOS {
// 	case "windows":
// 		fmt.Print("\x0c")
// 	default:
// 		fmt.Print("\033[2J")
// 	}
// }

// func wizGetLine() string {
// 	str, err := bufio.NewReader(os.Stdin).ReadString('\n')
// 	if err != nil {
// 		wizQuit(err)
// 	}
// 	return strings.TrimRight(str, "\r\n")
// }

// func wizGetOption(question string, selections []string) int {
// 	var activeSelection = 0
// 	for {
// 		if activeSelection >= len(selections) || activeSelection < 0 {
// 			activeSelection = 0
// 		}

// 		wizClearScreen()
// 		fmt.Println(question)
// 		for i, sel := range selections {
// 			if activeSelection == i {
// 				fmt.Printf("*")
// 			}
// 			fmt.Printf(") %d: %s\n", i+1, sel)
// 		}
// 		fmt.Printf("> ")
// 		var input [1]byte
// 		os.Stdin.Read(input[:])
// 		if input[0] == '\n' {
// 			return activeSelection
// 		}

// 		if input[0] == '\t' {
// 			activeSelection++
// 			continue
// 		}

// 		if input[0] >= '1' && input[0] <= '9' {
// 			activeSelection = int(input[0] - '1')
// 		}
// 	}
// }

func wizQuit(err error) {
	if err == nil {
		fmt.Println("exiting...")
		os.Exit(0)
	}
	fmt.Println(err)
	os.Exit(0)
}

func wizGetGame() {
	var line string

	if err := survey.AskOne(&survey.Input{
		Message: "Input your game directory to continue.",
	},
		&line); err != nil {
		wizQuit(err)
	}

	if line == "" {
		wizQuit(errors.New("Game files are needed to continue the wizard."))
	}

	var err error
	wizGamePool, err = content.Open(line)
	if err != nil {
		wizQuit(err)
	}
}

func wizStart() {
	wizGetGame()

	const (
		justQuit  = "Quit"
		packOnly  = "Continue without setting up server"
		complete  = "Setup Gophercraft Auth and Worldserver on this computer"
		worldOnly = "Setup Worldserver on this computer using remote Auth Server"
	)

	qs := &survey.Select{
		Message: "Choose a deployment configuration.",
		Options: []string{
			justQuit,
			packOnly,
			complete,
			worldOnly,
		},
	}

	var opt int
	err := survey.AskOne(qs, &opt)
	if err != nil {
		wizQuit(err)
	}

	switch opt {
	case 0:
		wizQuit(nil)
	case 1:
		wizDatapack()
	case 2:
		wizSetupAllCores()
		wizDatapack()
	case 3:
		wizSetupWorldServerStandalone()
		wizDatapack()
	}
}

func wizSetupAllCores() {
	var line string
	authLoc := getLocal().Concat("gcraft_auth")

	err := survey.AskOne(&survey.Input{
		Message: fmt.Sprint("Set auth core location? if left empty, I will use default: ", authLoc.Render()),
	}, &line)
	if err != nil {
		wizQuit(err)
	}

	if line == "" {
		line = authLoc.Render()
	}

	authLoc = etc.ParseSystemPath(line)
	if !authLoc.IsExtant() {
		if err := config.GenerateDefaultAuth(authLoc.Render()); err != nil {
			wizQuit(err)
		}
	} else {
		wizQuit(fmt.Errorf("path %s already exists.", authLoc.Render()))
	}

	authconfig, err := config.LoadAuth(authLoc.Render())
	if err != nil {
		wizQuit(err)
	}

	wizWorldID = 1

	worldLoc := getLocal().Concat("gcraft_world_1")
	wizWorldPath = worldLoc.Render()

	if err := authconfig.GenerateDefaultWorld(uint32(wizGamePool.Build()), 1, worldLoc.Render()); err != nil {
		wizQuit(err)
	}
}

func wizSetupWorldServerStandalone() {
	var finger, authAddress string

	for {
		fmt.Println("Enter the address to your auth server.")
		var authAddress string

		if err := survey.AskOne(&survey.Input{
			Message: "Enter the address:port to your Gophercraft Auth server.",
		}, &authAddress); err != nil {
			wizQuit(err)
		}

		conn, err := tls.Dial("tcp", authAddress, &tls.Config{
			InsecureSkipVerify: true,
		})
		if err != nil {
			fmt.Println(err)
			continue
		}

		cert := conn.ConnectionState().PeerCertificates[0]
		finger, err = sys.GetCertFingerprint(cert)
		if err != nil {
			wizQuit(err)
		}

		fmt.Println(authAddress, "fingerprint is", finger)

		var correct bool

		if err := survey.AskOne(&survey.Confirm{
			Message: "Is this correct?",
		}, &correct); err != nil {
			wizQuit(err)
		}

		if correct {
			break
		}
	}

	worldLoc := getLocal().Concat("gcraft_world_1")
	wizWorldPath = worldLoc.Render()
	wizWorldID = 1

	if err := config.GenerateDefaultWorld(uint32(wizGamePool.Build()), wizWorldID, wizWorldPath, authAddress, finger); err != nil {
		wizQuit(err)
	}
}

func wizDatapack() {
	if wizWorldPath == "" {
		wizWorldPath = getLocal().Concat("gcraft_world_1").Render()
	}

	generateDatapack(etc.ParseSystemPath(wizWorldPath).Concat("datapacks", "!base.zip").Render())
}
