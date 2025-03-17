package balena

// TODO Figure out the correct order of hierarchy

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type DeviceVariable struct {
	Name    string `json:"name"`
	Value   string `json:"value"`
	Created string `json:"created_at"`
}

type DeviceVariablesResponse struct {
	DeviceVariables []DeviceVariable `json:"d"`
}

func dataSourceDeviceVariables() *schema.Resource {
	return &schema.Resource{
		ReadContext: GetDeviceVariablesDataSource,
		Schema:      getDeviceVariablesDataSourceSchema(),
	}
}

func getDeviceVariablesDataSourceSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"device_uuid": {
			Type:     schema.TypeString,
			Required: true,
		},
		"variables": {
			Type:     schema.TypeMap,
			Computed: true,
		},
	}
}

func DescribeDeviceVariables(deviceUuid string) ([]DeviceVariable, diag.Diagnostics) {
	endpoint := fmt.Sprintf("/v7/device_environment_variable?$filter=%s", fmt.Sprintf("device/any(d:d/uuid eq '%s')", deviceUuid))
	res, err := client.client.R().Get(endpoint)
	if err != nil {
		return nil, diag.FromErr(err)
	}
	if res.StatusCode() != 200 {
		return nil, diag.FromErr(fmt.Errorf("error retrieving Fleet Variables: %s", res.Status()))
	}

	var deviceVariables DeviceVariablesResponse
	if err := json.Unmarshal(res.Body(), &deviceVariables); err != nil {
		return nil, diag.FromErr(fmt.Errorf("failed to unmarshal response from Balena device variables API: %w", err))
	}

	return deviceVariables.DeviceVariables, nil
}

func GetDeviceVariablesDataSource(_ context.Context, d *schema.ResourceData, _ interface{}) diag.Diagnostics {
	uuid := d.Get("device_uuid").(string)
	variables, err := DescribeDeviceVariables(uuid)
	if err != nil {
		return err
	}

	for dataSourceAttribute := range getFleetVariablesDataSourceSchema() {
		switch dataSourceAttribute {
		case "deviceUuid":
			_ = d.Set("device_uuid", uuid)
		case "variables":
			variableMap := make(map[string]string)
			for _, variable := range variables {
				variableMap[variable.Name] = variable.Value
			}
			_ = d.Set("variables", variableMap)
		}
	}

	d.SetId(fmt.Sprintf("deviceVariables:%s", uuid))
	return nil
}
