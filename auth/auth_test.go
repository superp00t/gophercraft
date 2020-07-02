package auth

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/superp00t/gophercraft/packet"
	"github.com/superp00t/gophercraft/srp"
)

func TestAuth(t *testing.T) {
	const pkt = `01429d3dbafc13bcd95daa3020f7860d947ab6788413c277284ee74da342d504372acfe08e2d92470c4279dd5b479229700a3303fb738cbd9fad90defa934d110ace3aacd346a4a51e0000`

	pktBytes, _ := hex.DecodeString(pkt)

	data, err := packet.UnmarshalAuthLogonProof_C(pktBytes)
	if err != nil {
		panic(err)
	}

	fmt.Println(spew.Sdump(data))
	fmt.Println("a is", srp.BigNumFromArray(data.A).String())
}
