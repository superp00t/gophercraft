package chat

import (
	"fmt"
	"testing"

	"github.com/davecgh/go-spew/spew"
)

func TestMarkup(t *testing.T) {
	str := "|cffffffff|Hitem:6384:0:0:0|h[Stylish Blue Shirt]|h|r"

	spew.Config.DisableMethods = true

	mk, err := ParseMarkup(str)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(spew.Sdump(mk))
}
