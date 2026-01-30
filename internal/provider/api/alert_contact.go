package uptimerobotapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
)

// maximum pagination depth to allow (10*50=500 entries)
const page_limit = 10

var alertContactType = map[string]int{
	"sms":         1,
	"e-mail":      2,
	"twitter":     3,
	"twitter-dm":  3,
	"boxcar":      4,
	"web-hook":    5,
	"webhook":     5,
	"pushbullet":  6,
	"zapier":      7,
	"pro-sms":     8,
	"pushover":    9,
	"slack":       11,
	"voice-call":  14,
	"splunk":      15,
	"pagerduty":   16,
	"opsgenie":    17,
	"telegram":    18,
	"ms-teams":    20,
	"google-chat": 21,
	"hangouts":    21,
	"discord":     23,
}
var AlertContactType = mapKeys(alertContactType)

var alertContactStatus = map[string]int{
	"not activated": 0,
	"paused":        1,
	"active":        2,
}

var AlertContactStatus = mapKeys(alertContactStatus)

type AlertContact struct {
	ID           string `json:"id"`
	FriendlyName string `json:"friendly_name"`
	Value        string `json:"value"`
	Type         string
	Status       string
}

func (client UptimeRobotApiClient) GetAlertContacts() (acs []AlertContact, err error) {
	data := url.Values{}
	maxRecords := 50
	data.Add("limit", fmt.Sprintf("%d", maxRecords))

	var total int
	offset := 0

	for i := 0; i < page_limit; i++ {
		data.Set("offset", fmt.Sprintf("%d", offset))

		body, err := client.MakeCall(
			"getAlertContacts",
			data.Encode(),
		)
		if err != nil {
			return nil, err
		}

		alertcontacts, ok := body["alert_contacts"].([]interface{})
		if !ok {
			j, _ := json.Marshal(body)
			err = errors.New("Unknown response from the server: " + string(j))
			return nil, err
		}

		for _, i := range alertcontacts {
			alertcontact, ok := i.(map[string]interface{})
			if !ok {
				continue
			}
			id, _ := alertcontact["id"].(string)
			friendlyName, _ := alertcontact["friendly_name"].(string)
			value := ""
			if alertcontact["value"] != nil {
				value, _ = alertcontact["value"].(string)
			}
			typeVal, _ := alertcontact["type"].(float64)
			statusVal, _ := alertcontact["status"].(float64)
			ac := AlertContact{
				id,
				friendlyName,
				value,
				intToString(alertContactType, int(typeVal)),
				intToString(alertContactStatus, int(statusVal)),
			}
			acs = append(acs, ac)
		}

		if totalVal, ok := body["total"].(float64); ok {
			total = int(totalVal)
		}
		offset += maxRecords
		if len(acs) >= total {
			break
		}
	}

	if len(acs) < total {
		err = fmt.Errorf("Hitting pagination limit of: %d", page_limit)
	}

	return
}

func (client UptimeRobotApiClient) GetAlertContact(id string) (ac AlertContact, err error) {
	ac.ID = id
	data := url.Values{}
	data.Add("alert_contacts", id)

	body, err := client.MakeCall(
		"getAlertContacts",
		data.Encode(),
	)
	if err != nil {
		return
	}

	alertcontacts, ok := body["alert_contacts"].([]interface{})
	if !ok {
		j, _ := json.Marshal(body)
		err = errors.New("Unknown response from the server: " + string(j))
		return
	}

	if len(alertcontacts) == 0 {
		err = fmt.Errorf("Alert contact not found: %s", id)
		return
	}

	// Find the alert contact with matching ID (API may return multiple contacts)
	var alertcontact map[string]interface{}
	for _, item := range alertcontacts {
		contact, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		contactID, _ := contact["id"].(string)
		if contactID == id {
			alertcontact = contact
			break
		}
	}

	if alertcontact == nil {
		err = fmt.Errorf("Alert contact not found: %s", id)
		return
	}

	ac.FriendlyName, _ = alertcontact["friendly_name"].(string)
	ac.Value, _ = alertcontact["value"].(string)
	if typeVal, ok := alertcontact["type"].(float64); ok {
		ac.Type = intToString(alertContactType, int(typeVal))
	}
	if statusVal, ok := alertcontact["status"].(float64); ok {
		ac.Status = intToString(alertContactStatus, int(statusVal))
	}

	return
}

type AlertContactCreateRequest struct {
	FriendlyName string
	Type         string
	Value        string
}

func (client UptimeRobotApiClient) CreateAlertContact(req AlertContactCreateRequest) (ac AlertContact, err error) {
	data := url.Values{}
	data.Add("friendly_name", req.FriendlyName)
	data.Add("type", fmt.Sprintf("%d", alertContactType[req.Type]))
	data.Add("value", req.Value)

	body, err := client.MakeCall(
		"newAlertContact",
		data.Encode(),
	)
	if err != nil {
		return
	}

	alertcontact, ok := body["alertcontact"].(map[string]interface{})
	if !ok {
		j, _ := json.Marshal(body)
		err = errors.New("Unknown response from the server: " + string(j))
		return
	}

	// The alert contact ID is a string value according to API docs but is
	// returned as a integer value by the newAlertContact API JSON. In other
	// places the API does correctly handle it as a string value.
	// The difference made by it being a string is that a zero prefix to the ID // number is preserved. A zero prefixed alert contact ID is thus far only
	// been observed on the default alert contact (created at account creation).
	// https://github.com/louy/terraform-provider-uptimerobot/pull/21
	return client.GetAlertContact(fmt.Sprintf("%.0f", alertcontact["id"].(float64)))
}

func (client UptimeRobotApiClient) DeleteAlertContact(id string) (err error) {
	data := url.Values{}
	data.Add("id", id)

	_, err = client.MakeCall(
		"deleteAlertContact",
		data.Encode(),
	)
	if err != nil {
		return
	}
	return
}

type AlertContactUpdateRequest struct {
	ID           string
	FriendlyName string
	Value        string
}

func (client UptimeRobotApiClient) UpdateAlertContact(req AlertContactUpdateRequest) (err error) {
	data := url.Values{}
	data.Add("id", req.ID)
	data.Add("friendly_name", req.FriendlyName)
	data.Add("value", req.Value)

	_, err = client.MakeCall(
		"editAlertContact",
		data.Encode(),
	)
	if err != nil {
		return
	}

	return
}
