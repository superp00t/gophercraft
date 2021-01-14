package dbc

import (
	"fmt"
	"strconv"

	"github.com/superp00t/etc"
	"github.com/superp00t/gophercraft/vsn"
)

type tagOptType int

const (
	noOpt tagOptType = iota
	lengthOpt
	onlyOpt
	disabledOpt
	locOpt
)

type tag struct {
	rulesets []ruleset
}

type ruleset struct {
	Versions []int64
	Rules    []tagOpt
}

type tagOpt struct {
	Type tagOptType
	Len  int64
}

func parseInt(e *etc.Buffer) int64 {
	str := ""

	for {
		rn, _, _ := e.ReadRune()

		switch rn {
		case '1', '2', '3', '4', '5', '6', '7', '8', '9', '0':
			str += string(rn)
		default:
			e.Jump(-1)
			i, err := strconv.ParseInt(str, 0, 46)
			if err != nil {
				return -1
			}
			return i
		}
	}
}

func parseRange(e *etc.Buffer) []int64 {
	rng := []int64{}

	numBuf := etc.NewBuffer()

	two := false

loop:
	for x := 0; ; x++ {
		if e.Available() == 0 {
			break
		}

		r, _, err := e.ReadRune()
		if err != nil {
			break
		}

		switch r {
		case 0:
			break loop
		case '(':
			e.Jump(-1)
			if x == 0 {
				return []int64{}
			}
			break loop
		case '-':
			if two {
				panic("dbc: two - characters are not allowed when parsing range selector.")
			}

			rng = append(rng, parseInt(numBuf))

			numBuf = etc.NewBuffer()
			two = true
		case '1', '2', '3', '4', '5', '6', '7', '8', '9', '0':
			numBuf.WriteRune(r)
		default:
			panic(fmt.Sprintf("%d unexpected character in tag: %s", x, string(r)))
		}
	}

	n2 := int64(-1)

	if numBuf.Len() > 0 {
		n2 = parseInt(numBuf)
	}

	rng = append(rng, n2)

	return rng
}

func typeKey(s string) tagOptType {
	switch s {
	case "only":
		return onlyOpt
	case "loc":
		return locOpt
	case "disabled":
		return disabledOpt
	default:
		panic("unknown type key " + s)
	}
}

func parseTag(s string) tag {
	fo := tag{}

	e := etc.FromString(s)

	for {
		if e.Available() == 0 {
			return fo
		}

		pComma, _, _ := e.ReadRune()
		if pComma == ',' {
			continue
		} else {
			e.Jump(-1)
		}

		rng := parseRange(e)

		rn, _, _ := e.ReadRune()
		if rn != '(' {
			panic("expected ( after range list")
		}

		rset := ruleset{}
		rset.Versions = rng

		tmpKey := ""

	argumentLoop:
		for {
			rn, _, err := e.ReadRune()
			if err != nil {
				panic(err)
			}

			if rn == ':' && tmpKey == "len" {
				rset.Rules = append(rset.Rules, tagOpt{
					Type: lengthOpt,
					Len:  parseInt(e),
				})
				tmpKey = ""
				continue argumentLoop
			}

			if tmpKey != "" && (rn == ',' || rn == ')') {
				rset.Rules = append(rset.Rules, tagOpt{
					Type: typeKey(tmpKey),
				})
				tmpKey = ""
			}

			if rn == ',' {
				continue argumentLoop
			}

			if rn == ')' {
				break argumentLoop
			}

			tmpKey += string(rn)
		}

		fo.rulesets = append(fo.rulesets, rset)
	}

	return fo
}

func versionMatch(build vsn.Build, vsns []int64) bool {
	if len(vsns) == 0 {
		return true
	}

	if len(vsns) == 1 {
		return vsn.Build(vsns[0]) == build
	}

	if len(vsns) == 2 {
		if vsns[0] == -1 {
			if build <= vsn.Build(vsns[1]) {
				return true
			}
		}

		if vsns[1] == -1 {
			if build >= vsn.Build(vsns[0]) {
				return true
			}
		}

		if build >= vsn.Build(vsns[0]) && build <= vsn.Build(vsns[1]) {
			return true
		}
	}

	return false
}

func (f tag) getValidOpts(build vsn.Build) (disabled bool, out []tagOpt) {
	// Look out for disabled fields
	for _, ruleset := range f.rulesets {
		for _, rule := range ruleset.Rules {
			if rule.Type == onlyOpt {
				if !versionMatch(build, ruleset.Versions) {
					return true, nil
				}
			}

			if rule.Type == disabledOpt {
				if versionMatch(build, ruleset.Versions) {
					return true, nil
				}
			}
		}
	}

	for _, ruleset := range f.rulesets {
		for _, rule := range ruleset.Rules {
			if versionMatch(build, ruleset.Versions) {
				out = append(out, rule)
			}
		}
	}

	return false, out
}
