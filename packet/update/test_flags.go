package update

import (
	"fmt"
	"testing"
)

func TestFlags(t *testing.T) {
	flg := GOInUse | GOTriggered | GOTransport

	fmt.Println(flg.String())

	flg2, err := ParseGameObjectFlags(flg.String())
	if err != nil {
		t.Fatal(err)
	}

	if flg != flg2 {
		t.Fatal(flg, "!==", flg2)
	}
}
