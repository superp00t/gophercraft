package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type Client struct {
	Client   *http.Client
	APIKey   string
	Endpoint string
}

func (cl *Client) PublishRealmInfo(realm, ip, stype string) error {
	ep := fmt.Sprintf("/v1/publishRealmInfo/%s/%s/%s", realm, ip, stype)
	raw, err := url.Parse(cl.Endpoint + ep)
	if err != nil {
		return err
	}
	rq := raw.Query()
	rq.Set("a", cl.APIKey)
	raw.RawQuery = rq.Encode()
	return cl.GetJSON(raw.String(), nil)
}

func (cl *Client) GetJSON(url string, v interface{}) error {
	rsp, err := cl.Client.Get(url)
	if err != nil {
		return err
	}
	return json.NewDecoder(rsp.Body).Decode(&v)
}

func NewClient(endpoint, apiKey string) *Client {
	cl := &Client{}
	cl.Client = &http.Client{}
	cl.APIKey = apiKey
	cl.Endpoint = endpoint
	if strings.HasSuffix(cl.Endpoint, "/") {
		cl.Endpoint = strings.TrimRight(cl.Endpoint, "/")
	}
	return cl
}
