package provider

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	uptimerobotapi "github.com/Revolgy-Business-Solutions/terraform-provider-uptimerobot/internal/provider/api"
)

func resourceAlertContact() *schema.Resource {
	return &schema.Resource{
		Create: resourceAlertContactCreate,
		Read:   resourceAlertContactRead,
		Update: resourceAlertContactUpdate,
		Delete: resourceAlertContactDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"friendly_name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"type": &schema.Schema{
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(uptimerobotapi.AlertContactType, false),
			},
			"value": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"status": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAlertContactCreate(d *schema.ResourceData, m interface{}) error {
	ac, err := m.(uptimerobotapi.UptimeRobotApiClient).CreateAlertContact(
		uptimerobotapi.AlertContactCreateRequest{
			FriendlyName: d.Get("friendly_name").(string),
			Type:         d.Get("type").(string),
			Value:        d.Get("value").(string),
		})
	if err != nil {
		return err
	}

	d.SetId(ac.ID)
	return updateAlertContactResource(d, ac)
}

func resourceAlertContactRead(d *schema.ResourceData, m interface{}) error {
	id := d.Id()

	ac, err := m.(uptimerobotapi.UptimeRobotApiClient).GetAlertContact(id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			d.SetId("")
			return nil
		}
		return err
	}

	return updateAlertContactResource(d, ac)
}

func resourceAlertContactUpdate(d *schema.ResourceData, m interface{}) error {
	id := d.Id()

	err := m.(uptimerobotapi.UptimeRobotApiClient).UpdateAlertContact(
		uptimerobotapi.AlertContactUpdateRequest{
			ID:           id,
			FriendlyName: d.Get("friendly_name").(string),
			Value:        d.Get("value").(string),
		})
	if err != nil {
		return err
	}

	return resourceAlertContactRead(d, m)
}

func resourceAlertContactDelete(d *schema.ResourceData, m interface{}) error {
	id := d.Id()

	err := m.(uptimerobotapi.UptimeRobotApiClient).DeleteAlertContact(id)
	if err != nil {
		return err
	}

	return nil
}

func updateAlertContactResource(d *schema.ResourceData, ac uptimerobotapi.AlertContact) error {
	if err := d.Set("friendly_name", ac.FriendlyName); err != nil {
		return fmt.Errorf("error setting friendly_name: %s", err)
	}
	if err := d.Set("value", ac.Value); err != nil {
		return fmt.Errorf("error setting value: %s", err)
	}
	if err := d.Set("type", ac.Type); err != nil {
		return fmt.Errorf("error setting type: %s", err)
	}
	if err := d.Set("status", ac.Status); err != nil {
		return fmt.Errorf("error setting status: %s", err)
	}
	return nil
}
