package bnet

import "testing"

func TestClient(t *testing.T) {
	cn, err := Dial("logon.scarletwow.eu:1119")
	if err != nil {
		t.Fatal(err)
	}

	t.Fatal(cn.Connect())
}
