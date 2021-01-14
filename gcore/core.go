//Package gcore implements the main functionality of the Gophercraft Core Authserver.
package gcore

import (
	"fmt"
	"strings"

	"github.com/superp00t/gophercraft/i18n"

	"github.com/superp00t/gophercraft/gcore/sys"

	"github.com/superp00t/etc"

	"time"

	"github.com/superp00t/etc/yo"
	"github.com/superp00t/gophercraft/gcore/config"
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

	for user, pass := range cfg.Admin {
		if err := core.ResetAccount(user, pass, sys.Tier_Admin); err != nil {
			return nil, err
		}
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

func (c *Core) StoreKey(user, locale, platform string, K []byte) {
	id := c.AccountID(user)

	loc, err := i18n.LocaleFromString(locale)
	if err != nil {
		// Fallback
		loc = i18n.English
	}

	c.DB.Where("id = ?", id).Cols("locale", "platform").Update(&Account{
		Locale:   loc,
		Platform: platform,
	})

	c.DB.Where("id = ?", id).Delete(new(SessionKey))

	c.DB.Insert(&SessionKey{
		ID: id,
		K:  K,
	})
}

func (c *Core) GetAccount(user string) (*Account, []GameAccount, error) {
	var acc Account
	found, err := c.DB.Where("username = ?", strings.ToUpper(user)).Get(&acc)
	if err != nil {
		return nil, nil, err
	}

	if !found {
		return nil, nil, fmt.Errorf("account %s not found", user)
	}

	var gameAccs []GameAccount
	c.DB.Where("owner = ?", acc.ID).Find(&gameAccs)

	return &acc, gameAccs, nil
}

func (r Realm) Offline() bool {
	return (time.Now().UnixNano() - r.LastUpdated.UnixNano()) > (time.Second * 15).Nanoseconds()
}

func (c *Core) ListRealms() []Realm {
	var rlmState []Realm
	if err := c.DB.Find(&rlmState); err != nil {
		panic(err)
	}
	return rlmState
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
