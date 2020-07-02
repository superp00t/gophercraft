package text

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/davecgh/go-spew/spew"
)

func TestDecode(t *testing.T) {
	type testStruct struct {
		Strfield   string
		IntField   uint32
		FloatField float64
		SliceField []float32
		Fielded    struct {
			Data string
		}
		Dict      Dict
		FloatDict map[float32]struct {
			Test string
		}
	}

	code := `{
		Strfield 1
		FloatField 3.5
	}
	{ 
		Strfield "hello\t\nworld" // Multiline strings are allowed, however you can use escape sequences
		FloatField 19.17
		SliceField
		{
			123
			596
			2313
			3414.4
		}
	}{
		Strfield "hey friends
			how's it going
		"
	}
	// Fields can also be omitted entirely
	{ // quotes in weird places
	}
	// Structs can be one-liners if you want them to
	{ Strfield hi FloatField "3.6"  }
	// Put fields into structs without fully bracketing.
	{
		Fielded.Data "Test"
	}
	// Dictionary
	{
		Dict
		{
			"Something" "else"
			EvenHave.Periods too
		}

		FloatDict
		{
			19.16 {
				Test "test"
			}
			20.17 {
				Test words50
			}
		}
	}
	`

	var test testStruct

	reader := strings.NewReader(code)
	decoder := NewDecoder(reader)

	writer := bytes.NewBuffer(nil)
	out := NewEncoder(writer)
	out.Indent = "  "

	for {
		err := decoder.Decode(&test)
		if err == io.EOF {
			break
		}

		if err != nil {
			t.Fatal(err)
		}

		fmt.Println("Strfield", test.Strfield)

		t.Log(spew.Sdump(test))

		if err := out.Encode(test); err != nil {
			t.Fatal(err)
		}

		t.Log(spew.Sdump(test.Dict))
	}

	fmt.Println(writer.String())
}
