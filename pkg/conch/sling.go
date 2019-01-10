// Copyright Joyent, Inc.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package conch

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"time"

	"github.com/dghubble/sling"
)

const (
	defaultUA      = "go-conch"
	defaultBaseURL = "https://conch.joyent.us"
)

var defaultTransport = &http.Transport{
	Proxy: http.ProxyFromEnvironment,
	Dial: (&net.Dialer{
		Timeout:   5 * time.Second,
		KeepAlive: 5 * time.Second,
		DualStack: true,
	}).Dial,
	TLSHandshakeTimeout: 5 * time.Second,
}

func (c *Conch) sling() *sling.Sling {
	if c.UA == "" {
		c.UA = defaultUA
	}

	if c.BaseURL == "" {
		c.BaseURL = defaultBaseURL
	}

	if c.CookieJar == nil {
		c.CookieJar, _ = cookiejar.New(nil)
	}

	if c.HTTPClient == nil {
		c.HTTPClient = &http.Client{
			Transport: defaultTransport,
			Jar:       c.CookieJar,

			// Preserve auth header on redirect
			// Inspired by: https://github.com/michiwend/gomusicbrainz/pull/4/files?utf8=%E2%9C%93&diff=unified
			// Under MIT License
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) > 30 {
					return fmt.Errorf("%d > 30 consecutive requests(redirects)", len(via))
				}
				if len(via) == 0 {
					// No redirects
					return nil
				}

				// This is a massive hack. In theory, go should already see
				// that these have the same host and copy the Authorization
				// header over on its own. Until I can track down why that's not
				// happening, this will get us back in business.
				// sungo [ 2018-06-21 ]
				if req.URL.Host == via[0].URL.Host {
					h, ok := via[0].Header["Authorization"]
					if ok {
						req.Header["Authorization"] = h
					}
				}
				return nil
			},
		}
	}

	s := sling.New().
		Client(c.HTTPClient).
		Base(c.BaseURL).
		Set("User-Agent", c.UA)

	u, _ := url.Parse(c.BaseURL)
	if c.JWToken != "" {
		if c.Expires == 0 {
			_ = c.recordJWTExpiry
		}

		s = s.Set("Authorization", "Bearer "+c.JWToken)

	} else if c.Session != "" {

		cookie := &http.Cookie{
			Name:  "conch",
			Value: c.Session,
		}
		c.CookieJar.SetCookies(
			u,
			[]*http.Cookie{cookie},
		)
	}

	return s
}

func (c *Conch) get(url string, data interface{}) error {
	req, err := c.sling().New().Get(url).Request()
	if err != nil {
		return err
	}

	_, err = c.httpDo(req, data)
	return err
}

func (c *Conch) httpDo(req *http.Request, data interface{}) (*http.Response, error) {
	res, err := c.HTTPClient.Do(req)
	if (res == nil) || (err != nil) {
		return res, err
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusUnauthorized {
		return res, ErrNotAuthorized
	}

	if res.StatusCode == http.StatusForbidden {
		return res, ErrForbidden
	}

	if res.StatusCode == http.StatusNotFound {
		return res, ErrDataNotFound
	}

	bodyBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return res, err
	}

	// BUG(sungo): an awfully simplistic view of the world
	if code := res.StatusCode; code >= 200 && code < 300 {
		if data != nil {
			// BUG(sungo): do we really want to throw away parse errors?
			json.Unmarshal(bodyBytes, data)
		}
		return res, nil
	}
	aerr := struct {
		Error string `json:"error"`
	}{""}
	if err := json.Unmarshal(bodyBytes, &aerr); err == nil {
		return res, errors.New(aerr.Error)
	}

	// In general, we should expect the API to give us error structures when
	// things go awry, but just in case not...
	return res, ErrHTTPNotOk
}

func (c *Conch) getWithQuery(url string, query interface{}, data interface{}) error {
	req, err := c.sling().New().Get(url).QueryStruct(query).Request()
	if err != nil {
		return err
	}
	_, err = c.httpDo(req, data)
	return err
}

func (c *Conch) httpDelete(url string) error {
	req, err := c.sling().New().Delete(url).Request()
	if err != nil {
		return err
	}
	_, err = c.httpDo(req, nil)
	return err
}

func (c *Conch) post(url string, payload interface{}, response interface{}) error {
	req, err := c.sling().New().
		Post(url).
		BodyJSON(payload).
		Request()

	if err != nil {
		return err
	}

	_, err = c.httpDo(req, response)
	return err
}

func (c *Conch) postNeedsResponse(
	url string,
	payload interface{},
	response interface{},

) (*http.Response, error) {
	req, err := c.sling().New().
		Post(url).
		BodyJSON(payload).
		Request()

	if err != nil {
		return nil, err
	}
	res, err := c.httpDo(req, response)
	return res, err
}

//////

// RawGet allows the user to perform an HTTP GET against the API, with the
// library handling all auth but *not* processing the response.
func (c *Conch) RawGet(url string) (*http.Response, error) {
	req, err := c.sling().New().Get(url).Request()
	if err != nil {
		return nil, err
	}

	return c.HTTPClient.Do(req)
}

// RawDelete allows the user to perform an HTTP DELETE against the API, with the
// library handling all auth but *not* processing the response.
func (c *Conch) RawDelete(url string) (*http.Response, error) {
	req, err := c.sling().New().Delete(url).Request()
	if err != nil {
		return nil, err
	}

	return c.HTTPClient.Do(req)
}

// RawPost allows the user to perform an HTTP POST against the API, with the
// library handling all auth but *not* processing the response.
// The provided body *must* be JSON for the server to accept it.
func (c *Conch) RawPost(url string, body io.Reader) (*http.Response, error) {
	req, err := c.sling().New().Post(url).
		Set("Content-Type", "application/json").Body(body).Request()
	if err != nil {
		return nil, err
	}

	return c.HTTPClient.Do(req)
}
