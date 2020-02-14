package chat

import (
	"fmt"
	"strings"

	"github.com/superp00t/etc"
)

type MarkupType uint8

const (
	Color     MarkupType = 'c'
	Hyperlink MarkupType = 'H'
)

type MarkupData interface {
	String() string
}

type HyperlinkData struct {
	Type   string
	Fields []string
}

func (h HyperlinkData) String() string {
	s := append([]string{h.Type}, h.Fields...)
	return strings.Join(s, ":")
}

type Markup struct {
	Type MarkupType
	Text string
	Data MarkupData
}

func (m *Markup) String() string {
	header := "|" + string(m.Type)

	if m.Type == Color {
		return header + m.Text + m.Data.String() + "|r"
	}

	if m.Type == Hyperlink {
		return header + m.Data.String() + "|h" + m.Text + "|h"
	}

	panic(string(m.Type))
	return ""
}

func ParseMarkup(s string) (*Markup, error) {
	in := etc.FromString(s)
	return parseMarkup(in)
}

func parseMarkup(in *etc.Buffer) (*Markup, error) {
	header := in.ReadByte()

	if header != '|' {
		return nil, fmt.Errorf("%c at %d not a markup string", header, in.Rpos())
	}

	mk := &Markup{}
	t := in.ReadByte()
	mk.Type = MarkupType(t)

	switch mk.Type {
	case Color:
		mk.Text = in.ReadFixedString(8)
		var err error
		mk.Data, err = parseMarkup(in)
		if err != nil {
			return nil, err
		}
	case Hyperlink:
		var err error
		data, err := in.ReadUntilToken("|h")
		if err != nil {
			return nil, err
		}

		strs := strings.Split(data, ":")
		if len(strs) < 2 {
			return nil, fmt.Errorf("bad hyperlink")
		}

		mk.Data = HyperlinkData{
			Type:   strs[0],
			Fields: strs[1:],
		}

		mk.Text, err = in.ReadUntilToken("|h")
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unknown string")
	}

	return mk, nil
}

func (m *Markup) ExtractHyperlinkData() (HyperlinkData, error) {
	var hlink HyperlinkData
	if m.Type != Hyperlink && m.Type != Color {
		return hlink, fmt.Errorf("invalid link type")
	}

	if m.Type == Color {
		mkup := m.Data.(*Markup)
		hl, ok := mkup.Data.(HyperlinkData)
		if !ok {
			return hlink, fmt.Errorf("malformed link")
		}

		hlink = hl
	} else {
		hlink = m.Data.(HyperlinkData)
	}
	return hlink, nil
}
