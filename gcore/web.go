package gcore

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/superp00t/etc"
	"github.com/superp00t/etc/yo"

	"github.com/dchest/captcha"
	"github.com/gorilla/mux"
	"github.com/superp00t/gophercraft/crypto/srp"
	"github.com/superp00t/gophercraft/gcore/sys"
)

func (r *Request) ScanJSON(v interface{}) error {
	b, err := ioutil.ReadAll(r.R.Body)
	if err != nil {
		return err
	}

	return json.Unmarshal(b, v)
}

func (c *Core) GetAuthStatus(r *Request) {
	getAuthStatusRequest := struct {
		Token string `json:"token"`
	}{}

	type getAuthStatusResponse struct {
		Valid   bool     `json:"valid"`
		Account string   `json:"account"`
		Tier    sys.Tier `json:"tier"`
	}

	err := r.ScanJSON(&getAuthStatusRequest)
	if err != nil {
		r.Respond(http.StatusBadRequest, "could not read json")
		return
	}

	var wt WebToken
	found, _ := c.DB.Where("token = ?", getAuthStatusRequest.Token).Get(&wt)
	if !found {
		r.Encode(http.StatusOK, getAuthStatusResponse{
			Valid: false,
		})
		return
	}

	if time.Since(wt.Expiry) > 0 {
		c.DB.Delete(&wt)
		r.Encode(http.StatusOK, getAuthStatusResponse{
			Valid: false,
		})
		return
	}

	var acc Account
	found, err = c.DB.Where("id = ?", wt.Account).Get(&acc)
	if !found {
		panic(err)
	}

	resp := getAuthStatusResponse{
		Valid:   true,
		Account: strings.ToLower(acc.Username),
		Tier:    acc.Tier,
	}

	r.Encode(http.StatusOK, resp)
}

func (c *Core) ResetAccount(user, pass string, tier sys.Tier) error {
	if user == "" {
		return fmt.Errorf("empty name")
	}

	var acc Account
	found, err := c.DB.Where("username = ?", user).Get(&acc)
	if err != nil {
		return err
	}

	acc.Username = user
	acc.Tier = tier
	acc.IdentityHash = srp.HashCredentials(user, pass)
	yo.Spew(acc.IdentityHash)

	if found {
		if _, err := c.DB.Where("id = ?", acc.ID).Cols("tier", "identity_hash").Update(&acc); err != nil {
			return err
		}
		return nil
	} else {
		if _, err := c.DB.Insert(&acc); err != nil {
			return err
		}
		_, err = c.DB.Insert(&GameAccount{
			Name:  "Zero",
			Owner: acc.ID,
		})
		return err
	}
}

func (c *Core) DoRegistration(u, p string) error {
	var acc []Account
	err := c.DB.Where("username = ?", u).Find(&acc)
	if err != nil {
		return err
	}

	if len(acc) > 0 {
		return fmt.Errorf("username taken")
	}

	idhash := srp.HashCredentials(u, p)

	acct := Account{
		Username:     u,
		IdentityHash: idhash,
	}

	_, err = c.DB.Insert(&acct)
	if err != nil {
		return err
	}

	_, err = c.DB.Insert(&GameAccount{
		Name:  "Zero",
		Owner: acct.ID,
	})

	return err
}

func (c *Core) AuthKey(apk string) int {
	if apk == c.APIKey() {
		return 2
	}

	return 0
}

func (c *Core) NewCaptcha(r *Request) {
	cp := captcha.New()
	r.Encode(http.StatusOK, CaptchaResponse{
		Status:    http.StatusOK,
		CaptchaID: cp,
	})
}

func (c *Core) UserExists(r *Request) {
	t := strings.ToUpper(r.Vars["username"])
	var acc []Account
	err := c.DB.Where("username = ?", t).Find(&acc)
	if err != nil {
		r.Respond(http.StatusInternalServerError, "Internal server error")
		return
	}

	r.Encode(http.StatusOK, UserExistsResponse{
		Status:     http.StatusOK,
		UserExists: len(acc) == 1,
	})
}

func (c *Core) Register(r *Request) {
	type registerRequest struct {
		Username        string `json:"username"`
		Password        string `json:"password"`
		CaptchaID       string `json:"captchaID"`
		CaptchaSolution string `json:"captchaSolution"`
	}

	var rr registerRequest

	if err := r.ScanJSON(&rr); err != nil {
		r.Respond(http.StatusBadRequest, "malformed json")
		return
	}

	type registerResponse struct {
		Error        string `json:"error"`
		ResetCaptcha bool   `json:"resetCaptcha"`
		WebToken     string `json:"webToken,omitempty"`
	}

	if rr.Username == "" || rr.Password == "" {
		r.Encode(http.StatusOK, registerResponse{
			Error: "username and password must not be empty",
		})
		return
	}

	if err := validateUsername(rr.Username); err != nil {
		r.Encode(http.StatusOK, registerResponse{
			Error: err.Error(),
		})
		return
	}

	if err := validatePassword(rr.Password); err != nil {
		r.Encode(http.StatusOK, registerResponse{
			Error: err.Error(),
		})
		return
	}

	if !captcha.VerifyString(rr.CaptchaID, rr.CaptchaSolution) {
		r.Encode(http.StatusOK, registerResponse{
			Error:        "Captcha failed.",
			ResetCaptcha: true,
		})
		return
	}

	if err := c.DoRegistration(rr.Username, rr.Password); err != nil {
		r.Encode(http.StatusOK, registerResponse{
			Error: err.Error(),
		})
		return
	}

	var acc Account

	found, err := c.DB.Where("username = ?", rr.Username).Get(&acc)
	if !found {
		panic(err)
	}

	wt := WebToken{
		Token:   etc.GenerateRandomUUID().String(),
		Account: acc.ID,
		Expiry:  time.Now().Add(12 * time.Hour),
	}

	c.DB.Insert(&wt)

	r.Encode(http.StatusOK, registerResponse{
		WebToken: wt.Token,
	})
}

func (c *Core) PublishRealmInfo(r Realm) uint64 {
	r.LastUpdated = time.Now()
	var rinf []Realm
	err := c.DB.Where("name = ?", r.Name).Find(&rinf)
	if err != nil {
		panic(err)
	}
	if len(rinf) == 0 {
		if _, err := c.DB.Insert(&r); err != nil {
			panic(err)
		}
	} else {
		if _, err := c.DB.AllCols().Update(&r); err != nil {
			panic(err)
		}
	}

	return r.ID
}

func (c *Core) RealmState() []Realm {
	var r []Realm
	c.DB.Find(&r)
	return r
}

func (c *Core) RealmList(r *Request) {
	r.Encode(http.StatusOK, map[string]interface{}{
		"status":  200,
		"listing": c.RealmState(),
	})
}

func (c *Core) WebAPI() http.Handler {
	r := mux.NewRouter()
	r.HandleFunc("/", c.Home)
	r.HandleFunc("/signUp", c.SignUp)
	r.PathPrefix("/assets/").Handler(http.StripPrefix("/assets/", http.FileServer(http.Dir(c.WebDirectory.Concat("assets").Render()))))

	v1 := r.PathPrefix("/v1/").Subrouter()
	v1.Handle("/signIn", c.Intercept(0, c.SignIn))
	v1.Handle("/realmList", c.Intercept(0, c.RealmList))
	v1.Handle("/getAuthStatus", c.Intercept(0, c.GetAuthStatus))
	v1.Handle("/newCaptcha", c.Intercept(0, c.NewCaptcha))
	v1.Handle("/userExists/{username}", c.Intercept(0, c.UserExists))
	v1.Handle("/register", c.Intercept(0, c.Register))
	v1.PathPrefix("/captcha/").Handler(captcha.Server(captcha.StdWidth, captcha.StdHeight))

	// admin/realm RPC functions

	// r.PathPrefix("/").Handler(http.FileServer(http.Dir(os.Getenv("GOPATH") + "src/github.com/superp00t/gophercraft/gcore/webapp/public/")))
	return r
}

func (c *Core) Intercept(required int, fn RequestHandler) *Interceptor {
	return &Interceptor{required, c, fn}
}

type Interceptor struct {
	requiredLevel int
	core          *Core
	fn            RequestHandler
}

func (s *Interceptor) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	rw.Header().Set("Access-Control-Allow-Origin", "*")
	rw.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	rw.Header().Set("Access-Control-Allow-Headers",
		"Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

	// TODO: Rate limiting and authorization
	lvl := s.core.AuthKey(req.URL.Query().Get("a"))

	if lvl < s.requiredLevel {
		r := &Request{
			RW:   rw,
			R:    req,
			Vars: mux.Vars(req),
		}
		r.Respond(http.StatusUnauthorized, "not enough clearance")
		return
	}

	s.fn(&Request{
		RW:   rw,
		R:    req,
		Vars: mux.Vars(req),
	})
}

func (c *Core) InfoHandler() http.Handler {
	r := mux.NewRouter()
	r.HandleFunc("/", func(rw http.ResponseWriter, r *http.Request) {
		rw.Write([]byte(`<p><a href="https://github.com/superp00t/gophercraft">Gophercraft ` + Version + `<a/></p>`))
	})
	return r
}

func validateUsername(input string) error {
	err := ""
	if len(input) > 16 {
		err = "username too long"
		goto end
	}

	if len(input) < 3 {
		err = "username too short"
		goto end
	}

	if ok, _ := regexp.MatchString("^[a-zA-Z]+$", input); ok == false {
		err = "invalid characters in name"
		goto end
	}

end:

	if err == "" {
		return nil
	}
	return errors.New(err)
}

func validatePassword(in string) error {
	err := ""

	input := []rune(in)

	if len(input) > 16 {
		err = "password too long"
		goto end
	}

	if len(input) < 6 {
		err = "password too short"
		goto end
	}

end:
	if err == "" {
		return nil
	}
	return errors.New(err)
}
