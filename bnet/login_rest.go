package bnet

import (
	"bytes"
	"net/http"
	"strings"
	"time"

	"github.com/golang/protobuf/jsonpb"
	"github.com/superp00t/etc/yo"
	"github.com/superp00t/gophercraft/bnet/login"
	"github.com/superp00t/gophercraft/crypto/srp"
)

func formInput(inputID, aType, label string, maxLength uint32) *login.FormInput {
	fi := new(login.FormInput)
	fi.InputId = &inputID
	fi.Type = &aType
	fi.Label = &label
	if maxLength > 0 {
		fi.MaxLength = &maxLength
	}
	return fi
}

func getFormInputs() *login.FormInputs {
	formInputs := &login.FormInputs{}
	formType := login.FormType_LOGIN_FORM
	formInputs.Type = &formType
	formInputs.Inputs = []*login.FormInput{
		formInput("account_name", "text", "E-mail", 320),
		formInput("password", "password", "Password", 16),
		formInput("log_in_submit", "submit", "Log In", 0),
	}
	return formInputs
}

func (lst *Listener) HandleLoginGet(rw http.ResponseWriter, r *http.Request) {
	fi := getFormInputs()
	yo.Ok("REST", r.Method, r.URL.String())
	m := marshal()
	str, _ := m.MarshalToString(fi)
	rw.Header().Set("Content-Type", "application/json;charset=utf-8")
	rw.Write([]byte(str))
}

func (lst *Listener) HandleLoginPost(rw http.ResponseWriter, r *http.Request) {
	yo.Ok("REST", r.Method, r.URL.String())
	var lform login.LoginForm
	if err := jsonpb.Unmarshal(r.Body, &lform); err != nil {
		res := login.LoginResult{}
		as := login.AuthenticationState_LOGIN
		res.AuthenticationState = &as
		res.ErrorCode = strp("UNABLE_TO_DECODE")
		res.ErrorMessage = strp("There was an internal error while connecting to Battle.net. Please try again later.")
		sendResult(rw, &res)
		return
	}

	var username, password string
	for _, v := range lform.GetInputs() {
		switch v.GetInputId() {
		case "account_name":
			username = v.GetValue()
		case "password":
			password = v.GetValue()
		}
	}

	fakeTicket := "TC-0000000000000000000000000000000000000000"

	username = strings.ToUpper(username)
	password = strings.ToUpper(password)

	creds := srp.HashCredentials(username, password)

	done := login.AuthenticationState_DONE
	invalidResult := &login.LoginResult{
		AuthenticationState: &done,
		// ErrorCode:           strp("ERROR_LOGON_INVALID_SERVER_PROOF"),
		// ErrorMessage:        strp("Invalid username or password."),
		LoginTicket: &fakeTicket,
	}

	acc, _, err := lst.Backend.GetAccount(username)
	if err != nil {
		sendResult(rw, invalidResult)
		return
	}

	if !bytes.Equal(creds, acc.IdentityHash) {
		sendResult(rw, invalidResult)
		return
	}

	result := &login.LoginResult{
		AuthenticationState: &done,
	}

	ticket := GenerateTicket()
	lst.Backend.StoreLoginTicket(username, ticket, time.Now().Add(3600*time.Second))

	result.LoginTicket = &ticket

	sendResult(rw, result)
}

func sendResult(rw http.ResponseWriter, v *login.LoginResult) {
	yo.Ok("Sent result")
	m := marshal()
	str, err := m.MarshalToString(v)
	if err != nil {
		yo.Fatal(err)
	}
	rw.Header().Set("Content-Type", "application/json; charset=utf-8")
	rw.Write([]byte(str))
}

func strp(s string) *string {
	return &s
}

func u32p(v uint32) *uint32 {
	return &v
}

func u64p(v uint64) *uint64 {
	return &v
}

func boolp(boolean bool) *bool {
	return &boolean
}

func marshal() *jsonpb.Marshaler {
	return &jsonpb.Marshaler{
		OrigName: true,
	}
}
