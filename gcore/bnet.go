package gcore

import "time"

func (c *Core) StoreLoginTicket(user, ticket string, expiry time.Time) {
	c.DB.Insert(&LoginTicket{
		user, ticket, expiry,
	})
}

func (c *Core) GetTicket(ticket string) (string, time.Time) {
	var ticks []LoginTicket

	c.DB.Where("ticket = ?", ticket).Find(&ticks)
	if len(ticks) == 0 {
		return "", time.Time{}
	}

	if time.Since(ticks[0].Expiry) > 0 {
		c.DB.Where("ticket = ?", ticket).Delete(new(LoginTicket))
		return "", time.Time{}
	}

	return ticks[0].Account, ticks[0].Expiry
}
