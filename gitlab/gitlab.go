package gitlab

import (
	"autodeploy/client"
	"autodeploy/marathon"
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

var (
	origin      string
	loginAction string
	loginURL    string
	username    string
	password    string
	session     client.Sessioner
)

// Config .
type Config struct {
	Origin      string
	LoginAction string
	Username    string
	Password    string
}

// Init init preset params.
func Init(cfg Config) (success bool, err error) {
	origin = cfg.Origin
	loginAction = cfg.LoginAction
	username = cfg.Username
	password = cfg.Password

	loginURL = origin + loginAction

	session, err = login()
	if err != nil {
		return false, err
	}
	return true, nil
}

// GetLatestTag get lastest tag.
func GetLatestTag(path string) (tag string, err error) {
	projectTagsURL := origin + "/" + path

	res, err := session.Get(projectTagsURL)

	if res.StatusCode == 404 {
		err = errors.New("project not found")
	}
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromResponse(res)
	if err != nil {
		return "", err
	}

	tagText := doc.Find("div.tags > ul").First().Find(".item-title").First().Text()
	reg, err := regexp.Compile("[^a-zA-Z0-9_.]+")
	tag = reg.ReplaceAllString(tagText, "")
	return tag, err
}

// NewTag create a new tag.
func NewTag(projectCfg marathon.Config) (newTag string, err error) {
	path := projectCfg.Maintainer + "/" + projectCfg.Name + "/tags"
	latestTag, err := GetLatestTag(path)
	if err != nil {
		return
	}
	newTag = addTagVersion(latestTag, "patch")

	// get authenticity_token
	newTagURL := origin + "/" + path + "/new"
	res, err := session.Get(newTagURL)
	if err != nil {
		return
	}
	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return
	}
	authenticityToken, exists := doc.Find("input[name=authenticity_token]").First().Attr("value")
	if !exists {
		err = errors.New("not found authenticity_token")
		return
	}

	postURL := origin + "/" + path
	formData := url.Values{
		"utf8":                {"✓"},
		"authenticity_token":  {authenticityToken},
		"tag_name":            {newTag},
		"ref":                 {"master"},
		"message":             {"new Tag " + newTag},
		"release_description": {""},
	}
	res, err = session.PostForm(postURL, formData)

	if res.StatusCode != 302 {
		err = errors.New("not 302 Found, create new tag failed")
	}
	return
}

func getBuildLogID(path string, tag string) (id string, err error) {
	piplinesTagsURL := path + "/pipelines?scope=tags"
	res, err := session.Get(piplinesTagsURL)
	if err != nil {
		return
	}
	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromResponse(res)
	if err != nil {
		return
	}

	id = ""
	doc.Find(".commit").Each(func(i int, selection *goquery.Selection) {
		curTag := selection.Find(".monospace.branch-name").First().Text()
		if curTag == tag {
			href, ok := selection.Find(".commit-link").First().Find("a").First().Attr("href")
			if !ok {
				err = errors.New("not found tag pipline id")
				return
			}
			splices := strings.Split(href, "/")
			id = splices[len(splices)-1]
		}
	})
	if id == "" {
		err = errors.New("not found tag pipline id")
		return
	}
	return
}

func getBuildLogURL(path string, buildID string) (buildLogURL string, err error) {
	piplinesURL := path + "/pipelines/" + buildID
	res, err := session.Get(piplinesURL)
	if err != nil {
		return
	}
	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromResponse(res)
	if err != nil {
		return
	}

	buildLogURL, ok := doc.Find(".pipeline-graph .stage-column:last-child .build-content a").First().Attr("href")
	if !ok {
		err = errors.New("not found pipline url")
		return
	}

	return
}

func judgeIsFinish(buildLogURL string) (ok bool, logContent string, image string, err error) {
	res, err := session.Get(buildLogURL + "/raw")
	if err != nil {
		ok = false
		return
	}

	contentBytes, err := ioutil.ReadAll(res.Body)
	logContent = string(contentBytes)

	matches := regexp.MustCompile(`building and pushing (.*?)\s`).FindStringSubmatch(logContent)
	if len(matches) > 1 {
		image = matches[1]
	}
	if strings.Contains(logContent, "[ERROR] ") {
		ok = false
		err = errors.New("Build failed")
		return
	}
	ok = strings.Contains(logContent, "Build succeeded")

	return
}

// WatchBuildLog watch build log.
func WatchBuildLog(projectCfg marathon.Config, tag string, showLogDetail bool) (ok bool, logContent string, image string, err error) {
	path := origin + "/" + projectCfg.Maintainer + "/" + projectCfg.Name
	buildLogID, err := getBuildLogID(path, tag)
	if err != nil {
		return
	}

	buildLogURL, err := getBuildLogURL(path, buildLogID)
	if err != nil {
		return
	}

	buildLogURL = origin + buildLogURL

	// log.Println("watching build log, waiting...")
	time.Sleep(time.Duration(60) * time.Second)

	currentLine := 0
	for {
		ok, logContent, image, err = judgeIsFinish(buildLogURL)
		if err != nil {
			break
		}

		if showLogDetail {
			s := string([]rune(logContent))
			splices := strings.Split(s, "\n")
			focusSlice := []string{}
			for _, splice := range splices {
				if strings.Contains(splice, "[INFO] ") {
					focusSlice = append(focusSlice, splice)
				} else if strings.Contains(splice, "[WARNING] ") {
					focusSlice = append(focusSlice, splice)
				} else if strings.Contains(splice, "[ERROR] ") {
					focusSlice = append(focusSlice, splice)
				} else if regexp.MustCompile(`\[\d{2}:\d{2}:\d{2}\]`).Match([]byte(splice)) {
					focusSlice = append(focusSlice, splice)
				}
			}
			if len(focusSlice)-1 > currentLine {
				fmt.Print("\n", strings.Join(focusSlice[currentLine:], "\n"))
				currentLine = len(focusSlice) - 1
			}
		}

		if ok {
			break
		}
		time.Sleep(time.Duration(10) * time.Second)
	}
	return
}
