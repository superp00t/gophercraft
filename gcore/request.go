package gcore

import (
	"encoding/json"
	"net/http"
)

type Request struct {
	RW        http.ResponseWriter
	R         *http.Request
	Vars      map[string]string
	AuthLevel int
}

type RequestHandler func(*Request)

func (r *Request) Encode(status int, v interface{}) {
	enc := json.NewEncoder(r.RW)

	if r.R.URL.Query().Get("fmt") != "" {
		enc.SetIndent("", " ")
	}
	r.RW.Header().Set("Content-Type", "application/json; charset: utf-8")
	r.RW.WriteHeader(status)
	enc.Encode(v)
}

type CaptchaResponse struct {
	Status    int    `json:"status"`
	CaptchaID string `json:"captchaID"`
}

type GenericResponse struct {
	Status int    `json:"status"`
	Error  string `json:"error"`
}

type UserExistsResponse struct {
	Status     int  `json:"status"`
	UserExists bool `json:"exists"`
}

func (r *Request) Respond(status int, err string) {
	r.Encode(status, &GenericResponse{
		Status: status,
		Error:  err,
	})
}
