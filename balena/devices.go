package balena

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"log"
	"strings"
)

type Device struct {
	Uuid                  string    `json:"uuid"`
	DeviceName            string    `json:"device_name"`
	LastVpnEvent          string    `json:"last_vpn_event"`
	LastConnectivityEvent string    `json:"last_connectivity_event"`
	IpAddress             string    `json:"ip_address"`
	MacAddresses          string    `json:"mac_addresses"`
	PublicAddress         string    `json:"public_address"`
	SupervisorVersion     string    `json:"supervisor_version"`
	OsVersion             string    `json:"os_version"`
	Longitude             string    `json:"longitude"`
	Latitude              string    `json:"latitude"`
	CustomLongitude       string    `json:"custom_longitude"`
	CustomerLatitude      string    `json:"custom_latitude"`
	DeviceTypeId          IDWrapper `json:"is_of__device_type"`
	FleetId               IDWrapper `json:"belongs_to__application"`
	Description           string    `json:"note"`
	Created               string    `json:"created_at"`
	RunningReleaseId      IDWrapper `json:"is_running__release"`
	PinnedReleaseId       string    `json:"is_pinned_on__release"`
}

type DeviceResponse struct {
	Devices []Device `json:"d"`
}

func getDeviceDataSourceSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"uuid": {
			Type:     schema.TypeString,
			Required: true,
		},
		"device_name": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"last_vpn_event": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"last_connectivity_event": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"ip_address": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"mac_addresses": {
			Type: schema.TypeList,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
			Computed: true,
		},
		"public_ip_address": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"supervisor_version": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"os_version": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"longitude": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"latitude": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"custom_longitude": {
			Type:     schema.TypeString,
			Computed: true,
			Optional: true,
		},
		"custom_latitude": {
			Type:     schema.TypeString,
			Computed: true,
			Optional: true,
		},
		"device_type_id": {
			Type:     schema.TypeInt,
			Computed: true,
		},
		"fleet_id": {
			Type:     schema.TypeInt,
			Computed: true,
		},
		"description": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"created": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"running_release_id": {
			Type:     schema.TypeInt,
			Computed: true,
		},
		"pinned_release_id": {
			Type:     schema.TypeInt,
			Computed: true,
		},
	}
}

func DescribeDevice(uuid string) (*Device, diag.Diagnostics) {
	endpoint := fmt.Sprintf("/v7/device(uuid='%s')", uuid)
	res, err := client.client.R().Get(endpoint)
	if err != nil {
		return nil, diag.FromErr(err)
	}
	if res.StatusCode() != 200 {
		return nil, diag.FromErr(fmt.Errorf("error retrieving Device: %s", res.Status()))
	}

	var deviceResponse DeviceResponse
	if err := json.Unmarshal(res.Body(), &deviceResponse); err != nil {
		return nil, diag.FromErr(fmt.Errorf("failed to unmarshal response from Balena device API: %w", err))
	}

	if len(deviceResponse.Devices) == 0 {
		return nil, diag.Errorf("no device found")
	} else if len(deviceResponse.Devices) > 1 {
		return nil, diag.Errorf("more than one device found")
	}

	return &deviceResponse.Devices[0], nil
}

func dataSourceDevice() *schema.Resource {
	return &schema.Resource{
		ReadContext: GetDeviceDataSource,
		Schema:      getDeviceDataSourceSchema(),
	}
}

func GetDeviceDataSource(_ context.Context, d *schema.ResourceData, _ interface{}) diag.Diagnostics {
	device, err := DescribeDevice(d.Get("uuid").(string))
	if err != nil {
		return err
	}

	for dataSourceAttribute := range getDeviceDataSourceSchema() {
		switch dataSourceAttribute {
		case "uuid":
			log.Print(device.Uuid)
		case "device_name":
			_ = d.Set("device_name", device.DeviceName)
		case "last_vpn_event":
			_ = d.Set("last_vpn_event", device.LastVpnEvent)
		case "last_connectivity_event":
			_ = d.Set("last_connectivity_event", device.LastConnectivityEvent)
		case "ip_address":
			_ = d.Set("ip_address", device.IpAddress)
		case "mac_addresses":
			_ = d.Set("mac_addresses", strings.Split(device.MacAddresses, " "))
		case "created":
			_ = d.Set("created", device.Created)
		case "latitude":
			_ = d.Set("latitude", device.Latitude)
		case "longitude":
			_ = d.Set("longitude", device.Longitude)
		case "custom_longitude":
			_ = d.Set("custom_longitude", device.CustomLongitude)
		case "custom_latitude":
			_ = d.Set("custom_latitude", device.CustomerLatitude)
		case "description":
			_ = d.Set("description", device.Description)
		case "fleet_id":
			_ = d.Set("fleet_id", device.FleetId)
		case "running_release_id":
			_ = d.Set("running_release_id", device.RunningReleaseId)
		case "pinned_release_id":
			_ = d.Set("pinned_release_id", device.PinnedReleaseId)
		default:
			return diag.Errorf("unhandled data source attribute: %s", dataSourceAttribute)
		}
	}

	d.SetId(fmt.Sprintf("device:%s", device.Uuid))
	return nil
}
