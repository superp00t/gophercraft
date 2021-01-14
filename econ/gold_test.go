package econ

import (
	"fmt"
	"testing"
)

func TestGold(t *testing.T) {
	// var m Money = Money(2147483647)
	// if m.String() != "214748 Gold, 36 Silver, 46 Copper" {
	// 	t.Fatal("wrong encoding", m.String())
	// }

	iString, err := ParseShortString("-100g5c")
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(iString)
}
