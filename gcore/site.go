package gcore

import (
	"bytes"
	"html/template"
	"net/http"
	"strings"
	"time"

	"github.com/superp00t/etc"

	"github.com/superp00t/etc/yo"
	"github.com/superp00t/gophercraft/crypto/srp"
)

type navElement struct {
	Active bool
	Link   string
	Name   string
}

type pageBase struct {
	Title       string
	Brand       string
	NavElements []*navElement
	PageBody    template.HTML
}

func (c *Core) newPageBase(activeName string) *pageBase {
	navBar := []*navElement{
		{false, "/", "Home"},
		{false, "/armory", "Armory"},
		{false, "/account", "Account"},
	}

	for _, v := range navBar {
		if v.Name == activeName {
			v.Active = true
		}
	}

	brandName := "Gophercraft"

	return &pageBase{
		Title:       brandName + " " + activeName,
		Brand:       brandName,
		NavElements: navBar,
	}
}

func (c *Core) getTemplate(named string) *template.Template {
	tpl, err := template.ParseFiles(c.WebDirectory.Concat("template", named).Render())
	if err != nil {
		yo.Fatal(err)
	}

	return tpl
}

func (c *Core) Home(rw http.ResponseWriter, r *http.Request) {
	pg := c.newPageBase("Home")

	base := c.getTemplate("base.html")

	data, _ := c.WebDirectory.Concat("template", "home.html").ReadAll()
	pg.PageBody = template.HTML(data)

	if err := base.Execute(rw, pg); err != nil {
		panic(err)
	}
}

func (c *Core) SignUp(rw http.ResponseWriter, r *http.Request) {
	pg := c.newPageBase("Sign Up")

	base := c.getTemplate("base.html")

	data, _ := c.WebDirectory.Concat("template", "signUp.html").ReadAll()
	pg.PageBody = template.HTML(data)

	if err := base.Execute(rw, pg); err != nil {
		panic(err)
	}
}

func (c *Core) SignIn(r *Request) {
	signInRequest := struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}{}

	signInResponse := struct {
		Error    string `json:"error"`
		WebToken string `json:"webToken"`
	}{}

	if err := r.ScanJSON(&signInRequest); err != nil {
		r.Respond(http.StatusBadRequest, "malformed json")
		return
	}

	var acc Account

	found, _ := c.DB.Where("username = ?", signInRequest.Username).Get(&acc)
	if !found {
		signInResponse.Error = "invalid username or password"
		r.Encode(http.StatusOK, signInResponse)
		return
	}

	username := strings.ToUpper(signInRequest.Username)
	password := strings.ToUpper(signInRequest.Password)

	creds := srp.HashCredentials(username, password)

	ok := bytes.Equal(creds, acc.IdentityHash)

	if !ok {
		signInResponse.Error = "invalid username or password"
		r.Encode(http.StatusOK, signInResponse)
		return
	}

	wt := WebToken{
		Token:   etc.GenerateRandomUUID().String(),
		Account: acc.ID,
		Expiry:  time.Now().Add(12 * time.Hour),
	}

	c.DB.Insert(&wt)

	signInResponse.WebToken = wt.Token

	r.Encode(http.StatusOK, signInResponse)
}
