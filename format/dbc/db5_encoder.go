package dbc

import (
	"github.com/superp00t/etc"
)

func EncodeDB5(e *etc.Buffer, locale uint32, v interface{}) error {
	e.WriteFixedString(4, "WDB5")
	return nil
}
