package realm

import (
	"math"

	"github.com/superp00t/gophercraft/econ"
)

var (
	MaxGoldNew, _ = econ.ParseShortString("9999999g99s99c")
)

func (s *Session) MaxMoney() econ.Money {
	if s.Build().RemovedIn(13164) {
		return econ.Money(math.MaxInt32)
	}

	return MaxGoldNew
}

func (s *Session) AddMoney(money econ.Money) {
	s.Char.Coinage += money
	if s.Char.Coinage < 0 {
		s.Char.Coinage = 0
	}

	// Before the money value was 64-bit.
	var displayMoney econ.Money
	maxMoney := s.MaxMoney()
	displayMoney = maxMoney

	if s.Char.Coinage > maxMoney {
		s.Warnf("Coinage has overflowed! Your actual money is %s", s.Char.Coinage)
		displayMoney = maxMoney
	}

	s.SetInt32("Coinage", int32(displayMoney))
	s.DB().Where("id = ?", s.PlayerID()).Cols("coinage").Update(s.Char)
	s.UpdateSelf()
}
