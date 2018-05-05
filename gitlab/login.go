package gitlab

import (
	"autodeploy/client"
	"errors"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// login return a new gitlab session.
func login() (session client.Sessioner, err error) {
	// get cookie
	cookie, err := getCookie()

	// set session
	session = client.NewSession().
		SetCookie(cookie)
		// AddHeader("Origin", origin).
		// AddHeader("Referer", loginURL)

	return session, err
}

// getCookie return a valid login cookie.
func getCookie() (cookie string, err error) {
	authenticityToken, cookie, err := getAuthenticityToken()
	if err != nil {
		return
	}

	formData := url.Values{
		"utf8":               {"✓"},
		"authenticity_token": {authenticityToken},
		"user[login]":        {username},
		"user[password]":     {password},
		"user[remember_me]":  {"0"},
	}

	session := client.NewSession().
		SetCookie(cookie).
		AddHeader("Origin", origin).
		AddHeader("Referer", loginURL)

	res, err := session.PostForm(loginURL, formData)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode != 302 {
		return "", errors.New("username or password is wrong")
	}

	cookies := strings.Join(res.Header["Set-Cookie"], ";")
	matches := regexp.MustCompile(`(_gitlab_session=.*?);`).FindStringSubmatch(cookies)
	cookie = matches[1]

	return
}

// getAuthenticityToken return authenticityToken and preset cookie.
func getAuthenticityToken() (authenticityToken string, cookie string, err error) {
	res, err := http.Get(loginURL)
	if err != nil {
		return "", "", err
	}
	defer res.Body.Close()

	cookie = strings.Split(res.Header["Set-Cookie"][0], ";")[0]

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return "", cookie, err
	}

	authenticityToken, exists := doc.Find("input[name=authenticity_token]").First().Attr("value")
	if !exists {
		err = errors.New("not found authenticity_token")
	}
	if err != nil {
		return "", cookie, err
	}

	return authenticityToken, cookie, err
}
