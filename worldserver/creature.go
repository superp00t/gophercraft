package worldserver

import (
	"github.com/superp00t/gophercraft/guid"
	"github.com/superp00t/gophercraft/packet/update"
)

func (s *Session) MaxPositiveAuras() int {
	return 32
}

type Creature struct {
	ID string
	*update.ValuesBlock
}

func (c *Creature) GUID() guid.GUID {
	if c == nil {
		return guid.Nil
	}
	return c.ValuesBlock.GetGUID("GUID")
}

func (c *Creature) DisplayID() uint32 {
	if c == nil {
		return 0
	}
	return c.ValuesBlock.GetUint32("DisplayID")
}

func (c *Creature) Entry() uint32 {
	if c == nil {
		return 0
	}
	return c.ValuesBlock.GetUint32("Entry")
}

func (c *Creature) GetPowerType() uint8 {
	if c == nil {
		return 0
	}

	return c.GetByte("Power")
}

func (c *Creature) Power() uint32 {
	if c == nil {
		return 0
	}

	switch c.GetPowerType() {
	case Mana:
		return c.GetUint32("Mana")
	case Rage:
		return c.GetUint32("Rage")
	case Focus:
		return c.GetUint32("Focus")
	case Energy:
		return c.GetUint32("Energy")
	}

	panic(c.GetPowerType())
}

func (c *Creature) MaxPower() uint32 {
	switch c.GetPowerType() {
	case Mana:
		return c.GetUint32("MaxMana")
	case Rage:
		return c.GetUint32("MaxRage")
	case Focus:
		return c.GetUint32("MaxFocus")
	case Energy:
		return c.GetUint32("MaxEnergy")
	}

	panic(c.GetPowerType())
}

func (c *Creature) Health() uint32 {
	return c.GetUint32("Health")
}

func (c *Creature) MaxHealth() uint32 {
	return c.GetUint32("MaxHealth")
}
