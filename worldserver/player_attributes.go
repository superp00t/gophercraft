package worldserver

func (s *Session) Health() uint32 {
	return s.GetUint32("Health")
}

func (s *Session) MaxHealth() uint32 {
	return s.GetUint32("MaxHealth")
}

func (s *Session) GetPowerType() uint8 {
	return s.GetByte("Power")
}

const (
	Mana = iota
	Rage
	Focus
	Energy
	Happiness
)

func (s *Session) Power() uint32 {
	switch s.GetPowerType() {
	case Mana:
		return s.GetUint32("Mana")
	case Rage:
		return s.GetUint32("Rage")
	case Focus:
		return s.GetUint32("Focus")
	case Energy:
		return s.GetUint32("Energy")
	}

	panic(s.GetPowerType())
}

func (s *Session) MaxPower() uint32 {
	switch s.GetPowerType() {
	case Mana:
		return s.GetUint32("MaxMana")
	case Rage:
		return s.GetUint32("MaxRage")
	case Focus:
		return s.GetUint32("MaxFocus")
	case Energy:
		return s.GetUint32("MaxEnergy")
	}

	panic(s.GetPowerType())
}

func (s *Session) VehicleSeatID() uint32 {
	return 0
}

func (s *Session) Pet() *Creature {
	return nil
}
