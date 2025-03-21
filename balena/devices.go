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

func GetDeviceId(deviceUuid string) string {
	return fmt.Sprintf("device:%s", deviceUuid)
}

func getDeviceDataSourceSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"uuid": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The UUID of the device. This value is unique across all fleets.",
		},
		"device_name": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The display name of the device. This value is unique within a fleet -- the device is only aware of the display name when bootstrapped.",
		},
		"last_vpn_event": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The last vpn event of the device",
		},
		"last_connectivity_event": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The last connectivity event of the device",
		},
		"ip_address": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The IP address of the device on the local area network.",
		},
		"mac_addresses": {
			Type: schema.TypeList,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
			Computed:    true,
			Description: "The MAC addresses of the device.",
		},
		"public_ip_address": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The public IP address of the network.",
		},
		"supervisor_version": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The supervisor version on the device.",
		},
		"os_version": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The OS version on the device.",
		},
		"longitude": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The longitude of the device. If the device is using a proxy, the longitude will be of the proxy.",
		},
		"latitude": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The longitude of the device. If the device is using a proxy, the latitude will be of the proxy.",
		},
		"custom_longitude": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The custom longitude of the device. This will return null if never set.",
		},
		"custom_latitude": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The custom latitude of the device. This will return null if never set.",
		},
		"device_type_id": {
			Type:        schema.TypeInt,
			Computed:    true,
			Description: "The ID of the device type. These IDs can be retrieved via the `device_type` API.",
		},
		"fleet_id": {
			Type:        schema.TypeInt,
			Computed:    true,
			Description: "The ID of the fleet. More information on the fleet can be found in the `balena_fleet` data source.",
		},
		"description": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The description of the device, representing the `note` field returned by the API.",
		},
		"created": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The time the device was created, represented as a string in ISO-Format.",
		},
		"running_release_id": {
			Type:        schema.TypeInt,
			Computed:    true,
			Description: "The ID of the running release of the device.",
		},
		"pinned_release_id": {
			Type:        schema.TypeInt,
			Computed:    true,
			Description: "The ID of the pinned release of the device. If the device is tracking latest, this ID will be null.",
		},
	}
}

func FetchDevice(uuid string) (*Device, diag.Diagnostics) {
	endpoint := fmt.Sprintf("/v7/device(uuid='%s')", uuid)
	res, err := client.client.R().Get(endpoint)
	if err != nil {
		return nil, diag.FromErr(err)
	}
	if !is200Level(res.StatusCode()) {
		return nil, diag.FromErr(fmt.Errorf("error retrieving Device: %d", res.StatusCode()))
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
		Description: "This data source provides information about a device given its unique UUID.",
	}
}

func GetDeviceDataSource(_ context.Context, d *schema.ResourceData, _ interface{}) diag.Diagnostics {
	device, err := FetchDevice(d.Get("uuid").(string))
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
		case "os_version":
			_ = d.Set("os_version", device.OsVersion)
		case "supervisor_version":
			_ = d.Set("supervisor_version", device.SupervisorVersion)
		case "public_ip_address":
			_ = d.Set("public_ip_address", device.PublicAddress)
		case "description":
			_ = d.Set("description", device.Description)
		case "fleet_id":
			_ = d.Set("fleet_id", device.FleetId.ID)
		case "running_release_id":
			_ = d.Set("running_release_id", device.RunningReleaseId)
		case "device_type_id":
			_ = d.Set("device_type_id", device.DeviceTypeId.ID)
		case "pinned_release_id":
			_ = d.Set("pinned_release_id", device.PinnedReleaseId)
		default:
			return diag.Errorf("unhandled data source attribute: %s", dataSourceAttribute)
		}
	}

	d.SetId(GetDeviceId(device.Uuid))
	return nil
}
