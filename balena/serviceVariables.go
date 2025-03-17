package balena

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type ServiceVariable struct {
	Name    string `json:"name"`
	Value   string `json:"value"`
	Created string `json:"created_at"`
}

type ServiceVariableResponse struct {
	ServiceVariables []ServiceVariable `json:"d"`
}

func dataSourceServiceVariable() *schema.Resource {
	return &schema.Resource{
		ReadContext: GetServiceVariableDataSource,
		Schema:      getServiceVariableDataSourceSchema(false),
	}
}

func dataSourceServiceVariableSensitive() *schema.Resource {
	return &schema.Resource{
		ReadContext: GetServiceVariableDataSource,
		Schema:      getServiceVariableDataSourceSchema(true),
	}
}

func dataSourceServiceVariables() *schema.Resource {
	return &schema.Resource{
		ReadContext: GetServiceVariablesDataSource,
		Schema:      getServiceVariablesDataSourceSchema(),
	}
}

func getServiceVariablesDataSourceSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"service_id": {
			Type:     schema.TypeInt,
			Required: true,
		},
		"variables": {
			Type:     schema.TypeMap,
			Computed: true,
		},
	}
}

func getServiceVariableDataSourceSchema(sensitive bool) map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"service_id": {
			Type:     schema.TypeInt,
			Required: true,
		},
		"variable_name": {
			Type:     schema.TypeString,
			Required: true,
		},
		"value": {
			Type:      schema.TypeString,
			Computed:  true,
			Sensitive: sensitive,
		},
	}
}

func DescribeServiceVariables(serviceId int) ([]ServiceVariable, diag.Diagnostics) {
	endpoint := fmt.Sprintf("/v7/service_environment_variable?$filter=%s", fmt.Sprintf("service eq %d", serviceId))
	res, err := client.client.R().Get(endpoint)
	if err != nil {
		return nil, diag.FromErr(err)
	}
	if res.StatusCode() != 200 {
		return nil, diag.FromErr(fmt.Errorf("error retrieving Fleet Variables: %s", res.Status()))
	}

	var serviceVariables ServiceVariableResponse
	if err := json.Unmarshal(res.Body(), &serviceVariables); err != nil {
		return nil, diag.FromErr(fmt.Errorf("failed to unmarshal response from Balena service variables API: %w", err))
	}

	return serviceVariables.ServiceVariables, nil
}

func GetServiceVariablesDataSource(_ context.Context, d *schema.ResourceData, _ interface{}) diag.Diagnostics {
	serviceId := d.Get("service_id").(int)
	variables, err := DescribeServiceVariables(serviceId)
	if err != nil {
		return err
	}

	for dataSourceAttribute := range getServiceVariablesDataSourceSchema() {
		switch dataSourceAttribute {
		case "service_id":
			_ = d.Set("service_id", serviceId)
		case "variables":
			variableMap := make(map[string]string)
			for _, variable := range variables {
				variableMap[variable.Name] = variable.Value
			}
			_ = d.Set("variables", variableMap)
		default:
			return diag.Errorf("unhandled data source attribute: %s", dataSourceAttribute)
		}
	}

	d.SetId(fmt.Sprintf("serviceVariables:%d", serviceId))
	return nil
}

func GetServiceVariableDataSource(_ context.Context, d *schema.ResourceData, _ interface{}) diag.Diagnostics {
	serviceId := d.Get("service_id").(int)
	variableName := d.Get("variable_name").(string)
	var variableValue string
	found := false
	variables, err := DescribeServiceVariables(serviceId)
	if err != nil {
		return err
	}

	for _, variable := range variables {
		if variableName == variable.Name {
			variableValue = variable.Value
			found = true
			break
		}
	}

	if !found {
		return diag.Errorf("no variable %s configured for the service %d", variableName, serviceId)
	}

	for dataSourceAttribute := range getServiceVariableDataSourceSchema(false) {
		switch dataSourceAttribute {
		case "service_id":
			_ = d.Set("service_id", serviceId)
		case "variable_name":
			_ = d.Set("variable_name", variableName)
		case "value":
			_ = d.Set("value", variableValue)
		default:
			return diag.Errorf("unhandled data source attribute: %s", dataSourceAttribute)
		}
	}

	d.SetId(fmt.Sprintf("serviceVariable:%d:%s", serviceId, variableName))
	return nil
}
