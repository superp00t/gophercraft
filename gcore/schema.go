package gcore

import (
	"time"

	"github.com/superp00t/gophercraft/gcore/config"
	"github.com/superp00t/gophercraft/gcore/sys"
	"github.com/superp00t/gophercraft/i18n"
	"github.com/superp00t/gophercraft/vsn"
)

type LoginTicket struct {
	Account string
	Ticket  string
	Expiry  time.Time
}

type WebToken struct {
	Token   string `xorm:"'token' pk"`
	Account uint64
	Expiry  time.Time
}

type Account struct {
	ID           uint64 `xorm:"'id' pk autoincr"`
	Tier         sys.Tier
	Locale       i18n.Locale
	Platform     string
	Username     string
	IdentityHash []byte
}

type GameAccount struct {
	ID    uint64 `xorm:"'id' pk autoincr"`
	Name  string `xorm:"'name'"`
	Owner uint64 `xorm:"'owner'"`
}

type SessionKey struct {
	ID uint64 `xorm:"'id' pk"`
	K  []byte
}

type Realm struct {
	ID            uint64    `xorm:"'id' pk" json:"id"`
	Name          string    `xorm:"'name'" json:"name"`
	Version       vsn.Build `json:"version"`
	Locked        bool
	Type          config.RealmType `json:"type"`
	Address       string           `json:"address"`
	Description   string           `json:"description"`
	ActivePlayers uint32           `json:"activePlayers"`
	Timezone      uint32
	LastUpdated   time.Time `json:"lastUpdated"`
}

type RegistrationBody struct {
	Username  string `json:"username"`
	Password  string `json:"password"`
	Expansion uint32 `json:"expansion"`
}

type CVar struct {
	RealmID uint64 `xorm:"'server_id'"`
	Key     string `xorm:"'key'"`
	Value   string
}
