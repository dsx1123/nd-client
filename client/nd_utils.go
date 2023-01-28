package client

import "net/http"

const (
	LOGIN   = "/login"
	REFRESH = "/refresh"
)

type Client struct {
	url         string
	username    string
	password    string
	skipVerify  bool
	loginDomain string
	client      *http.Client
}

type authPayload struct {
	UserName   string `json:"userName"`
	UserPasswd string `json:"userPasswd"`
	Domain     string `json:"domain"`
}
