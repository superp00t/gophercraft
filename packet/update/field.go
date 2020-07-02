package update

import (
	"reflect"
	"strings"
)

type FieldTag string

func (ft FieldTag) explode() []string {
	return strings.Split(string(ft), ",")
}

func (ft FieldTag) IsPrivate() bool {
	for _, t := range ft.explode() {
		if t == "private" {
			return true
		}
	}

	return false
}

func (ft FieldTag) IsParty() bool {
	for _, t := range ft.explode() {
		if t == "party" {
			return true
		}
	}

	return false
}

type FieldRef struct {
	ChunkOffset uint32
	reflect.Value
}
