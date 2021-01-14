package commands

import (
	"reflect"
	"strconv"

	"github.com/superp00t/gophercraft/guid"
	"github.com/superp00t/gophercraft/realm"
)

func cmdVmod(s *realm.Session, params []string) {
	if len(params) < 1 {
		s.Warnf(".vmod ClassName 3.14")
		return
	}

	var args []interface{}

	for x := 0; x < len(params)-1; x++ {
		arg := params[x]
		idx, err := strconv.ParseInt(arg, 0, 64)
		if err != nil {
			args = append(args, arg)
		} else {
			args = append(args, int(idx))
		}
	}

	newValue := params[len(params)-1]

	off, val, err := s.FindValueOffset(args...)
	if err != nil {
		s.Warnf("%s", err)
		return
	}

	s.Warnf("Value found: 0x%04X (%d) %s", off, off, val.Type())

	if val.Type() == reflect.TypeOf(guid.GUID{}) {
		id, err := guid.FromString(newValue)
		if err != nil {
			s.Warnf("%s", err)
			return
		}
		val.Set(reflect.ValueOf(id))
		s.ValuesBlock.ChangeMask.Set(off, true)
		s.ValuesBlock.ChangeMask.Set(off+1, true)
		return
	}

	switch val.Kind() {
	case reflect.Uint8, reflect.Uint32:
		u64, err := strconv.ParseUint(newValue, 0, 32)
		if err != nil {
			s.Warnf("%s", err)
			return
		}
		val.SetUint(u64)
		s.ValuesBlock.ChangeMask.Set(off, true)
	case reflect.Int32:
		i64, err := strconv.ParseInt(newValue, 0, 32)
		if err != nil {
			s.Warnf("%s", err)
			return
		}
		val.SetInt(i64)
		s.ValuesBlock.ChangeMask.Set(off, true)
	case reflect.Bool:
		tru := newValue == "true"
		val.SetBool(tru)
		s.ValuesBlock.ChangeMask.Set(off, true)
	default:
		s.Warnf("unknown kind %s", val.Kind())
		return
	}
	s.Map().PropagateChanges(s.GUID())
}
