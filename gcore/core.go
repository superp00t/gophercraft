package gcore

import (
	"fmt"
	"log"
	"strings"

	"github.com/superp00t/etc"

	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/fatih/color"
	"github.com/superp00t/etc/yo"
	"github.com/superp00t/gophercraft/auth"
	"github.com/superp00t/gophercraft/gcore/config"
	"github.com/superp00t/gophercraft/packet"
	"github.com/superp00t/gophercraft/vsn"
	"xorm.io/xorm"
)

var (
	Version = vsn.CoreVersion
)

type Core struct {
	Auth         *config.Auth
	WebDirectory etc.Path

	Key string
	DB  *xorm.Engine
}

func NewCore(cfg *config.Auth) (*Core, error) {
	db, err := xorm.NewEngine(cfg.DBDriver, cfg.DBURL)
	if err != nil {
		return nil, err
	}
	core := &Core{DB: db, Auth: cfg}
	core.WebDirectory = etc.Import("github.com/superp00t/gophercraft/gcore/webapp")

	schemas := []interface{}{
		new(SessionKey),
		new(Account),
		new(GameAccount),
		new(Realm),
		new(LoginTicket),
		new(TrustedKey),
		new(WebToken),
		new(CVar),
	}

	for _, v := range schemas {
		err = core.DB.Sync2(v)
		if err != nil {
			return nil, err
		}
	}

	_, err = core.DB.Count(new(Account))
	if err != nil {
		return nil, err
	}

	if core.Key == "" {
		core.Key = etc.GenerateRandomUUID().String()
		yo.Ok("API key auto-generated: ", core.Key)
	}

	go func() {
		for {
			deleted, err := core.DB.Where("expiry < ?", time.Now()).Delete(new(WebToken))
			if err != nil {
				yo.Fatal(err)
			}

			if deleted > 0 {
				yo.Ok("Wiped", deleted, "expired tokens")
			}

			time.Sleep(20 * time.Minute)
		}
	}()

	return core, nil
}

func (c *Core) GetAccountID(user string) (uint64, error) {
	var acc []Account

	err := c.DB.Where("username = ?", user).Find(&acc)
	if err != nil {
		return 0, err
	}

	if len(acc) == 0 {
		return 0, fmt.Errorf("gcore: empty set")
	}

	return acc[0].ID, nil
}

func (c *Core) AccountID(user string) uint64 {
	id, err := c.GetAccountID(user)
	if err != nil {
		yo.Fatal(err)
	}

	return id
}

func (c *Core) StoreKey(user string, K []byte) {
	id := c.AccountID(user)

	c.DB.Where("id = ?", id).Delete(new(SessionKey))

	c.DB.Insert(&SessionKey{
		ID: id,
		K:  K,
	})
}

func (c *Core) GetAccount(user string) *auth.Account {
	var accs []Account
	c.DB.Where("username = ?", strings.ToUpper(user)).Find(&accs)
	if len(accs) == 0 {
		return nil
	}

	return &auth.Account{
		Username:     accs[0].Username,
		IdentityHash: accs[0].IdentityHash,
	}
}

func (c *Core) ListRealms(user string, build uint32) []packet.RealmListing {
	var acc []Account
	c.DB.Where("username = ?", user).Find(&acc)
	if len(acc) == 0 {
		log.Println("No user found!")
		return nil
	}

	var rlmState []Realm
	c.DB.Find(&rlmState)

	var rlm []packet.RealmListing
	for _, v := range rlmState {
		if v.Version == build {
			pkt := packet.RealmListing{}
			pkt.Type = packet.ConvertRealmType(v.Type)
			pkt.Locked = false
			pkt.Flags = 0x00
			if (time.Now().UnixNano() - v.LastUpdated.UnixNano()) > (time.Second * 15).Nanoseconds() {
				pkt.Flags = 0x02 // offline
			}
			pkt.Name = v.Name
			pkt.Address = v.Address
			pkt.Population = 1.0
			pkt.Timezone = 1
			pkt.ID = uint8(v.ID)
			// TODO, query this info from worldserver GRPC

			// c, _ := c.DB.Where("realm_id = ?", v.ID).Where("account = ?", acc[0].ID).Count(new(Character))
			// pkt.Characters = uint8(c)
			rlm = append(rlm, pkt)
		}
	}

	log.Println(spew.Sdump(rlm))

	return rlm
}

const banner = `
 ____ ____ ___  _  _ ____ ____ ____ ____ ____ ____ ___
 |__, [__] |--' |--| |=== |--< |___ |--< |--| |---  | 

 The programs included with Gophercraft are free software;
the exact distribution terms for each program are described in LICENSE.

`

func PrintLicense() {
	color.Set(color.FgCyan)
	fmt.Println(banner)
	color.Unset()
}

func (c *Core) APIKey() string {
	return c.Key
}

func (c *Core) GetCVar(k string) string {
	var cv []CVar
	c.DB.Where("key = ?", k).Where("server_id = ?", 0).Find(&cv)
	if len(cv) == 0 {
		return ""
	}
	return cv[0].Value
}

func (c *Core) SetCVar(k, v string) {
	cvar := &CVar{0, k, v}

	c.DB.Where("key = ?", k).AllCols().Update(cvar)
}
