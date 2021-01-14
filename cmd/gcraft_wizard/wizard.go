package main

import (
	"archive/zip"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"regexp"

	"github.com/AlecAivazis/survey"
	"github.com/fatih/color"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/superp00t/etc"
	"github.com/superp00t/gophercraft/format/content"
	"github.com/superp00t/gophercraft/gcore/config"
	"github.com/superp00t/gophercraft/gcore/dbsupport"
	"github.com/superp00t/gophercraft/gcore/sys"
	"google.golang.org/grpc"
)

var (
	wiz          = ""
	wizGamePool  content.Volume
	wizWorldID   uint64
	wizWorldPath string
	wizWorldName string
	wizPack      *zip.Writer
	wizDBDriver  string
	wizDBURL     string
	wizUser      string
	wizPassword  string
)

func wizWarn(args ...interface{}) {
	color.Set(color.FgYellow)
	fmt.Print(wiz)
	fmt.Println(args...)
	color.Unset()
}

func wizOk(args ...interface{}) {
	fmt.Print(wiz)

	for i, v := range args {
		switch attr := v.(type) {
		case color.Attribute:
			color.Set(attr)
		default:
			if i > 0 {
				fmt.Print(" ")
			}

			fmt.Print(v)
		}
	}

	fmt.Println()
	color.Unset()
}

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

func wizValidate(str string) bool {
	ok, _ := regexp.MatchString("^[[:alnum:]]{1,16}$", str)
	return ok
}

func wizQuit(err error) {
	if err == nil {
		fmt.Println("exiting...")
		os.Exit(0)
	}
	fmt.Println(err)
	os.Exit(0)
}

func wizGetLogin() {
	for {
		if err := survey.AskOne(&survey.Input{
			Message: "Enter your username (admin)",
		}, &wizUser); err != nil {
			wizQuit(err)
		}

		if !wizValidate(wizUser) {
			wizWarn("Invalid username. (max length 16 characters, alphanumeric)")
			continue
		}
		break
	}

	for {
		if err := survey.AskOne(&survey.Password{
			Message: "Enter your password.",
		}, &wizPassword); err != nil {
			wizQuit(err)
		}

		if !wizValidate(wizPassword) {
			wizWarn("Invalid password. (max length 16 characters, alphanumeric)")
			continue
		}

		break
	}
}

func wizGetGame() {
	if wizGameDir == "" {
		if err := survey.AskOne(&survey.Input{
			Message: "Input your game directory to continue.",
		},
			&wizGameDir); err != nil {
			wizQuit(err)
		}

		if wizGameDir == "" {
			wizQuit(errors.New("Game files are needed to continue the wizard."))
		}
	}

	var err error
	wizGamePool, err = content.Open(wizGameDir)
	if err != nil {
		wizQuit(err)
	}
}

func wizGetServer() {
	gcDir := etc.LocalDirectory().Concat("Gophercraft")

	listServers, err := ioutil.ReadDir(gcDir.Render())
	if err != nil {
		wizQuit(err)
	}

	if len(listServers) == 0 {
		return
	}

	opts := []string{
		"Quit",
		"New server configuration",
	}

	for _, srv := range listServers {
		if gcDir.Concat(srv.Name()).Exists("World.txt") {
			opts = append(opts, srv.Name())
		}
	}

	qs := &survey.Select{
		Message: "Modify an existing installation?",
		Options: opts,
	}

	var opt int
	err = survey.AskOne(qs, &opt)
	if err != nil {
		wizQuit(err)
	}

	switch {
	case opt == 0:
		wizQuit(nil)
	case opt == 1:
		return
	case opt > 1:
		wizWorldName = opts[opt]
		wizWorldPath = gcDir.Concat(wizWorldName).Render()
	}
}

func wizCreateDB(name string) {
	for {
		wizOk("Note: Gophercraft may in fact support databases not on this list. Check", color.FgGreen, "https://github.com/superp00t/gophercraft/wiki#databases", color.Reset, "for more info!")

		qs := &survey.Select{
			Message: fmt.Sprintf("Choose a database backend for %s.", name),
			Options: append([]string{
				"Quit",
			}, dbsupport.Supported...),
		}
		err := survey.AskOne(qs, &wizDBDriver)
		if wizDBDriver == "Quit" {
			wizQuit(nil)
		}

		err = survey.AskOne(&survey.Input{
			Message: fmt.Sprintf("Enter your database path. For %s the format is %s", wizDBDriver, dbsupport.PathFormat[wizDBDriver]),
		}, &wizDBURL)
		if err != nil {
			wizQuit(err)
		}

		err = dbsupport.Create(wizDBDriver, wizDBURL)
		if err != nil {
			wizWarn(err)
			continue
		}

		return
	}
}

func wizStart() {
	wizOk("Hello, I'm the", color.FgHiBlue, "Gophercraft Wizard!", color.Reset, "I will be your guide through the magical land of", color.FgGreen, "Gophercraft.", color.Reset)

	etc.LocalDirectory().Concat("Gophercraft").MakeDir()

	wizGetGame()
	wizGetServer()

	const (
		justQuit   = "Quit"
		packOnly   = "Continue without setting up server"
		complete   = "Setup Gophercraft Auth and Worldserver on this computer"
		localWorld = "Setup a new Worldserver to use with a local Authserver"
		worldOnly  = "Setup Worldserver on this computer using remote Auth Server"
	)

	qs := &survey.Select{
		Message: "Choose a deployment configuration.",
		Options: []string{
			justQuit,
			packOnly,
			complete,
			localWorld,
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
		wizSetupLocal(true)
		wizDatapack()
	case 3:
		wizSetupLocal(false)
		wizDatapack()
	case 4:
		wizSetupWorldServerStandalone()
		wizDatapack()
	}
}

func wizConfirmOrNew(message, newMessage, confirmString string) string {
	var correct bool

	wizOk(confirmString)

	if err := survey.AskOne(&survey.Confirm{
		Message: message,
	}, &correct); err != nil {
		wizQuit(err)
	}

	if correct {
		return confirmString
	}

	var input string
	err := survey.AskOne(&survey.Input{
		Message: fmt.Sprint(newMessage),
	}, &input)
	if err != nil {
		wizQuit(err)
	}

	return input
}

func wizAsk(args ...interface{}) bool {
	correct := true

	if err := survey.AskOne(&survey.Confirm{
		Message: fmt.Sprintln(args...),
		Default: true,
	}, &correct); err != nil {
		wizQuit(err)
	}

	return correct
}

func wizSetupLocal(setupAuth bool) {
	var line string
	authLoc := getFolder().Concat("Auth")
	if setupAuth == false && !authLoc.IsExtant() || setupAuth {
		line = wizConfirmOrNew("Is this where the auth server should be installed?", "Set authserver location.", authLoc.Render())
	} else {
		line = authLoc.Render()
	}

	// err := survey.AskOne(&survey.Input{
	// 	Message: fmt.Sprint("Set auth core location? if left empty, I will use default: ", authLoc.Render()),
	// }, &line)
	// if err != nil {
	// 	wizQuit(err)
	// }

	// if line == "" {
	// 	line = authLoc.Render()
	// }

	authLoc = etc.ParseSystemPath(line)
	if !authLoc.IsExtant() {
		wizCreateDB("the auth server")

		wizOk("Create an admin account for the auth server.")

		wizGetLogin()

		if err := config.GenerateDefaultAuth(wizDBDriver, wizDBURL, wizUser, wizPassword, authLoc.Render()); err != nil {
			wizQuit(err)
		}
	} else {
		if setupAuth {
			wizQuit(fmt.Errorf("path %s already exists.", authLoc.Render()))
		}
	}

	authconfig, err := config.LoadAuth(authLoc.Render())
	if err != nil {
		wizQuit(err)
	}

	realm, err := authconfig.RealmsFile()
	if err != nil {
		panic(err)
	}

	var last uint64 = 1
	for k := range realm.Realms {
		if k >= last {
			last = k + 1
		}
	}

	if wizWorldName == "" {
		err = survey.AskOne(&survey.Input{
			Message: "What would you like to name your new server?",
		}, &wizWorldName)
		if err != nil {
			wizQuit(err)
		}

		if wizWorldName == "Auth" || wizWorldName == "" {
			wizQuit(fmt.Errorf("You can't name your worldserver that"))
		}
	}

	wizCreateDB(wizWorldName)

	wizWorldID = last

	worldLoc := getFolder().Concat(wizWorldName)
	wizWorldPath = worldLoc.Render()

	if err := authconfig.GenerateDefaultWorld(uint32(wizGamePool.Build()), wizWorldName, wizWorldID, wizDBDriver, wizDBURL, worldLoc.Render()); err != nil {
		wizQuit(err)
	}
}

func wizSetupWorldServerStandalone() {
	var finger, authAddress string

	for {
		fmt.Println("Enter the address to your auth server.")

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

		conn.Close()

		if correct {
			break
		}
	}

	// Query the server to get the next available realm ID
	gc, err := grpc.Dial(
		authAddress,
		grpc.WithInsecure(),
	)
	if err != nil {
		panic(err)
	}

	cl := sys.NewAuthServiceClient(gc)

	for {
		wizOk("You need to enter your admin credentials to register a new world server.")

		wizGetLogin()

		creds, err := cl.CheckCredentials(context.Background(), &sys.Credentials{
			Account:  wizUser,
			Password: wizPassword,
		})
		if err != nil {
			wizQuit(err)
		}

		if creds.Status != sys.Status_SysOK {
			wizWarn("Your password is incorrect.")
			continue
		}

		if creds.Tier < sys.Tier_Admin {
			wizWarn("You have to be an admin to register a server remotely.")
		}
	}

	nextRealm, err := cl.GetNextRealmID(context.Background(), &empty.Empty{})
	if err != nil {
		wizQuit(err)
	}

	wizOk("Next available Realm ID slot is", nextRealm.RealmID)

	err = survey.AskOne(&survey.Input{
		Message: "What would you like to name your new server?",
	}, &wizWorldName)
	if err != nil {
		wizQuit(err)
	}

	if wizWorldName == "Auth" || wizWorldName == "" {
		wizQuit(fmt.Errorf("You can't name your worldserver that"))
	}

	wizCreateDB(wizWorldName)

	worldLoc := getFolder().Concat(wizWorldName)
	wizWorldPath = worldLoc.Render()

	if err := config.GenerateDefaultWorld(uint32(wizGamePool.Build()), wizWorldName, nextRealm.RealmID, wizDBDriver, wizDBURL, wizWorldPath, authAddress, finger); err != nil {
		wizQuit(err)
	}

	nextRealm.RealmFingerprint, err = sys.GetCertFileFingerprint(worldLoc.Concat("cert.pem").Render())
	if err != nil {
		wizQuit(err)
	}

	status, err := cl.AddRealmToConfig(context.Background(), nextRealm)
	if err != nil {
		wizQuit(err)
	}

	if status.Status != sys.Status_SysUnauthorized {
		panic("this should not happen: we checked credentials earlier: " + status.Status.String())
	}

	gc.Close()
}

func wizDatapack() {
	if wizWorldPath == "" {
		generateDatapack(etc.LocalDirectory().Concat(fmt.Sprintf("!base %s.zip", wizGamePool.Build())).Render())
		return
	}

	generateDatapack(etc.ParseSystemPath(wizWorldPath).Concat("datapacks", "!base.zip").Render())
}
