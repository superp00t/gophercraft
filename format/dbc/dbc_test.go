package dbc

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/superp00t/etc"
)

func TestTagParse(t *testing.T) {
	tagTest(t, "(loc),5875-12340(only,len:35)", tag{
		rulesets: []ruleset{
			{
				[]int64{},
				[]tagOpt{
					{Type: locOpt},
				},
			},
			{
				[]int64{5875, 12340},
				[]tagOpt{
					{Type: onlyOpt},
					{Type: lengthOpt, Len: 35},
				},
			},
		},
	})

	// range selector
	tagTest(t, ",5875-(len:50),-5875(len:48)", tag{
		rulesets: []ruleset{
			{
				[]int64{5875, -1},
				[]tagOpt{
					{Type: lengthOpt, Len: 50},
				},
			},

			{
				[]int64{-1, 5875},
				[]tagOpt{
					{Type: lengthOpt, Len: 48},
				},
			},
		},
	})

	tagTest(t, "(loc)", tag{
		rulesets: []ruleset{
			{
				[]int64{},
				[]tagOpt{
					{Type: locOpt},
				},
			},
		},
	})
}

func tagTest(t *testing.T, str string, shouldBe tag) {
	if tg := parseTag(str); !reflect.DeepEqual(tg, shouldBe) {
		fmt.Println("Error parsing", str)
		fmt.Println("Should be: ", spew.Sdump(shouldBe))
		fmt.Println("Is: ", spew.Sdump(tg))
		t.Fatal("invalid tag parsing:", str)
	}
}

func TestFiles(t *testing.T) {
	testFiles := etc.Import("github.com/superp00t/gophercraft/format/dbc/testfiles")
	fmt.Printf("'%s'\n", testFiles.Render())

	f, err := testFiles.Concat("12340", "BarberShopStyle.dbc").ReadAll()
	if err != nil {
		t.Fatal(err)
	}

	d, err := Parse(0, f)
	if err != nil {
		t.Fatal(err)
	}

	var barb []Ent_BarberShopStyle

	err = d.ParseRecords(&barb)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(spew.Sdump(barb[:10]))
}
