package config

import (
	"os"
	"testing"
)

func TestGenerate(t *testing.T) {
	pth := "/tmp/gcauth"

	os.RemoveAll(pth)

	if err := GenerateDefaultAuth(pth); err != nil {
		t.Fatal(err)
	}
}
