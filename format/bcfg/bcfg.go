package bcfg

import (
	"strings"

	"github.com/superp00t/etc"
)

type Config struct {
	Comment string
	Data    map[string][]string
}

func (c *Config) Encode() []byte {
	b := ""
	b += "# " + c.Comment + "\n\n"
	for k, v := range c.Data {
		b += k + " = " + strings.Join(v, " ") + "\n"
	}
	return []byte(b)
}

func Parse(input []byte) (*Config, error) {
	e := etc.MkBuffer(input)

	c := new(Config)
	c.Data = make(map[string][]string)
	for {
		ln, _ := e.ReadString('\n')

		if ln == "" {
			break
		}

		if ln[0] == '#' {
			c.Comment = ln[2:]
			e.ReadString('\n')
			continue
		}

		str := strings.SplitN(ln, " = ", 2)
		c.Data[str[0]] = strings.Split(str[1], " ")
	}

	return c, nil
}
