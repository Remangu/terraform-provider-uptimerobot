package provider

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	uptimerobotapi "github.com/Revolgy-Business-Solutions/terraform-provider-uptimerobot/internal/provider/api"
)

func resourceStatusPage() *schema.Resource {
	return &schema.Resource{
		Create: resourceStatusPageCreate,
		Read:   resourceStatusPageRead,
		Update: resourceStatusPageUpdate,
		Delete: resourceStatusPageDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"friendly_name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"custom_domain": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  "",
			},
			"password": &schema.Schema{
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
			},
			"sort": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(uptimerobotapi.StatusPageSort, false),
				Default:      "a-z",
			},
			"status": &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(uptimerobotapi.StatusPageStatus, false),
				Default:      "active",
			},
			"monitors": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					if k == "monitors.#" && old == "1" && new == "0" && d.Get("monitors.0").(int) == 0 {
						return true
					}
					return false
				},
			},
			"dns_address": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"standard_url": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"custom_url": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceStatusPageCreate(d *schema.ResourceData, m interface{}) error {
	rawMonitors := d.Get("monitors").([]interface{})
	monitors := make([]int, len(rawMonitors))
	for i := range rawMonitors {
		monitors[i] = rawMonitors[i].(int)
	}

	sp, err := m.(uptimerobotapi.UptimeRobotApiClient).CreateStatusPage(uptimerobotapi.StatusPageCreateRequest{
		FriendlyName: d.Get("friendly_name").(string),
		CustomDomain: d.Get("custom_domain").(string),
		Password:     d.Get("password").(string),
		Monitors:     monitors,
		Sort:         d.Get("sort").(string),
		Status:       d.Get("status").(string),
	})
	if err != nil {
		return err
	}

	d.SetId(fmt.Sprintf("%d", sp.ID))
	return updateStatusPageResource(d, sp)
}

func resourceStatusPageRead(d *schema.ResourceData, m interface{}) error {
	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return err
	}

	sp, err := m.(uptimerobotapi.UptimeRobotApiClient).GetStatusPage(id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			d.SetId("")
			return nil
		}
		return err
	}

	return updateStatusPageResource(d, sp)
}

func resourceStatusPageUpdate(d *schema.ResourceData, m interface{}) error {
	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return err
	}

	rawMonitors := d.Get("monitors").([]interface{})
	monitors := make([]int, len(rawMonitors))
	for i := range rawMonitors {
		monitors[i] = rawMonitors[i].(int)
	}

	sp, err := m.(uptimerobotapi.UptimeRobotApiClient).UpdateStatusPage(uptimerobotapi.StatusPageUpdateRequest{
		ID:           id,
		FriendlyName: d.Get("friendly_name").(string),
		CustomDomain: d.Get("custom_domain").(string),
		Password:     d.Get("password").(string),
		Monitors:     monitors,
		Sort:         d.Get("sort").(string),
		Status:       d.Get("status").(string),
	})
	if err != nil {
		return err
	}

	return updateStatusPageResource(d, sp)
}

func resourceStatusPageDelete(d *schema.ResourceData, m interface{}) error {
	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return err
	}

	err = m.(uptimerobotapi.UptimeRobotApiClient).DeleteStatusPage(id)
	if err != nil {
		return err
	}

	return nil
}
func updateStatusPageResource(d *schema.ResourceData, sp uptimerobotapi.StatusPage) error {
	if err := d.Set("friendly_name", sp.FriendlyName); err != nil {
		return fmt.Errorf("error setting friendly_name: %s", err)
	}
	if err := d.Set("standard_url", sp.StandardURL); err != nil {
		return fmt.Errorf("error setting standard_url: %s", err)
	}
	if err := d.Set("custom_url", sp.CustomURL); err != nil {
		return fmt.Errorf("error setting custom_url: %s", err)
	}
	if err := d.Set("sort", sp.Sort); err != nil {
		return fmt.Errorf("error setting sort: %s", err)
	}
	if err := d.Set("status", sp.Status); err != nil {
		return fmt.Errorf("error setting status: %s", err)
	}
	if err := d.Set("dns_address", sp.DNSAddress); err != nil {
		return fmt.Errorf("error setting dns_address: %s", err)
	}
	if err := d.Set("monitors", sp.Monitors); err != nil {
		return fmt.Errorf("error setting monitors: %s", err)
	}
	return nil
}
