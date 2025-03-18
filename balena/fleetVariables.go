package balena

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type FleetVariable struct {
	Id      int    `json:"id"`
	Name    string `json:"name"`
	Value   string `json:"value"`
	Created string `json:"created_at"`
}

type FleetVariablesResponse struct {
	FleetVariables []FleetVariable `json:"d"`
}

func GetSingularFleetVariableId(fleetId int, variableName string) string {
	return fmt.Sprintf("fleet-variable:%d:%s", fleetId, variableName)
}

func GetPluralFleetVariableId(fleetId int) string {
	return fmt.Sprintf("fleet-variable:%d", fleetId)
}

func dataSourceFleetVariable() *schema.Resource {
	return &schema.Resource{
		ReadContext: GetFleetVariableDataSource,
		Schema:      getFleetVariableDataSourceSchema(false),
	}
}

func dataSourceFleetVariableSensitive() *schema.Resource {
	return &schema.Resource{
		ReadContext: GetFleetVariableDataSource,
		Schema:      getFleetVariableDataSourceSchema(true),
	}
}

func dataSourceFleetVariables() *schema.Resource {
	return &schema.Resource{
		ReadContext: GetFleetVariablesDataSource,
		Schema:      getFleetVariablesDataSourceSchema(),
	}
}

func getFleetVariableDataSourceSchema(sensitive bool) map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"fleet_id": {
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

func getFleetVariablesDataSourceSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"fleet_id": {
			Type:     schema.TypeInt,
			Required: true,
		},
		"variables": {
			Type:     schema.TypeMap,
			Computed: true,
		},
	}
}

func DescribeFleetVariables(fleetId int) ([]FleetVariable, diag.Diagnostics) {
	endpoint := fmt.Sprintf("/v7/application_environment_variable?$filter=%s", fmt.Sprintf("application eq %d", fleetId))
	res, err := client.client.R().Get(endpoint)
	if err != nil {
		return nil, diag.FromErr(err)
	}
	if !is200Level(res.StatusCode()) {
		return nil, diag.FromErr(fmt.Errorf("error retrieving Fleet Variables: %s", res.Status()))
	}

	var fleetVariables FleetVariablesResponse
	if err := json.Unmarshal(res.Body(), &fleetVariables); err != nil {
		return nil, diag.FromErr(fmt.Errorf("failed to unmarshal response from Balena fleet variables API: %w", err))
	}

	return fleetVariables.FleetVariables, nil
}

func GetFleetVariablesDataSource(_ context.Context, d *schema.ResourceData, _ interface{}) diag.Diagnostics {
	fleetId := d.Get("fleet_id").(int)
	variables, err := DescribeFleetVariables(d.Get("fleet_id").(int))
	if err != nil {
		return err
	}

	for dataSourceAttribute := range getFleetVariablesDataSourceSchema() {
		switch dataSourceAttribute {
		case "fleet_id":
			_ = d.Set("fleet_id", fleetId)
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

	d.SetId(GetPluralFleetVariableId(fleetId))
	return nil
}

func GetFleetVariableDataSource(_ context.Context, d *schema.ResourceData, _ interface{}) diag.Diagnostics {
	fleetId := d.Get("fleet_id").(int)
	variableName := d.Get("variable_name").(string)
	var variableValue string
	found := false
	variables, err := DescribeFleetVariables(fleetId)
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
		return diag.Errorf("no variable %s configured for the fleet %d", variableName, fleetId)
	}

	for dataSourceAttribute := range getFleetVariableDataSourceSchema(false) {
		switch dataSourceAttribute {
		case "fleet_id":
			_ = d.Set("fleet_id", fleetId)
		case "variable_name":
			_ = d.Set("variable_name", variableName)
		case "value":
			_ = d.Set("value", variableValue)
		default:
			return diag.Errorf("unhandled data source attribute: %s", dataSourceAttribute)
		}
	}

	d.SetId(GetSingularFleetVariableId(fleetId, variableName))
	return nil
}

func resourceFleetVariable() *schema.Resource {
	return privateFleetVariableResource(false)
}

func resourceFleetVariableSensitive() *schema.Resource {
	return privateFleetVariableResource(true)
}

func privateFleetVariableResource(sensitive bool) *schema.Resource {
	return &schema.Resource{
		CreateContext: ResourceFleetVariableCreate,
		UpdateContext: ResourceFleetVariableUpdate,
		ReadContext:   GetFleetVariableDataSource,
		DeleteContext: ResourceFleetVariableDelete,
		Schema: map[string]*schema.Schema{
			"fleet_id": {
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

func ResourceFleetVariableCreate(_ context.Context, d *schema.ResourceData, _ interface{}) diag.Diagnostics {
	fleetId := d.Get("fleet_id").(int)
	variableName := d.Get("variable_name").(string)
	variableValue := d.Get("value").(string)

	fleetVariables, err := DescribeFleetVariables(fleetId)
	if err != nil {
		return err
	}

	for _, variable := range fleetVariables {
		if variableName == variable.Name {
			return diag.Errorf("variable %s already exists", variableName)
		}
	}

	err = CreateFleetVariable(fleetId, variableName, variableValue)
	if err != nil {
		return err
	}

	d.SetId(GetSingularFleetVariableId(fleetId, variableName))
	return nil
}

func ResourceFleetVariableUpdate(_ context.Context, d *schema.ResourceData, _ interface{}) diag.Diagnostics {
	fleetId := d.Get("fleet_id").(int)
	variableName := d.Get("variable_name").(string)
	variableValue := d.Get("value").(string)
	found := false

	fleetVariables, err := DescribeFleetVariables(fleetId)
	if err != nil {
		return err
	}

	for _, variable := range fleetVariables {
		if variableName == variable.Name {
			found = true
			err = UpdateFleetVariable(variable.Id, variableValue)
			if err != nil {
				return err
			}
			break
		}
	}

	if !found {
		return diag.Errorf("no variable %s configured for the fleet %d", variableName, fleetId)
	}

	return nil
}

func ResourceFleetVariableDelete(_ context.Context, d *schema.ResourceData, _ interface{}) diag.Diagnostics {
	fleetId := d.Get("fleet_id").(int)
	variableName := d.Get("variable_name").(string)
	found := false

	fleetVariables, err := DescribeFleetVariables(fleetId)
	if err != nil {
		return err
	}

	for _, variable := range fleetVariables {
		if variableName == variable.Name {
			found = true
			err = DeleteFleetVariable(variable.Id)
			if err != nil {
				return err
			}
			break
		}
	}

	if !found {
		return diag.Errorf("no variable %s configured for the fleet %d", variableName, fleetId)
	}

	return nil

}

func CreateFleetVariable(fleetId int, variableName string, variableValue string) diag.Diagnostics {
	res, err := client.client.R().
		SetBody(map[string]interface{}{
			"application": fleetId,
			"name":        variableName,
			"value":       variableValue,
		}).
		Post("/v7/application_environment_variable")

	if err != nil {
		return diag.FromErr(err)
	}

	if !is200Level(res.StatusCode()) {
		return diag.FromErr(fmt.Errorf("error creating fleet variable with statuscode %d", res.StatusCode()))
	}
	return nil
}

func UpdateFleetVariable(fleetVariableId int, variableValue string) diag.Diagnostics {
	res, err := client.client.R().
		SetBody(map[string]interface{}{
			"value": variableValue,
		}).
		Patch(fmt.Sprintf("/v7/application_environment_variable(%d)", fleetVariableId))
	if err != nil {
		return diag.FromErr(err)
	}

	if !is200Level(res.StatusCode()) {
		return diag.FromErr(fmt.Errorf("error updating fleet variable with statuscode %d", res.StatusCode()))
	}

	return nil
}

func DeleteFleetVariable(fleetVariableId int) diag.Diagnostics {
	res, err := client.client.R().Delete(fmt.Sprintf("/v7/application_environment_variable(%d)", fleetVariableId))
	if err != nil {
		return diag.FromErr(err)
	}

	if !is200Level(res.StatusCode()) {
		return diag.FromErr(fmt.Errorf("error deleting fleet variable with statuscode %d", res.StatusCode()))
	}
	return nil
}
