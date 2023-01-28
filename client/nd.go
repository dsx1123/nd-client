package client

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"time"
)

func New(url string, username string, password string, skipVerify bool, loginDomain string) *Client {
	c := &Client{
		url:         url,
		username:    username,
		password:    password,
		skipVerify:  skipVerify,
		loginDomain: loginDomain,
	}
	jar, err := cookiejar.New(nil)
	if err != nil {
		log.Fatalf("Got error while creating cookie jar %s", err.Error())
	}

	tc := &tls.Config{InsecureSkipVerify: skipVerify}
	tr := &http.Transport{
		MaxIdleConns:    10,
		IdleConnTimeout: 30 * time.Second,
		TLSClientConfig: tc,
	}

	c.client = &http.Client{
		Transport: tr,
		Jar:       jar,
	}

	return c
}

func (c *Client) GetUrl() string {
	return c.url
}

func (c *Client) GetUsername() string {
	return c.username
}

func (c *Client) Login() (*[]byte, error) {
	// login will initialize client by login
	ct := "application/json"
	var result *[]byte

	var authPayload = authPayload{
		UserName:   c.username,
		UserPasswd: c.password,
		Domain:     c.loginDomain,
	}

	payload, err := json.Marshal(authPayload)
	if err != nil {
		return result, err
	}

	result, err = c.Send(LOGIN, http.MethodPost, payload, ct)
	if err != nil {
		return result, err
	}
	return result, err

}

func (c *Client) Refresh() (*[]byte, error) {
	var err error
	ct := ""
	payload := []byte{}
	result, err := c.Send(REFRESH, http.MethodPost, payload, ct)
	if err != nil {
		return result, err
	}
	return result, err
}

func (c *Client) Send(endpoint string, method string, payload []byte, ct string) (*[]byte, error) {
	var resp *http.Response
	var errMsg string
	var err error
	var result []byte

	// refresh token before send http request
	if endpoint != REFRESH && endpoint != LOGIN {
		_, err = c.Refresh()
		// if refresh failed, try to login again to aquire new token
		if err != nil {
			c.Login()
		}
	}
	url := c.url + endpoint

	switch m := method; m {
	case http.MethodGet:
		resp, err = c.client.Get(url)
		if err != nil {
			return nil, err
		}
	case http.MethodPost:
		resp, err = c.client.Post(url, ct, bytes.NewBuffer(payload))
		if err != nil {
			return nil, err
		}
	case http.MethodPut:
		req, err := http.NewRequest("PUT", url, bytes.NewBuffer(payload))
		if ct != "" {
			req.Header.Add("Content-Type", ct)
		}
		if err != nil {
			return nil, err
		}
		resp, err = c.client.Do(req)
		if err != nil {
			return nil, err
		}
	case http.MethodDelete:
		req, err := http.NewRequest("DELETE", url, bytes.NewBuffer(payload))
		if ct != "" {
			req.Header.Add("Content-Type", ct)
		}
		if err != nil {
			return nil, err
		}
		resp, err = c.client.Do(req)
		if err != nil {
			return nil, err
		}
	default:
		errMsg = fmt.Sprintf("invalid method %s", method)
		return &result, errors.New("errMsg")
	}
	defer resp.Body.Close()

	if !(resp.StatusCode >= 200 && resp.StatusCode < 300) {
		body, _ := io.ReadAll(resp.Body)
		errMsg = fmt.Sprint(string(body))
		return &[]byte{}, errors.New(errMsg)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		errMsg = "IO error, could not read from response"
		return &result, errors.New(errMsg)
	}

	result = body
	return &result, err
}
