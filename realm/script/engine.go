package script

import (
	"fmt"
	"io"
	"reflect"
	"sync"

	"github.com/superp00t/etc/yo"

	lua "github.com/yuin/gopher-lua"
)

type Engine struct {
	state      *lua.LState
	guardTypes sync.Mutex
	types      map[reflect.Type]*lua.LTable
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
	return &Engine{
		state: L,
		types: make(map[reflect.Type]*lua.LTable),
	}
}

func (e *Engine) Close() {
	e.state.Close()
}

func (e *Engine) SetEnum(name string, value int64) {
	e.state.SetGlobal(name, lua.LNumber(value))
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

func (e *Engine) Invoke(fn *lua.LFunction, args []interface{}, outArg reflect.Value) error {
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
				e.guardTypes.Lock()

				argType, ok := e.types[reflect.TypeOf(v)]
				if !ok {
					return fmt.Errorf("no wrapped interface for %s", reflect.TypeOf(v))
				}
				e.guardTypes.Unlock()

				ud := e.state.NewUserData()
				ud.Value = v
				e.state.SetMetatable(ud, argType)

				e.state.Push(ud)
			}
		}
	}

	fmt.Println("PCall")

	err := e.state.PCall(len(args), 1, nil)
	if err != nil {
		fmt.Println("PCall er", err)
		return err
	}
	fmt.Println("PCall over")

	if !outArg.IsValid() {
		return nil
	}

	if e.state.Get(-1).Type() == lua.LTNil {
		e.state.Pop(1)
		return nil
	}

	switch outArg.Kind() {
	case reflect.Bool:
		outArg.SetBool(e.state.ToBool(-1))
	case reflect.Uint32, reflect.Uint8, reflect.Uint16, reflect.Uint64:
		outArg.SetUint(uint64(e.state.ToNumber(-1)))
	case reflect.Float32, reflect.Float64:
		outArg.SetFloat(float64(e.state.ToNumber(-1)))
	case reflect.String:
		outArg.SetString(e.state.ToString(-1))
	}

	e.state.Pop(1)
	return nil
}

func (e *Engine) getInType(arg reflect.Type, luaOffset int, istate *lua.LState) interface{} {
	var nxtArg interface{}

	var cb *lua.LFunction

	if arg == reflect.TypeOf(cb) {
		return istate.ToFunction(luaOffset)
	}

	switch arg.Kind() {
	case reflect.Ptr, reflect.Struct, reflect.Interface:
		uData := istate.ToUserData(luaOffset)
		nxtArg = uData.Value
	case reflect.String:
		nxtArg = istate.ToString(luaOffset)
	case reflect.Int:
		nxtArg = istate.ToInt(luaOffset)
	case reflect.Uint8:
		nxtArg = uint8(istate.ToInt(luaOffset))
	case reflect.Uint16:
		nxtArg = uint16(istate.ToInt(luaOffset))
	case reflect.Uint32:
		nxtArg = uint32(istate.ToInt(luaOffset))
	case reflect.Uint64:
		nxtArg = uint64(istate.ToInt64(luaOffset))
	case reflect.Int16:
		nxtArg = int16(istate.ToInt(luaOffset))
	case reflect.Int32:
		nxtArg = int32(istate.ToInt(luaOffset))
	case reflect.Int64:
		nxtArg = istate.ToInt64(luaOffset)
	case reflect.Float32:
		nxtArg = float32(istate.ToNumber(luaOffset))
	case reflect.Float64:
		nxtArg = float64(istate.ToNumber(luaOffset))
	case reflect.Bool:
		nxtArg = bool(istate.ToBool(luaOffset))
	default:
		panic("unhandled type " + arg.String())
	}

	return nxtArg
}

func (e *Engine) generateBinding(fn reflect.Value) lua.LGFunction {
	return func(istate *lua.LState) int {
		var args []reflect.Value
		typeOf := fn.Type()

		// if typeOf.NumIn() > 0 {
		// 	if typeOf.In(0).Kind() == reflect.Ptr {
		// 		lval := istate.Get(1)
		// 		if lval.Type() != lua.LTUserData {
		// 			panic("mismatch type")
		// 		}

		// 		lv := lval.(*lua.LUserData)
		// 		yo.Spew(lv.Value)

		// 		yo.Warn("arg 1", lval.String())
		// 	} else {
		// 		yo.Warn("Not ptr", typeOf.In(0))
		// 	}
		// }

		for x := 0; x < typeOf.NumIn(); x++ {
			arg := typeOf.In(x)

			luaOffset := x + 1

			nxtArg := e.getInType(arg, luaOffset, istate)

			args = append(args, reflect.ValueOf(nxtArg))
		}

		out := fn.Call(args)
		for _, v := range out {
			if v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
				if v.IsNil() {
					istate.Push(lua.LNil)
					continue
				}
			}

			if v.Kind() == reflect.Struct || v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
				typeMetatable := e.types[v.Type()]
				value := e.state.NewUserData()
				value.Value = v.Interface()
				if typeMetatable != nil {
					e.state.SetMetatable(value, typeMetatable)
				} else {
					yo.Warn("no metatable for", v.Type())
				}
				istate.Push(value)
				continue
			}

			switch v.Kind() {
			case reflect.String:
				istate.Push(lua.LString(v.String()))
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				istate.Push(lua.LNumber(v.Int()))
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				istate.Push(lua.LNumber(v.Uint()))
			case reflect.Float32, reflect.Float64:
				istate.Push(lua.LNumber(v.Float()))
			case reflect.Bool:
				istate.Push(lua.LBool(v.Bool()))
			default:
				panic("unhandled return type " + v.Type().String())
			}
		}

		return len(out)
	}
}

func (e *Engine) WrapInterface(name string, interfacePointer interface{}, newFunc interface{}, exports []string) {
	interfaceType := reflect.TypeOf(interfacePointer)

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
				nxtArg := e.getInType(nFn.In(x-2), x, L)
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

	fmt.Println("setting type", interfaceType)
	e.guardTypes.Lock()
	e.types[interfaceType] = mt
	e.guardTypes.Unlock()

	var methods = map[string]lua.LGFunction{}

	for _, methodName := range exports {
		method, ok := interfaceType.MethodByName(methodName)
		if !ok {
			panic("not found: " + methodName)
		}

		methods[method.Name] = e.generateBinding(method.Func)
	}

	e.state.SetField(mt, "__index", e.state.SetFuncs(e.state.NewTable(), methods))
}

func (e *Engine) SetCallback(name string, _func interface{}) {
	bound := e.generateBinding(reflect.ValueOf(_func))
	e.state.SetGlobal(name, e.state.NewFunction(bound))
}

func (e *Engine) DoString(file string) error {
	return e.state.DoString(file)
}
