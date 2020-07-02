package worldserver

import (
	"reflect"
	"strconv"

	"github.com/superp00t/gophercraft/guid"
)

func x_VMod(c *C) {
	if len(c.Args) < 1 {
		c.Session.Warnf(".vmod ClassName 3.14")
		return
	}

	var args []interface{}

	for x := 0; x < len(c.Args)-1; x++ {
		arg := c.Args[x]
		idx, err := strconv.ParseInt(arg, 0, 64)
		if err != nil {
			args = append(args, arg)
		} else {
			args = append(args, int(idx))
		}
	}

	newValue := c.Args[len(c.Args)-1]

	off, val, err := c.Session.FindValueOffset(args...)
	if err != nil {
		c.Session.Warnf("%s", err)
		return
	}

	c.Session.Warnf("Value found: 0x%04X (%d) %s", off, off, val.Type())

	if val.Type() == reflect.TypeOf(guid.GUID{}) {
		id, err := guid.FromString(newValue)
		if err != nil {
			c.Session.Warnf("%s", err)
			return
		}
		val.Set(reflect.ValueOf(id))
		c.Session.ValuesBlock.ChangeMask.Set(off, true)
		c.Session.ValuesBlock.ChangeMask.Set(off+1, true)
		return
	}

	switch val.Kind() {
	case reflect.Uint8, reflect.Uint32:
		u64, err := strconv.ParseUint(newValue, 0, 32)
		if err != nil {
			c.Session.Warnf("%s", err)
			return
		}
		val.SetUint(u64)
		c.Session.ValuesBlock.ChangeMask.Set(off, true)
	case reflect.Int32:
		i64, err := strconv.ParseInt(newValue, 0, 32)
		if err != nil {
			c.Session.Warnf("%s", err)
			return
		}
		val.SetInt(i64)
		c.Session.ValuesBlock.ChangeMask.Set(off, true)
	case reflect.Bool:
		tru := newValue == "true"
		val.SetBool(tru)
		c.Session.ValuesBlock.ChangeMask.Set(off, true)
	default:
		c.Session.Warnf("unknown kind %s", val.Kind())
		return
	}
	c.Session.Map().PropagateChanges(c.Session.GUID())
}
