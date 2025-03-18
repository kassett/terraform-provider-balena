package balena

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type Service struct {
	Name    string `json:"service_name"`
	Id      int    `json:"id"`
	Created string `json:"created_at"`
}

type ServicesResponse struct {
	Services []Service `json:"d"`
}

func dataSourceServices() *schema.Resource {
	return &schema.Resource{
		ReadContext: GetServicesDataSource,
		Schema:      getServicesDataSourceSchema(),
	}
}
func getServicesDataSourceSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"fleet_id": {
			Type:     schema.TypeInt,
			Required: true,
		},
		"services": {
			Type:     schema.TypeList,
			Computed: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"name": {
						Type:     schema.TypeString,
						Computed: true,
					},
					"service_id": {
						Type:     schema.TypeInt,
						Computed: true,
					},
					"created": {
						Type:     schema.TypeString,
						Computed: true,
					},
				},
			},
		},
	}
}

func DescribeServices(fleetId int) ([]Service, diag.Diagnostics) {
	fleet, err := DescribeFleet("", fleetId)
	if err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf("/v7/service?$filter=%s", fmt.Sprintf("application/app_name eq '%s'", fleet.AppName))
	res, _ := client.client.R().Get(endpoint)
	if !is200Level(res.StatusCode()) {
		return nil, diag.FromErr(fmt.Errorf("error retrieving Services: %s", res.Status()))
	}

	var servicesResponse ServicesResponse
	if err := json.Unmarshal(res.Body(), &servicesResponse); err != nil {
		return nil, diag.FromErr(fmt.Errorf("failed to unmarshal response from Balena services API: %w", err))
	}

	return servicesResponse.Services, nil
}

func GetServicesDataSource(_ context.Context, d *schema.ResourceData, _ interface{}) diag.Diagnostics {
	services, err := DescribeServices(d.Get("fleet_id").(int))
	if err != nil {
		return err
	}

	for dataSourceAttribute := range getServicesDataSourceSchema() {
		switch dataSourceAttribute {
		case "fleet_id":
			_ = d.Set("fleet_id", d.Get("fleet_id").(int))
		case "services":
			response := make([]map[string]interface{}, 0)
			for _, service := range services {
				response = append(response, map[string]interface{}{
					"name":       service.Name,
					"service_id": service.Id,
					"created":    service.Created,
				})
			}
			if err := d.Set("services", response); err != nil {
				return diag.Errorf("failed to set services: %v", err)
			}
		default:
			return diag.Errorf("unhandled data source attribute: %s", dataSourceAttribute)
		}
	}

	d.SetId(fmt.Sprintf("services:%d", d.Get("fleet_id").(int)))
	return err
}
