package balena

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type DeviceTag struct {
	Id    int    `json:"id"`
	Key   string `json:"tag_key"`
	Value string `json:"value"`
}

type DeviceTagResponse struct {
	Tags []DeviceTag `json:"d"`
}

func GetDeviceTagsId(deviceUuid string) string {
	return fmt.Sprintf("device-tags:%s", deviceUuid)
}

func dataSourceDeviceTags() *schema.Resource {
	return &schema.Resource{
		ReadContext: GetDeviceTagsDataSource,
		Schema:      getDeviceTagsDataSourceSchema(),
	}
}

func getDeviceTagsDataSourceSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"device_uuid": {
			Type:     schema.TypeString,
			Required: true,
		},
		"tags": {
			Type:     schema.TypeMap,
			Computed: true,
		},
	}
}

func DescribeDeviceTags(uuid string) ([]DeviceTag, diag.Diagnostics) {
	endpoint := fmt.Sprintf("/v7/device_tag?$filter=%s", fmt.Sprintf("device/uuid eq '%s'", uuid))
	res, err := client.client.R().Get(endpoint)
	if err != nil || (res != nil && !is200Level(res.StatusCode())) {
		var message string
		if err != nil {
			message = fmt.Sprintf("with the statuscode %d", res.StatusCode())
		} else {
			message = fmt.Sprintf("with the error %s", err)
		}
		return nil, diag.Errorf("retrieving device tags for %s failed %s", uuid, message)
	}

	var deviceTagsResponse DeviceTagResponse
	if err := json.Unmarshal(res.Body(), &deviceTagsResponse); err != nil {
		return nil, diag.FromErr(err)
	}

	return deviceTagsResponse.Tags, nil
}

func GetDeviceTagsDataSource(_ context.Context, d *schema.ResourceData, _ interface{}) diag.Diagnostics {
	deviceUuid := d.Get("device_uuid").(string)

	tags, err := DescribeDeviceTags(deviceUuid)
	if err != nil {
		return err
	}

	for dataSourceAttribute := range getDeviceTagsDataSourceSchema() {
		switch dataSourceAttribute {
		case "device_uuid":
			_ = d.Set("device_uuid", deviceUuid)
		case "tags":
			newTags := map[string]string{}
			for _, deviceTag := range tags {
				newTags[deviceTag.Key] = deviceTag.Value
			}
			_ = d.Set("tags", newTags)
		default:
			return diag.Errorf("unknown data source: %s", dataSourceAttribute)
		}
	}

	d.SetId(GetDeviceTagsId(deviceUuid))
	return nil
}
