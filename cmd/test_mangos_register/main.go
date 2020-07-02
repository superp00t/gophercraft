package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/superp00t/gophercraft/srp"
)

func main() {
	if len(os.Args) < 3 {
		return
	}

	g := srp.Generator.Copy()
	N := srp.Prime.Copy()

	auth := srp.HashCredentials(os.Args[1], os.Args[2])
	salt := srp.BigNumFromRand(32)

	_, v := srp.CalculateVerifier(auth, g, N, salt)

	fmt.Printf("INSERT INTO `classicrealmd`.`account` (username, v, s) VALUES(\n")
	fmt.Printf("	'%s',\n", strings.ToUpper(os.Args[1]))
	fmt.Printf("	'%s',\n", v.ToHex())
	fmt.Printf("	'%s'\n", salt.ToHex())
	fmt.Println(");")

}
