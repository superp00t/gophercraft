package script

import (
	"io"

	lua "github.com/yuin/gopher-lua"
)

type Filesystem interface {
	Exists(path string) bool
	Open(path string) (io.ReadCloser, error)
}

type loader struct {
	Filesystem
}

func (ld *loader) loaderPreload(L *lua.LState) int {
	panic("cannot use preload/require")

	// name := L.CheckString(1)
	// preload := L.GetField(L.GetField(L.Get(lua.EnvironIndex), "package"), "preload")
	// if _, ok := preload.(*lua.LTable); !ok {
	// 	L.RaiseError("package.preload must be a table")
	// }
	// lv := L.GetField(preload, name)
	// if lv == lua.LNil {
	// 	L.Push(lua.LString(fmt.Sprintf("no field package.preload['%s']", name)))
	// 	return 1
	// }
	// L.Push(lv)
	return 1
}

// func (ld *loader) getCurrentFolder(L *lua.LState) string {
// 	pth := string(L.GetGlobal("package").(*lua.LTable).RawGetString("path").(lua.LString))
// 	els := strings.Split(pth, "/")
// 	fldr := strings.Join(els[:len(els)-1], "/") + "/"
// 	return fldr
// }

func (ld *loader) loaderLua(L *lua.LState) int {
	panic("cannot use preload/require")

	// name := L.CheckString(1)
	// path := name + ".lua"
	// if ld.Exists(path) == false {
	// 	pPath := ld.getCurrentFolder(L) + path
	// 	path = pPath
	// }

	// file, err := ld.Open(path)
	// if err != nil {
	// 	panic(err)
	// }

	// pkg := L.GetGlobal("package").(*lua.LTable)
	// // originalPath := pkg.RawGetString("path").(lua.LString)
	// pkg.RawSetString("path", lua.LString(path))

	// fn, err := L.Load(file, name)
	// if err != nil {
	// 	panic(err)
	// }

	// L.Push(fn)
	// file.Close()
	return 1
}

func OpenPackage(L *lua.LState) int {
	ld := &loader{nil}

	packagemod := L.RegisterModule(lua.LoadLibName, loFuncs)

	L.SetField(packagemod, "preload", L.NewTable())

	loaders := L.CreateTable(2, 0)
	L.RawSetInt(loaders, 1, L.NewFunction(ld.loaderPreload))
	L.RawSetInt(loaders, 2, L.NewFunction(ld.loaderLua))

	L.SetField(packagemod, "loaders", loaders)
	L.SetField(L.Get(lua.RegistryIndex), "_LOADERS", loaders)

	loaded := L.NewTable()
	L.SetField(packagemod, "loaded", loaded)
	L.SetField(L.Get(lua.RegistryIndex), "_LOADED", loaded)

	L.SetField(packagemod, "path", lua.LString("/"))
	L.SetField(packagemod, "cpath", lua.LString(""))

	L.Push(packagemod)
	return 1
}

var loFuncs = map[string]lua.LGFunction{
	"loadlib": loLoadLib,
	"seeall":  loSeeAll,
}

func loLoadLib(L *lua.LState) int {
	L.RaiseError("loadlib is not supported")
	return 0
}

func loSeeAll(L *lua.LState) int {
	mod := L.CheckTable(1)
	mt := L.GetMetatable(mod)
	if mt == lua.LNil {
		mt = L.CreateTable(0, 1)
		L.SetMetatable(mod, mt)
	}
	L.SetField(mt, "__index", L.Get(lua.GlobalsIndex))
	return 0
}
