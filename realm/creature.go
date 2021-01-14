package realm

import (
	"github.com/superp00t/etc"
	"github.com/superp00t/etc/yo"
	"github.com/superp00t/gophercraft/crypto"
	"github.com/superp00t/gophercraft/format/dbc"
	"github.com/superp00t/gophercraft/guid"
	"github.com/superp00t/gophercraft/packet"
	"github.com/superp00t/gophercraft/packet/update"
	"github.com/superp00t/gophercraft/realm/wdb"
)

func (s *Session) MaxPositiveAuras() int {
	return 32
}

type Creature struct {
	*update.ValuesBlock
	ID            string
	MovementBlock *update.MovementBlock
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

func (c *Creature) TypeID() guid.TypeID {
	return guid.TypeUnit
}

func (c *Creature) Values() *update.ValuesBlock {
	return c.ValuesBlock
}

func (c *Creature) Movement() *update.MovementBlock {
	return c.MovementBlock
}

func (ws *Server) NewCreature(template *wdb.CreatureTemplate, position update.Position) *Creature {
	id := guid.RealmSpecific(guid.Creature, ws.RealmID(), ws.NextDynamicCounter(guid.TypeUnit))
	c := new(Creature)
	var err error
	c.ValuesBlock, err = update.NewValuesBlock(ws.Build(), guid.TypeMaskObject|guid.TypeMaskUnit)
	if err != nil {
		panic(err)
	}

	c.MovementBlock = &update.MovementBlock{
		UpdateFlags: update.UpdateFlagLiving | update.UpdateFlagHasPosition | update.UpdateFlagAll,
		Speeds:      make(update.Speeds),
		Position:    position,
		Info: &update.MovementInfo{
			Position: position,
		},
	}

	for speedType, speed := range DefaultSpeeds {
		c.MovementBlock.Speeds[speedType] = speed
	}

	c.MovementBlock.Speeds[update.Run] = template.SpeedRun
	c.MovementBlock.Speeds[update.Walk] = template.SpeedWalk

	c.ID = template.ID
	c.SetGUID("GUID", id)
	c.SetUint32("Entry", template.Entry)

	scale := template.Scale
	if scale == 0 {
		scale = 1.0
	}

	c.SetFloat32("ScaleX", scale)
	display := template.DisplayIDs[etc.RandomIndex(template.DisplayIDs)]
	c.SetUint32("DisplayID", display)
	// c.SetUint32("NativeDisplayID", display)
	c.SetUint32("Level", crypto.RandUint32(template.MinLevel, template.MaxLevel))
	c.SetUint32("FactionTemplate", template.Faction)

	health := crypto.RandUint32(template.MinLevelHealth, template.MaxLevelHealth)
	c.SetUint32("Health", health)
	c.SetUint32("MaxHealth", health)
	c.SetUint32("Mana", 0)
	c.SetUint32("MaxMana", 0)

	c.SetFloat32("BoundingRadius", 1.0)
	c.SetFloat32("CombatReach", 1.0)

	// c.SetUint32("Armor", template.Armor)

	// NpcFlags: should not tied to any particular version.
	c.SetBit("Gossip", template.Gossip)
	c.SetBit("QuestGiver", template.QuestGiver)
	c.SetBit("Vendor", template.Vendor)
	c.SetBit("FlightMaster", template.FlightMaster)
	c.SetBit("Trainer", template.Trainer)
	c.SetBit("SpiritHealer", template.SpiritHealer)
	c.SetBit("SpiritGuide", template.SpiritGuide)
	c.SetBit("Innkeeper", template.Innkeeper)
	c.SetBit("Banker", template.Banker)
	c.SetBit("Petitioner", template.Petitioner)
	c.SetBit("TabardDesigner", template.TabardDesigner)
	c.SetBit("BattleMaster", template.BattleMaster)
	c.SetBit("Auctioneer", template.Auctioneer)
	c.SetBit("StableMaster", template.StableMaster)
	c.SetBit("Repairer", template.Repairer)

	c.SetBit("ServerControlled", template.ServerControlled)       // 0x1
	c.SetBit("NonAttackable", template.NonAttackable)             // 0x2
	c.SetBit("RemoveClientControl", template.RemoveClientControl) // 0x4
	c.SetBit("PlayerControlled", template.PlayerControlled)       // 0x8
	c.SetBit("Rename", template.Rename)                           // 0x10
	c.SetBit("PetAbandon", template.PetAbandon)                   // 0x20
	c.SetBit("OOCNotAttackable", template.OOCNotAttackable)       // 0x100
	c.SetBit("Passive", template.Passive)                         // 0x200
	c.SetBit("PVP", template.PVP)                                 // 0x1000
	c.SetBit("IsSilenced", template.IsSilenced)                   // 0x2000
	c.SetBit("IsPersuaded", template.IsPersuaded)                 // 0x4000
	c.SetBit("Swimming", template.Swimming)                       // 0x8000
	c.SetBit("RemoveAttackIcon", template.RemoveAttackIcon)       // 0x10000
	c.SetBit("IsPacified", template.IsPacified)                   // 0x20000
	c.SetBit("IsStunned", template.IsStunned)                     // 0x40000
	c.SetBit("InCombat", template.InCombat)                       // 0x80000
	c.SetBit("InTaxiFlight", template.InTaxiFlight)               // 0x100000
	c.SetBit("Disarmed", template.Disarmed)                       // 0x200000
	c.SetBit("Confused", template.Confused)                       // 0x400000
	c.SetBit("Fleeing", template.Fleeing)                         // 0x800000
	c.SetBit("Possessed", template.Possessed)                     // 0x1000000
	c.SetBit("NotSelectable", template.NotSelectable)             // 0x2000000
	c.SetBit("Skinnable", template.Skinnable)                     // 0x4000000
	c.SetBit("AurasVisible", template.AurasVisible)               // 0x8000000
	c.SetBit("Sheathe", template.Sheathe)                         // 0x40000000
	c.SetBit("NoKillReward", template.NoKillReward)               // 0x80000000

	return c
}

func (s *Session) HandleCreatureQuery(e *etc.Buffer) {
	entry := e.ReadUint32()
	creatureGUID := s.decodeUnpackedGUID(e)
	creatureObject := s.Map().GetObject(creatureGUID)

	var crt *wdb.CreatureTemplate
	s.DB().GetData(entry, &crt)

	resp := packet.NewWorldPacket(packet.SMSG_CREATURE_QUERY_RESPONSE)

	if crt == nil {
		resp.WriteUint32(entry | 0x80000000)
	} else {
		var creatureTypeFlags uint32
		if crt.Tameable {
			creatureTypeFlags |= 1
		} // Makes the mob tameable (must also be a beast and have family set)
		if crt.VisibleToGhosts {
			creatureTypeFlags |= 2
		} // Sets Creatures that can ALSO be seen when player is a ghost. Used in CanInteract function by client, can’t be attacked
		if crt.BossLevel {
			creatureTypeFlags |= 4
		}
		if crt.DontPlayWoundParryAnim {
			creatureTypeFlags |= 8
		}
		if crt.HideFactionTooltip {
			creatureTypeFlags |= 16
		} // Controls something in client tooltip related to creature faction
		if crt.SpellAttackable {
			creatureTypeFlags |= 64
		}
		if crt.DeadInteract {
			creatureTypeFlags |= 128
		}
		if crt.HerbLoot {
			creatureTypeFlags |= 256
		} // Uses Skinning Loot Field
		if crt.MiningLoot {
			creatureTypeFlags |= 512
		} // Makes Mob Corpse Mineable – Uses Skinning Loot Field
		if crt.DontLogDeath {
			creatureTypeFlags |= 1024
		}
		if crt.MountedCombat {
			creatureTypeFlags |= 2048
		}
		if crt.CanAssist {
			creatureTypeFlags |= 4096
		} //	Can aid any player or group in combat. Typically seen for escorting NPC’s
		if crt.PetHasActionBar {
			creatureTypeFlags |= 8192
		} // 	checked from calls in Lua_PetHasActionBar
		if crt.MaskUID {
			creatureTypeFlags |= 16384
		}
		if crt.EngineerLoot {
			creatureTypeFlags |= 32768
		} //	Makes Mob Corpse Engineer Lootable – Uses Skinning Loot Field
		if crt.ExoticPet {
			creatureTypeFlags |= 65536
		} // Tamable as an exotic pet. Normal tamable flag must also be set.
		if crt.UseDefaultCollisionBox {
			creatureTypeFlags |= 131072
		}
		if crt.IsSiegeWeapon {
			creatureTypeFlags |= 262144
		}
		if crt.ProjectileCollision {
			creatureTypeFlags |= 524288
		}
		if crt.HideNameplate {
			creatureTypeFlags |= 1048576
		}
		if crt.DontPlayMountedAnim {
			creatureTypeFlags |= 2097152
		}
		if crt.IsLinkAll {
			creatureTypeFlags |= 4194304
		}
		if crt.InteractOnlyWithCreator {
			creatureTypeFlags |= 8388608
		}
		if crt.ForceGossip {
			creatureTypeFlags |= 134217728
		}

		resp.WriteUint32(entry)
		resp.WriteCString(crt.Name.GetLocalized(s.Locale))
		resp.WriteCString("")
		resp.WriteCString("")
		resp.WriteCString("")
		resp.WriteCString(crt.SubName.GetLocalized(s.Locale))
		resp.WriteUint32(creatureTypeFlags)
		resp.WriteUint32(crt.CreatureType)

		if crt.Family == "" {
			resp.WriteUint32(0)
		} else {
			var creatureFamily *dbc.Ent_CreatureFamily
			s.DB().GetData(crt.Family, &creatureFamily)
			if creatureFamily == nil {
				yo.Warn("No creature family found for", crt.Family)
				resp.WriteUint32(0)
			} else {
				resp.WriteUint32(creatureFamily.ID)
			}
		}

		resp.WriteUint32(crt.Rank)
		resp.WriteUint32(0)
		resp.WriteUint32(crt.PetSpellDataId)

		var displayID uint32

		if creatureObject != nil {
			displayID = creatureObject.Values().GetUint32("DisplayID")
		} else {
			if len(crt.DisplayIDs) > 0 {
				displayID = crt.DisplayIDs[0]
			}
		}

		resp.WriteUint32(displayID)
		civilian := uint16(0)

		if crt.Civilian {
			civilian = 1
		}

		resp.WriteUint16(civilian)
	}

	s.SendAsync(resp)
}
