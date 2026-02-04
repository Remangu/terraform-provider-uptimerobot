package uptimerobotapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

var statusPageStatus = map[string]int{
	"paused": 0,
	"active": 1,
}
var StatusPageStatus = mapKeys(statusPageStatus)

var statusPageSort = map[string]int{
	"a-z":            1,
	"z-a":            2,
	"up-down-paused": 3,
	"down-up-paused": 4,
}
var StatusPageSort = mapKeys(statusPageSort)

type StatusPage struct {
	ID           int
	FriendlyName string `json:"friendly_name"`
	// CustomDomain string `json:"custom_domain"`
	StandardURL string `json:"standard_url"`
	CustomURL   string `json:"custom_url"`
	Sort        string
	Status      string
	DNSAddress  string
	Monitors    []int `json:"monitors"`
}

func (client UptimeRobotApiClient) GetStatusPage(id int) (sp StatusPage, err error) {
	sp.ID = id
	data := url.Values{}
	data.Add("psps", fmt.Sprintf("%d", id))

	body, err := client.MakeCall(
		"getPSPs",
		data.Encode(),
	)
	if err != nil {
		return
	}

	psps, ok := body["psps"].([]interface{})
	if !ok {
		j, _ := json.Marshal(body)
		err = errors.New("Unknown response from the server: " + string(j))
		return
	}

	if len(psps) == 0 {
		err = fmt.Errorf("Status page not found: %d", id)
		return
	}

	psp, ok := psps[0].(map[string]interface{})
	if !ok {
		err = errors.New("Invalid status page data from server")
		return
	}

	sp.FriendlyName, _ = psp["friendly_name"].(string)
	sp.StandardURL, _ = psp["standard_url"].(string)
	if psp["custom_url"] != nil {
		sp.CustomURL, _ = psp["custom_url"].(string)
	}
	if sortVal, ok := psp["sort"].(float64); ok {
		sp.Sort = intToString(statusPageSort, int(sortVal))
	}
	if statusVal, ok := psp["status"].(float64); ok {
		sp.Status = intToString(statusPageStatus, int(statusVal))
	}

	monitor, ok := psp["monitors"].(float64)
	if ok {
		sp.Monitors = []int{int(monitor)}
	} else {
		rawMonitors, ok := psp["monitors"].([]interface{})
		if ok {
			monitors := make([]int, len(rawMonitors))
			for k, v := range rawMonitors {
				if monVal, ok := v.(float64); ok {
					monitors[k] = int(monVal)
				}
			}
			sp.Monitors = monitors
		}
	}

	sp.DNSAddress = "stats.uptimerobot.com"

	return
}

type StatusPageCreateRequest struct {
	FriendlyName string
	CustomDomain string
	Password     string
	Monitors     []int
	Status       string
	Sort         string
}

// CreateStatusPage creates a new status page
func (client UptimeRobotApiClient) CreateStatusPage(req StatusPageCreateRequest) (sp StatusPage, err error) {
	data := url.Values{}
	data.Add("type", fmt.Sprintf("%d", 1))
	data.Add("friendly_name", req.FriendlyName)
	data.Add("custom_domain", req.CustomDomain)
	if req.Password != "" {
		data.Add("password", req.Password)
	}
	if len(req.Monitors) == 0 {
		data.Add("monitors", "0")
	} else {
		var monitors = req.Monitors
		var strMonitors = make([]string, len(monitors))
		for i, v := range monitors {
			strMonitors[i] = strconv.Itoa(v)
		}
		data.Add("monitors", strings.Join(strMonitors, "-"))
	}
	data.Add("sort", fmt.Sprintf("%d", statusPageSort[req.Sort]))
	// FIXME - Got error from UptimeRobot: {"message":"\"status\" is not allowed","parameter_name":"status","passed_value":"1","type":"invalid_parameter"}
	// data.Add("status", fmt.Sprintf("%d", statusPageStatus[req.Status]))

	body, err := client.MakeCall(
		"newPSP",
		data.Encode(),
	)
	if err != nil {
		return
	}
	psp, ok := body["psp"].(map[string]interface{})
	if !ok {
		err = errors.New("Invalid response from server when creating status page")
		return
	}
	idVal, ok := psp["id"].(float64)
	if !ok {
		err = errors.New("Invalid status page ID in response")
		return
	}
	return client.GetStatusPage(int(idVal))
}

type StatusPageUpdateRequest struct {
	ID           int
	FriendlyName string
	CustomDomain string
	Password     string
	Monitors     []int
	Status       string
	Sort         string
}

// UpdateStatusPage updates an existing status page
func (client UptimeRobotApiClient) UpdateStatusPage(req StatusPageUpdateRequest) (sp StatusPage, err error) {
	data := url.Values{}
	data.Add("id", fmt.Sprintf("%d", req.ID))
	data.Add("type", fmt.Sprintf("%d", 1))
	data.Add("friendly_name", req.FriendlyName)
	data.Add("custom_domain", req.CustomDomain)
	if req.Password != "" {
		data.Add("password", req.Password)
	}
	if len(req.Monitors) == 0 {
		data.Add("monitors", "0")
	} else {
		var monitors = req.Monitors
		var strMonitors = make([]string, len(monitors))
		for i, v := range monitors {
			strMonitors[i] = strconv.Itoa(v)
		}
		data.Add("monitors", strings.Join(strMonitors, "-"))
	}
	data.Add("sort", fmt.Sprintf("%d", statusPageSort[req.Sort]))
	data.Add("status", fmt.Sprintf("%d", statusPageStatus[req.Status]))

	_, err = client.MakeCall(
		"editPSP",
		data.Encode(),
	)
	if err != nil {
		return
	}
	return client.GetStatusPage(req.ID)
}

// DeleteStatusPage updates an existing status page
func (client UptimeRobotApiClient) DeleteStatusPage(id int) (err error) {
	data := url.Values{}
	data.Add("id", fmt.Sprintf("%d", id))

	_, err = client.MakeCall(
		"deletePSP",
		data.Encode(),
	)
	if err != nil {
		return
	}
	return
}
