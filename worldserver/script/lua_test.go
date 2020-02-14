package script

import (
	"fmt"
	"testing"
	"time"

	"github.com/superp00t/etc"
	"github.com/superp00t/gophercraft/datapack"
)

func timeExe(t *testing.T, e *Engine, process, code string) {
	start := time.Now()
	if err := e.DoString(code); err != nil {
		t.Fatal(err)
	}
	fmt.Println(process, "completed in", time.Since(start))
}

type abstractTestItem interface {
	Debug()
}

type testItem struct {
	Data string
}

func (t *testItem) Debug() {
	fmt.Println("debug: ", t.Data)
}

func (t *testItem) Dothing() {
	fmt.Println("I did a thing!")
}

func TestEngine(t *testing.T) {
	scriptPath := etc.Import("github.com/superp00t/gophercraft/worldserver/script")

	load, err := datapack.Open(scriptPath.Concat("packs").Render())
	if err != nil {
		t.Fatal(err)
	}

	e := NewEngine()

	e.WrapInterface("testItem", abstractTestItem(&testItem{}), func(data string) (abstractTestItem, error) {
		return &testItem{
			data,
		}, nil
	})

	e.SetCallback("PushInfo", func(data int64, name string) {
		fmt.Println(data, name)
	})

	b, err := etc.Import("github.com/superp00t/gophercraft/worldserver/script/testCallbacks.lua").ReadAll()
	if err != nil {
		panic(err)
	}

	for _, v := range load.Volumes {
		for _, sc := range v.ServerScripts {
			testPath := "testpack:Scripts/" + sc
			fl, err := load.Open(testPath)
			if err != nil {
				panic(err)
			}

			if err := e.DoReader(testPath, fl); err != nil {
				panic(err)
			}

			fl.Close()
		}
	}

	timeExe(t, e, "Invoke callbacks", string(b))

	timeExe(t, e, "Test class", `
		ti = testItem:new("hello world")
		ti:Debug()
		-- ti:DoThing() should fail (not part of abstractTestItem interface)
	`)

	timeExe(t, e, "calc fibonacci", `
	function fibonacci(n, x)
	if n<3 then
				return 1
		else
				return fibonacci(n-1, x+1) + fibonacci(n-2, x+1)
		end
end

	print(fibonacci(35, 0))
	`)

	timeExe(t, e, "enum test", `
	local enummt = {
		__index = function(table, key) 
			if rawget(table.enums, key) then 
				return key
			end
		end
	}
	
	local function Enum(t)
		local e = { enums = t }
		return setmetatable(e, enummt)
	end
	
	local Screen = Enum {
		Main = 1,
		TransmogUI = 2
	}
	
	print(Screen.TransmogUI)
	`)
}
