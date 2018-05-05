package marathon

import (
	"autodeploy/client"
	"encoding/json"
	"io/ioutil"
	"strings"
	"time"

	"github.com/bitly/go-simplejson"
)

var session client.Sessioner

// Config .
type Config struct {
	Maintainer   string
	Name         string
	MarathonName string
	Short        string
}

type volume struct {
	ContainerPath string `json:"containerPath"`
	HostPath      string `json:"hostPath"`
	Mode          string `json:"mode"`
}

type portMapping struct {
	ContainerPort int               `json:"containerPort"`
	HostPort      int               `json:"hostPort"`
	ServicePort   int               `json:"servicePort"`
	Protocol      string            `json:"protocol"`
	Name          string            `json:"name"`
	Labels        map[string]string `json:"labels"`
}

type docker struct {
	Image          string        `json:"image"`
	Network        string        `json:"network"`
	PortMappings   []portMapping `json:"portMappings"`
	Privileged     bool          `json:"privileged"`
	Parameters     []string      `json:"parameters"`
	ForcePullImage bool          `json:"forcePullImage"`
}

type deployParams struct {
	Type    string   `json:"type"`
	Volumes []volume `json:"volumes"`
	Docker  docker   `json:"docker"`
}

type app struct {
	ID    string `json:"id"`
	Ports []int  `json:"ports"`
}

func init() {
	session = client.NewSession().SetBasicAuth("wisedu", "wiseduauth")
	return
}

// Deploy .
func Deploy(appName string, image string) (ok bool, err error) {
	appInfo, err := getAppID(appName)
	params := map[string]deployParams{
		"container": deployParams{
			Type: "DOCKER",
			Volumes: []volume{
				volume{
					ContainerPath: "/opt/logs",
					HostPath:      "/opt/logs",
					Mode:          "RW",
				},
			},
			Docker: docker{
				Image:   image,
				Network: "BRIDGE",
				PortMappings: []portMapping{
					portMapping{
						ContainerPort: 8080,
						HostPort:      0,
						ServicePort:   appInfo.Ports[0],
						Protocol:      "tcp",
						Name:          appName,
						Labels:        map[string]string{},
					},
				},
				Privileged:     false,
				Parameters:     []string{},
				ForcePullImage: true,
			},
		},
	}
	jsonParams, _ := json.Marshal(params)

	res, err := session.Put("http://172.16.7.23:8080/v2/apps/"+appInfo.ID, string(jsonParams))
	if err != nil {
		return
	}

	resp, err := ioutil.ReadAll(res.Body)
	js, err := simplejson.NewJson([]byte(string(resp)))
	if err != nil {
		return
	}

	deploymentID, err := js.Get("deploymentId").String()
	if err != nil {
		return
	}

	time.Sleep(time.Duration(5) * time.Second)
	ok, err = checkDeployDone(deploymentID)
	return
}

func checkDeployDone(deploymentID string) (ok bool, err error) {
	found := false
	for {
		res, err := session.Get("http://172.16.7.23:8080/v2/deployments")
		if err != nil {
			return false, err
		}
		resp, err := ioutil.ReadAll(res.Body)
		js, err := simplejson.NewJson([]byte(string(resp)))
		jsDeploymentArr, err := js.Array()

		found = false
		for i := range jsDeploymentArr {
			jsDeployment := js.GetIndex(i)
			if id, err := jsDeployment.Get("id").String(); id == deploymentID {
				if err != nil {
					break
				}
				found = true
			}
		}
		if found {
			time.Sleep(time.Duration(5) * time.Second)
		} else {
			ok = true
			break
		}
	}
	return
}

func getAppID(appName string) (appInfo app, err error) {
	appInfo = app{}

	res, err := session.Get("http://172.16.7.23:8080/v2/groups")
	if err != nil {
		return
	}

	resp, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return
	}

	js, err := simplejson.NewJson([]byte(string(resp)))

	apps, _ := js.Get("apps").Array()
	for i := range apps {
		jsApp := js.Get("apps").GetIndex(i)

		// get app id
		id, _ := jsApp.Get("id").String()

		if strings.Contains(id, appName) {
			jsPorts := jsApp.Get("ports")
			jsPortsArr, _ := jsPorts.Array()

			// get app ports
			ports := []int{}
			for j := range jsPortsArr {
				port, _ := jsPorts.GetIndex(j).Int()
				ports = append(ports, port)
			}

			appInfo.ID = id
			appInfo.Ports = ports
		}
	}

	return
}
