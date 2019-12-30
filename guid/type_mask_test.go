package guid

import "testing"

func TestTypeMask(t *testing.T) {
	tm := TypeMaskObject | TypeMaskUnit | TypeMaskPlayer

	value, err := tm.Resolve(5875)
	if err != nil {
		t.Fatal(err)
	}

	if value != 25 {
		t.Fatal("invalid typemask resolution", value)
	}
}
