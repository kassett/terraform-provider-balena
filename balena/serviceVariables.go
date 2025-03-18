package balena

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type ServiceVariable struct {
	Id      int    `json:"id"`
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
	if !is200Level(res.StatusCode()) {
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

func resourceServiceVariable() *schema.Resource {
	return privateServiceVariableResource(false)
}

func resourceServiceVariableSensitive() *schema.Resource {
	return privateServiceVariableResource(true)
}

func privateServiceVariableResource(sensitive bool) *schema.Resource {
	return &schema.Resource{
		CreateContext: ResourceServiceVariableCreate,
		UpdateContext: ResourceServiceVariableUpdate,
		ReadContext:   GetServiceVariableDataSource,
		DeleteContext: ResourceServiceVariableDelete,
		Schema: map[string]*schema.Schema{
			"service_id": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
			"variable_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"value": {
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: sensitive,
			},
		},
	}
}

func ResourceServiceVariableCreate(_ context.Context, d *schema.ResourceData, _ interface{}) diag.Diagnostics {
	serviceId := d.Get("service_id").(int)
	variableName := d.Get("variable_name").(string)
	variableValue := d.Get("value").(string)

	serviceVariables, err := DescribeServiceVariables(serviceId)
	if err != nil {
		return err
	}

	for _, variable := range serviceVariables {
		if variableName == variable.Name {
			return diag.Errorf("variable %s already exists", variableName)
		}
	}

	err = CreateServiceVariable(serviceId, variableName, variableValue)
	if err != nil {
		return err
	}

	d.SetId(fmt.Sprintf("serviceVariable:%d:%s", serviceId, variableName))
	return nil
}

func ResourceServiceVariableUpdate(_ context.Context, d *schema.ResourceData, _ interface{}) diag.Diagnostics {
	serviceId := d.Get("service_id").(int)
	variableName := d.Get("variable_name").(string)
	variableValue := d.Get("value").(string)
	found := false

	serviceVariables, err := DescribeServiceVariables(serviceId)
	if err != nil {
		return err
	}

	for _, variable := range serviceVariables {
		if variableName == variable.Name {
			found = true
			err = UpdateServiceVariable(variable.Id, variableValue)
			if err != nil {
				return err
			}
			break
		}
	}

	if !found {
		return diag.Errorf("no variable %s configured for the service %d", variableName, serviceId)
	}

	return nil
}

func ResourceServiceVariableDelete(_ context.Context, d *schema.ResourceData, _ interface{}) diag.Diagnostics {
	serviceId := d.Get("service_id").(int)
	variableName := d.Get("variable_name").(string)
	found := false

	serviceVariables, err := DescribeServiceVariables(serviceId)
	if err != nil {
		return err
	}

	for _, variable := range serviceVariables {
		if variableName == variable.Name {
			found = true
			err = DeleteServiceVariable(variable.Id)
			if err != nil {
				return err
			}
			break
		}
	}

	if !found {
		return diag.Errorf("no variable %s configured for the service %d", variableName, serviceId)
	}

	return nil

}

func CreateServiceVariable(serviceId int, variableName string, variableValue string) diag.Diagnostics {
	res, err := client.client.R().
		SetBody(map[string]interface{}{
			"service": serviceId,
			"name":    variableName,
			"value":   variableValue,
		}).
		Post("/v7/service_environment_variable")

	if err != nil {
		return diag.FromErr(err)
	}

	if !is200Level(res.StatusCode()) {
		return diag.FromErr(fmt.Errorf("error creating service variable with statuscode %d", res.StatusCode()))
	}
	return nil
}

func UpdateServiceVariable(serviceVariableId int, variableValue string) diag.Diagnostics {
	res, err := client.client.R().
		SetBody(map[string]interface{}{
			"value": variableValue,
		}).
		Patch(fmt.Sprintf("/v7/service_environment_variable(%d)", serviceVariableId))
	if err != nil {
		return diag.FromErr(err)
	}

	if !is200Level(res.StatusCode()) {
		return diag.FromErr(fmt.Errorf("error updating service variable with statuscode %d", res.StatusCode()))
	}

	return nil
}

func DeleteServiceVariable(serviceVariableId int) diag.Diagnostics {
	res, err := client.client.R().Delete(fmt.Sprintf("/v7/service_environment_variable(%d)", serviceVariableId))
	if err != nil {
		return diag.FromErr(err)
	}

	if !is200Level(res.StatusCode()) {
		return diag.FromErr(fmt.Errorf("error deleting service variable with statuscode %d", res.StatusCode()))
	}
	return nil
}
