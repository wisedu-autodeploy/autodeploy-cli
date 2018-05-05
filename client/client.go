package client

import (
	"net/http"
	"net/url"
	"strings"
)

// Session save cookie and headers.
type Session struct {
	cookie    string
	headers   map[string]string
	basicAuth [2]string
}

// NewSession new a *Session.
func NewSession() *Session {
	return &Session{"", map[string]string{}, [2]string{}}
}

// Sessioner .
type Sessioner interface {
	SetCookie(string) *Session
	AddHeader(string, string) *Session
	SetBasicAuth(string, string) *Session
	Get(string) (*http.Response, error)
	PostForm(string, url.Values) (*http.Response, error)
	Put(string, string) (*http.Response, error)
}

// SetCookie set cookie.
func (session *Session) SetCookie(cookie string) *Session {
	session.cookie = cookie
	return session
}

// AddHeader add header.
func (session *Session) AddHeader(key string, val string) *Session {
	session.headers[key] = val
	return session
}

// SetBasicAuth set basic auth
func (session *Session) SetBasicAuth(username string, password string) *Session {
	session.basicAuth = [2]string{username, password}
	return session
}

// Get make a get request with session.
func (session *Session) Get(url string) (res *http.Response, err error) {
	client := &http.Client{
		// CheckRedirect: func(req *http.Request, via []*http.Request) error {
		// 	return http.ErrUseLastResponse
		// },
	}
	req, err := http.NewRequest("GET", url, nil)
	req.SetBasicAuth(session.basicAuth[0], session.basicAuth[1])
	req.Header.Set("Cookie", session.cookie)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_4) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/65.0.3325.181 Safari/537.36")
	for key, val := range session.headers {
		req.Header.Set(key, val)
	}
	res, err = client.Do(req)
	return
}

// PostForm make a postForm request with session.
func (session *Session) PostForm(url string, formData url.Values) (res *http.Response, err error) {
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	req, err := http.NewRequest("POST", url, strings.NewReader(formData.Encode()))
	req.Header.Set("Cookie", session.cookie)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_4) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/65.0.3325.181 Safari/537.36")
	for key, val := range session.headers {
		req.Header.Set(key, val)
	}
	res, err = client.Do(req)
	return
}

// Put put a request.
func (session *Session) Put(url string, formData string) (res *http.Response, err error) {
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	req, err := http.NewRequest("PUT", url, strings.NewReader(formData))
	req.SetBasicAuth(session.basicAuth[0], session.basicAuth[1])
	req.Header.Set("Cookie", session.cookie)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_4) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/65.0.3325.181 Safari/537.36")
	for key, val := range session.headers {
		req.Header.Set(key, val)
	}
	res, err = client.Do(req)
	return
}
