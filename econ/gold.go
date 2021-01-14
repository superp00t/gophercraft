//Package econ provides economy-related structs and functions.
package econ

import (
	"fmt"
	"strconv"
)

// Money represents a coinage state in the in-game economy.
// You cannot have negative money in-game, but you can set money as negative in order to subtract from a balance by addition.
type Money int64

const (
	Copper Money = 1
	Silver Money = 100 * Copper
	Gold   Money = 100 * Silver
)

func (m Money) Int32() int32 {
	// TODO: check for overflow
	return int32(m)
}

func (mn Money) Coins() []int64 {
	m := int64(mn)
	tgold := m / int64(Gold)
	silver := (m - (tgold * int64(Gold))) / int64(Silver)
	copper := (m - (silver * int64(Silver))) - (tgold * int64(Gold))
	return []int64{tgold, silver, copper}
}

func (m Money) String() string {
	c := m.Coins()
	return fmt.Sprintf("%d Gold, %d Silver, %d Copper", c[0], c[1], c[2])
}

func (m Money) ShortString() string {
	c := m.Coins()

	str := ""

	if c[0] != 0 {
		str += fmt.Sprintf("%dg", c[0])
	}

	if c[1] != 0 {
		str += fmt.Sprintf("%ds", c[1])
	}

	if c[2] != 0 {
		str += fmt.Sprintf("%dc", c[2])
	}

	if str == "" {
		return "0c"
	}

	return str
}

func getCoinKey(s string) (string, string) {
	l := len(s)
	key := string(s[l-1])
	val := s[:l-1]
	return key, val
}

func ParseShortString(input string) (Money, error) {
	pRead := input[:]

	if len(pRead) == 0 {
		return 0, fmt.Errorf("econ: Gold string is empty")
	}

	var sign bool
	var money Money

	if pRead[0] == '-' {
		sign = true
		pRead = pRead[1:]
	}

denomination:
	for len(pRead) > 0 {
		var intString string
		for len(pRead) > 0 {
			c := pRead[0]
			pRead = pRead[1:]
			if c >= '0' && c <= '9' {
				intString += string(c)
			} else {
				switch c {
				case 'c':
					i, err := strconv.ParseInt(intString, 10, 64)
					if err != nil {
						return 0, err
					}

					money += Money(i)
					continue denomination
				case 's':
					i, err := strconv.ParseInt(intString, 10, 64)
					if err != nil {
						return 0, err
					}

					money += Money(i) * Silver
					continue denomination
				case 'g':
					i, err := strconv.ParseInt(intString, 10, 64)
					if err != nil {
						return 0, err
					}

					money += Money(i) * Gold
					continue denomination
				default:
					return 0, fmt.Errorf("econ: unknown denomination %s", string(c))
				}
			}
		}
	}

	if sign {
		money *= -1
	}

	return money, nil
}

func ParseMoney(in string) (Money, error) {
	i, err := strconv.ParseInt(in, 0, 64)
	if err != nil {
		m, err := ParseShortString(in)
		if err != nil {
			return 0, err
		}
		return m, nil
	}

	return Money(i), nil
}
