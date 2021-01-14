package packet

import (
	"github.com/superp00t/gophercraft/guid"
	"github.com/superp00t/gophercraft/vsn"
)

const (
	GossipIconChat      = iota // White chat bubble
	GossipIconVendor           // 1 Brown bag
	GossipIconTaxi             // 2 Flight
	GossipIconTrainer          // 3 Book
	GossipIconInteract1        // 4	Interaction wheel
	GossipIconInteract2        // 5	Interaction wheel
	GossipIconGold             // 6 Brown bag with yellow dot (gold)
	GossipIconTalk             // White chat bubble with black dots (...)
	GossipIconTabard           // 8 Tabard
	GossipIconBattle           // 9 Two swords
	GossipIconDot              // 10 Yellow dot
	GossipIconChat11           // 11	White chat bubble
	GossipIconChat12           // 12	White chat bubble
	GossipIconChat13           // 13	White chat bubble
	GossipIconInvalid14        // 14	INVALID - DO NOT USE
	GossipIconInvalid15        // 15	INVALID - DO NOT USE
	GossipIconChat16           // 16	White chat bubble
	GossipIconChat17           // 17	White chat bubble
	GossipIconChat18           // 18	White chat bubble
	GossipIconChat19           // 19	White chat bubble
	GossipIconChat20           // 20	White chat bubble
	GossipIconTransmog         // 21	Transmogrifier?
)

type GossipItem struct {
	ItemID  uint32
	Icon    uint8
	Coded   bool
	Message string
}

type GossipQuestItem struct {
	QuestID    uint32
	QuestIcon  uint32
	QuestLevel int32
	QuestTitle string
}

type Gossip struct {
	Speaker    guid.GUID
	TextEntry  uint32
	Items      []GossipItem
	QuestItems []GossipQuestItem
}

func NewGossip(id guid.GUID, textID uint32) *Gossip {
	return &Gossip{Speaker: id, TextEntry: textID}
}

func (g *Gossip) SetTextEntry(entry uint32) {
	g.TextEntry = entry
}

func (g *Gossip) AddItem(itemID uint32, icon uint8, coded bool, message string) {
	g.Items = append(g.Items, GossipItem{itemID, icon, coded, message})
}

func (g *Gossip) Packet(build vsn.Build) *WorldPacket {
	p := NewWorldPacket(SMSG_GOSSIP_MESSAGE)
	g.Speaker.EncodeUnpacked(build, p)
	p.WriteUint32(g.TextEntry)
	p.WriteUint32(uint32(len(g.Items)))
	for _, item := range g.Items {
		p.WriteUint32(item.ItemID)
		p.WriteByte(item.Icon)
		p.WriteBool(item.Coded)
		p.WriteCString(item.Message)
	}

	p.WriteUint32(uint32(len(g.QuestItems)))
	for _, qItem := range g.QuestItems {
		p.WriteUint32(qItem.QuestID)
		p.WriteUint32(qItem.QuestIcon)
		p.WriteInt32(qItem.QuestLevel)
		p.WriteCString(qItem.QuestTitle)
	}

	return p
}

func (g *Gossip) GetSpeaker() guid.GUID {
	return g.Speaker
}
