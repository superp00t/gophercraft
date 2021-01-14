package realm

import (
	"math"

	"github.com/superp00t/etc"
	"github.com/superp00t/etc/yo"
	"github.com/superp00t/gophercraft/guid"
	"github.com/superp00t/gophercraft/packet"
	"github.com/superp00t/gophercraft/packet/update"
	"github.com/superp00t/gophercraft/vsn"
	"github.com/superp00t/gophercraft/realm/wdb"
)

type GameObject struct {
	Position update.Position
	*update.ValuesBlock
}

func (g *GameObject) GUID() guid.GUID {
	return g.GetGUID("GUID")
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

func (g *GameObject) SetPosition(pos update.Position) {
	if g.ValuesBlock.Descriptor.Build.RemovedIn(vsn.V3_3_5a) {
		g.SetFloat32("PosX", pos.X)
		g.SetFloat32("PosY", pos.Y)
		g.SetFloat32("PosZ", pos.Z)
		g.SetFloat32("Facing", pos.O)
	}

	g.Position = pos
}

func (g *GameObject) Speeds() update.Speeds {
	return nil
}

func (g *GameObject) SetRotation(orientation, rot0, rot1, rot2, rot3 float32) {
	if rot2 == 0 && rot3 == 0 {
		rot2 = float32(math.Sin(float64(orientation) / 2))
		rot3 = float32(math.Cos(float64(orientation) / 2))
	}

	g.Position.O = orientation
	g.SetPosition(g.Position)

	g.SetFloat32ArrayValue("Rotation", 0, rot0)
	g.SetFloat32ArrayValue("Rotation", 1, rot1)
	g.SetFloat32ArrayValue("Rotation", 2, rot2)
	g.SetFloat32ArrayValue("Rotation", 3, rot3)
}

func (ws *Server) NextGameObjectGUID() guid.GUID {
	g := guid.RealmSpecific(guid.GameObject, ws.RealmID(), ws.NextDynamicCounter(guid.TypeGameObject))
	return g
}

func (ws *Server) CreateGameObject(tpl *wdb.GameObjectTemplate, pos update.Position) *GameObject {
	valuesBlock, err := update.NewValuesBlock(ws.Build(), guid.TypeMaskObject|guid.TypeMaskGameObject)
	if err != nil {
		panic(err)
	}
	gobj := &GameObject{
		pos,
		valuesBlock,
	}

	gobj.SetGUID("GUID", ws.NextGameObjectGUID())

	gobj.SetUint32("Entry", tpl.Entry)
	gobj.SetFloat32("ScaleX", tpl.Size)

	gobj.SetUint32("DisplayID", tpl.DisplayID)
	gobj.SetUint32("TypeID", tpl.Type)
	gobj.SetUint32("Faction", tpl.Faction)

	gobj.SetUint32("Flags", uint32(tpl.Flags))

	gobj.SetPosition(pos)
	gobj.SetRotation(pos.O, 0, 0, 0, 0)

	gobj.SetUint32("AnimProgress", 100)
	gobj.SetUint32("State", 0x01)
	return gobj
}

func (m *Map) SpawnGameObject(gobjID string, pos update.Position) error {
	ws := m.Phase.Server

	tpl, err := ws.DB.GetGameObjectTemplate(gobjID)
	if err != nil {
		return err
	}

	gobj := ws.CreateGameObject(tpl, pos)

	return m.AddObject(gobj)
}

func (s *Session) HandleGameObjectQuery(e *etc.Buffer) {
	entry := e.ReadUint32()

	tpl, err := s.DB().GetGameObjectTemplateByEntry(entry)
	if tpl == nil {
		yo.Warn("entry not found", entry, err)

		resp := packet.NewWorldPacket(packet.SMSG_GAMEOBJECT_QUERY_RESPONSE)
		resp.WriteUint32(entry | 0x80000000)
		s.SendAsync(resp)
	} else {
		resp := packet.NewWorldPacket(packet.SMSG_GAMEOBJECT_QUERY_RESPONSE)
		resp.WriteUint32(entry)
		resp.WriteUint32(tpl.Type)
		resp.WriteUint32(tpl.DisplayID)
		resp.WriteCString(tpl.Name.GetLocalized(s.Locale))
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
	return gobj.GetUint32("TypeID")
}

func (gobj *GameObject) Entry() uint32 {
	return gobj.GetUint32("Entry")
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

func (s *Session) GetGameObjectTemplateByEntry(entry uint32) *wdb.GameObjectTemplate {
	var gobjTemplate *wdb.GameObjectTemplate
	s.DB().GetData(entry, gobjTemplate)
	return gobjTemplate
}

func (gobj *GameObject) Movement() *update.MovementBlock {
	return &update.MovementBlock{
		UpdateFlags: update.UpdateFlagHasPosition,
		Position:    gobj.Position,
	}
}
