package worldserver

import (
	"fmt"
	"math"

	"github.com/superp00t/etc"
	"github.com/superp00t/etc/yo"
	"github.com/superp00t/gophercraft/guid"
	"github.com/superp00t/gophercraft/packet"
	"github.com/superp00t/gophercraft/packet/update"
	"github.com/superp00t/gophercraft/worldserver/wdb"
)

type GameObject struct {
	*update.ValuesBlock
	GOPosition update.Position
}

func (g *GameObject) GUID() guid.GUID {
	return g.GetGUIDValue(update.ObjectGUID)
}

func (g *GameObject) TypeID() guid.TypeID {
	return guid.TypeGameObject
}

func (g *GameObject) Values() *update.ValuesBlock {
	return g.ValuesBlock
}

func (g *GameObject) Living() bool {
	return false
}

func (g *GameObject) Position() update.Position {
	return g.GOPosition
}

func (g *GameObject) Speeds() update.Speeds {
	return nil
}

func (g *GameObject) SetRotation(orientation, rot0, rot1, rot2, rot3 float32) {
	if rot2 == 0 && rot3 == 0 {
		rot2 = float32(math.Sin(float64(orientation) / 2))
		rot3 = float32(math.Cos(float64(orientation) / 2))
	}

	g.SetFloat32Value(update.GObjectFacing, orientation)
	g.SetFloat32ArrayValue(update.GObjectRotation, 0, rot0)
	g.SetFloat32ArrayValue(update.GObjectRotation, 1, rot1)
	g.SetFloat32ArrayValue(update.GObjectRotation, 2, rot2)
	g.SetFloat32ArrayValue(update.GObjectRotation, 3, rot3)
}

func (m *Map) SpawnGameObject(gobjID string, pos update.Position) error {
	ws := m.Phase.Server

	var tpl wdb.GameObjectTemplate

	found, err := ws.DB.Where("id = ?", gobjID).Get(&tpl)
	if !found {
		if err != nil {
			panic(err)
		}

		return fmt.Errorf("could not find gameobject %s", gobjID)
	}

	gobj := &GameObject{update.NewValuesBlock(), pos}
	gobj.SetGUIDValue(update.ObjectGUID, guid.RealmSpecific(guid.GameObject, ws.RealmID(), ws.gameObjectCounter))
	ws.gameObjectCounter++
	gobj.SetTypeMask(ws.Config.Version, guid.TypeMaskObject|guid.TypeMaskGameObject)
	gobj.SetUint32Value(update.ObjectEntry, tpl.Entry)
	gobj.SetFloat32Value(update.ObjectScaleX, tpl.Size)

	gobj.SetUint32Value(update.GObjectDisplayID, tpl.DisplayID)
	gobj.SetUint32Value(update.GObjectTypeID, tpl.Type)
	gobj.SetUint32Value(update.GObjectFaction, tpl.Faction)

	flg, err := update.ParseGameObjectFlags(tpl.Flags)
	if err != nil {
		panic(err)
	}

	gobj.SetUint32Value(update.GObjectFlags, uint32(flg))
	gobj.SetFloat32Value(update.GObjectPosX, pos.X)
	gobj.SetFloat32Value(update.GObjectPosY, pos.Y)
	gobj.SetFloat32Value(update.GObjectPosZ, pos.Z)
	gobj.SetUint32Value(update.GObjectAnimProgress, 100)
	gobj.SetUint32Value(update.GObjectState, 0x01)
	gobj.Set(update.GObjectRotation, make([]*float32, 4))
	gobj.SetRotation(pos.O, 0, 0, 0, 0)

	return m.AddObject(gobj)
}

func (s *Session) HandleGameObjectQuery(e *etc.Buffer) {
	entry := e.ReadUint32()

	var tpl wdb.GameObjectTemplate

	found, err := s.DB().Where("entry = ?", entry).Get(&tpl)
	if !found {
		if err != nil {
			panic(err)
		}

		yo.Warn("entry not found", entry)

		resp := packet.NewWorldPacket(packet.SMSG_GAMEOBJECT_QUERY_RESPONSE)
		resp.WriteUint32(entry | 0x80000000)
		s.SendAsync(resp)
	} else {
		resp := packet.NewWorldPacket(packet.SMSG_GAMEOBJECT_QUERY_RESPONSE)
		resp.WriteUint32(entry)
		resp.WriteUint32(tpl.Type)
		resp.WriteUint32(tpl.DisplayID)
		resp.WriteCString(tpl.Name)
		resp.WriteByte(0)
		resp.WriteByte(0)
		resp.WriteByte(0)
		resp.WriteByte(0)
		for x := 0; x < 24; x++ {
			if x < len(tpl.Data) {
				resp.WriteUint32(tpl.Data[x])
			} else {
				resp.WriteUint32(0)
			}
		}
		s.SendAsync(resp)
	}
}

func (gobj *GameObject) GameObjectType() uint32 {
	return gobj.GetUint32Value(update.GObjectTypeID)
}

func (gobj *GameObject) Entry() uint32 {
	return gobj.GetUint32Value(update.ObjectEntry)
}

func (s *Session) HandleGameObjectUse(e *etc.Buffer) {
	g := s.decodeUnpackedGUID(e)
	yo.Ok("Using", g)

	if g.HighType() != guid.GameObject {
		return
	}

	wo := s.Map().GetObject(g)
	if wo == nil {
		return
	}

	gobj := wo.(*GameObject)

	switch gobj.GameObjectType() {
	case packet.GAMEOBJECT_TYPE_CHAIR:
		s.SitChair(gobj)
	}
}

func (s *Session) GetGameObjectTemplateByEntry(entry uint32) wdb.GameObjectTemplate {
	var gobjTemplate wdb.GameObjectTemplate
	found, err := s.DB().Where("entry = ?", entry).Get(&gobjTemplate)
	if !found {
		panic(err)
	}

	return gobjTemplate
}
