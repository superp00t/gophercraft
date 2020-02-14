package script

import (
	"fmt"
	"io"
	"reflect"

	lua "github.com/yuin/gopher-lua"
)

type Engine struct {
	state *lua.LState
}

func NewEngine() *Engine {
	L := lua.NewState(lua.Options{SkipOpenLibs: true})
	for _, pair := range []struct {
		n string
		f lua.LGFunction
	}{
		{lua.LoadLibName, OpenPackage}, // Must be first
		{lua.BaseLibName, lua.OpenBase},
		{lua.TabLibName, lua.OpenTable},
	} {
		if err := L.CallByParam(lua.P{
			Fn:      L.NewFunction(pair.f),
			NRet:    0,
			Protect: true,
		}, lua.LString(pair.n)); err != nil {
			panic(err)
		}
	}
	return &Engine{L}
}

func (e *Engine) Close() {
	e.state.Close()
}

func (e *Engine) DoReader(path string, rdr io.Reader) error {
	e.state.GetGlobal("package").(*lua.LTable).RawSetString("path", lua.LString(path))
	lf, err := e.state.Load(rdr, path)
	if err != nil {
		return err
	}
	e.state.Push(lf)
	return e.state.PCall(0, lua.MultRet, nil)
}

func (e *Engine) GetUserData(name string, value interface{}) *lua.LUserData {
	ud := e.state.NewUserData()
	ud.Value = value
	e.state.SetMetatable(ud, e.state.GetTypeMetatable(name))
	return ud
}

func (e *Engine) Invoke(fn *lua.LFunction, args []interface{}) error {
	e.state.Push(fn)

	for _, v := range args {
		switch arg := v.(type) {
		case uint32:
			e.state.Push(lua.LNumber(arg))
		case uint64:
			e.state.Push(lua.LNumber(arg))
		case int:
			e.state.Push(lua.LNumber(arg))
		case int64:
			e.state.Push(lua.LNumber(arg))
		case float32:
			e.state.Push(lua.LNumber(arg))
		case float64:
			e.state.Push(lua.LNumber(arg))
		default:
			if in, ok := v.(lua.LValue); ok {
				e.state.Push(in)
			} else {
				return fmt.Errorf("invalid type %s", reflect.TypeOf(v))
			}
		}
	}

	return e.state.PCall(len(args), lua.MultRet, nil)
}

func getInType(arg reflect.Type, luaOffset int, istate *lua.LState) interface{} {
	var nxtArg interface{}

	var cb *lua.LFunction

	if arg == reflect.TypeOf(cb) {
		return istate.ToFunction(luaOffset)
	}

	switch arg.Kind() {
	case reflect.Ptr, reflect.Struct:
		uData := istate.ToUserData(luaOffset)
		nxtArg = uData.Value
	case reflect.String:
		nxtArg = istate.ToString(luaOffset)
	case reflect.Int:
		nxtArg = istate.ToInt(luaOffset)
	case reflect.Uint32:
		nxtArg = uint32(istate.ToInt(luaOffset))
	case reflect.Uint64:
		nxtArg = uint64(istate.ToInt64(luaOffset))
	case reflect.Int64:
		nxtArg = istate.ToInt64(luaOffset)
	case reflect.Float32:
		nxtArg = float32(istate.ToNumber(luaOffset))
	case reflect.Float64:
		nxtArg = float64(istate.ToNumber(luaOffset))
	default:
		panic("unhandled type " + arg.String())
	}

	return nxtArg
}

func (e *Engine) generateBinding(fn interface{}) lua.LGFunction {
	return func(istate *lua.LState) int {
		var args []reflect.Value
		typeOf := reflect.TypeOf(fn)
		for x := 0; x < typeOf.NumIn(); x++ {
			arg := typeOf.In(x)

			luaOffset := x + 1

			nxtArg := getInType(arg, luaOffset, istate)

			args = append(args, reflect.ValueOf(nxtArg))
		}

		out := reflect.ValueOf(fn).Call(args)
		for _, v := range out {
			switch v.Kind() {
			case reflect.String:
				istate.Push(lua.LString(v.String()))
			case reflect.Int, reflect.Int64:
				istate.Push(lua.LNumber(v.Int()))
			case reflect.Float32, reflect.Float64:
				istate.Push(lua.LNumber(v.Float()))
			default:
				panic(v.Type())
			}
		}

		return len(out)
	}
}

func (e *Engine) WrapInterface(name string, face, newFunc interface{}) {
	mt := e.state.NewTypeMetatable(name)
	e.state.SetGlobal(name, mt)
	if newFunc != nil {
		// set constructor function
		vFn := reflect.ValueOf(newFunc)
		nFn := vFn.Type()
		if nFn.NumOut() < 1 || nFn.NumOut() > 2 {
			panic("invalid number of return arguments in " + name + " constructor")
		}

		fnc := func(L *lua.LState) int {
			var args []reflect.Value

			for x := 2; x < nFn.NumIn()+2; x++ {
				nxtArg := getInType(nFn.In(x-2), x, L)
				args = append(args, reflect.ValueOf(nxtArg))
			}

			outArgs := vFn.Call(args)
			constructed := outArgs[0]
			// check for error
			if len(outArgs) > 1 {
				err := outArgs[1]
				if err.Interface() != nil {
					panic(err.Interface().(error))
				}
			}

			ud := L.NewUserData()
			ud.Value = constructed.Interface()
			L.SetMetatable(ud, L.GetTypeMetatable(name))
			L.Push(ud)
			return 1
		}

		e.state.SetField(mt, "new", e.state.NewFunction(fnc))
	}

	tOf := reflect.TypeOf(face)

	var methods = map[string]lua.LGFunction{}
	for x := 0; x < tOf.NumMethod(); x++ {
		method := tOf.Method(x)
		methods[method.Name] = e.generateBinding(method.Func.Interface())
	}

	e.state.SetField(mt, "__index", e.state.SetFuncs(e.state.NewTable(), methods))
}

func (e *Engine) SetCallback(name string, _func interface{}) {
	bound := e.generateBinding(_func)
	e.state.SetGlobal(name, e.state.NewFunction(bound))
}

func (e *Engine) DoString(file string) error {
	return e.state.DoString(file)
}
