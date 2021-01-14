package realm

import (
	"github.com/superp00t/gophercraft/realm/wdb"
)

func (s *Session) GetLoc(str string) string {
	var loc *wdb.LocString
	s.DB().GetData(str, &loc)
	if loc == nil {
		return str
	}

	return loc.Text.GetLocalized(s.Locale)
}
