package gcore

import (
	"time"

	"github.com/superp00t/gophercraft/gcore/sys"
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
	Username     string
	IdentityHash []byte
}

type GameAccount struct {
	ID      uint64 `xorm:"'id' pk autoincr"`
	Name    string `xorm:"'name'"`
	Version uint32 `xorm:"'version'"`
	Owner   uint64 `xorm:"'owner'"`
}

type SessionKey struct {
	ID uint64 `xorm:"'id' pk"`
	K  []byte
}

type Realm struct {
	ID            uint64    `xorm:"'id' pk" json:"id"`
	Name          string    `xorm:"'name'" json:"name"`
	Version       uint32    `json:"version"`
	Type          string    `json:"type"`
	Address       string    `json:"address"`
	Description   string    `json:"description"`
	ActivePlayers uint32    `json:"activePlayers"`
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
