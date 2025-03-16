package balena

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type Fleet struct {
	FleetID            int       `json:"id"`
	OrganizationID     IDWrapper `json:"organization"`
	Slug               string    `json:"slug"`
	AppName            string    `json:"app_name"`
	ReleaseId          IDWrapper `json:"should_be_running__release"`
	DeviceType         IDWrapper `json:"device_type"`
	TrackLatestRelease bool      `json:"should_track_latest_release"`
	Public             bool      `json:"is_public"`
	Host               bool      `json:"is_host"`
	Archived           bool      `json:"is_archived"`
	Created            string    `json:"created_at"`
	Uuid               string    `json:"uuid"`
}

type FleetResponse struct {
	Fleets []Fleet `json:"d"`
}

func getFleetDataSourceSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"fleet_id": {
			Type:          schema.TypeInt,
			Optional:      true,
			ConflictsWith: []string{"slug"},
			Default:       -1,
		},
		"slug": {
			Type:          schema.TypeString,
			Optional:      true,
			ConflictsWith: []string{"fleet_id"},
			Default:       "",
		},
		"device_type_id": {
			Type:     schema.TypeInt,
			Computed: true,
			Optional: true,
		},
		"app_name": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"track_latest_release": {
			Type:     schema.TypeBool,
			Computed: true,
		},
		"release_id": {
			Type:     schema.TypeInt,
			Computed: true,
		},
		"organization_id": {
			Type:     schema.TypeInt,
			Computed: true,
		},
		"public": {
			Type:     schema.TypeBool,
			Computed: true,
		},
		"host": {
			Type:     schema.TypeBool,
			Computed: true,
		},
		"archived": {
			Type:     schema.TypeBool,
			Computed: true,
		},
		"created": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"uuid": {
			Type:     schema.TypeString,
			Computed: true,
		},
	}
}

func DescribeFleet(slug string, fleetId int) (*Fleet, diag.Diagnostics) {
	var endpoint string
	if slug != "" {
		endpoint = fmt.Sprintf("/v7/application(slug='%s')", slug)
	} else {
		endpoint = fmt.Sprintf("/v7/application(id=%d)", fleetId)
	}

	res, err := client.client.R().Get(endpoint)
	if err != nil {
		return nil, diag.FromErr(err)
	}
	if res.StatusCode() != 200 {
		return nil, diag.FromErr(fmt.Errorf("error retrieving Fleet: %s", res.Status()))
	}

	var fleetResponse FleetResponse

	err = json.Unmarshal(res.Body(), &fleetResponse)

	if len(fleetResponse.Fleets) == 0 {
		return nil, diag.Errorf("no fleet found")
	} else if len(fleetResponse.Fleets) > 1 {
		return nil, diag.Errorf("more than one fleet found")
	}

	return &fleetResponse.Fleets[0], nil
}

func dataSourceFleet() *schema.Resource {
	return &schema.Resource{
		ReadContext: GetFleetDataSource,
		Schema:      getFleetDataSourceSchema(),
	}
}

func GetFleetDataSource(_ context.Context, d *schema.ResourceData, _ interface{}) diag.Diagnostics {
	if d.Get("slug").(string) != "" && d.Get("fleet_id").(int) != -1 {
		return diag.Errorf("only one of id or slug can be specified")
	} else if d.Get("slug").(string) == "" && d.Get("fleet_id").(int) == -1 {
		return diag.Errorf("either id or slug must be specified")
	}

	fleet, err := DescribeFleet(d.Get("slug").(string), d.Get("fleet_id").(int))
	if err != nil {
		return err
	}

	for dataSourceAttribute := range getFleetDataSourceSchema() {
		switch dataSourceAttribute {
		case "fleet_id":
			_ = d.Set("fleet_id", fleet.FleetID)
		case "slug":
			_ = d.Set("slug", fleet.Slug)
		case "organization_id":
			_ = d.Set("organization_id", fleet.OrganizationID.ID)
		case "app_name":
			_ = d.Set("app_name", fleet.AppName)
		case "device_type_id":
			_ = d.Set("device_type_id", fleet.DeviceType.ID)
		case "public":
			_ = d.Set("public", fleet.Public)
		case "host":
			_ = d.Set("host", fleet.Host)
		case "archived":
			_ = d.Set("archived", fleet.Archived)
		case "created":
			_ = d.Set("created", fleet.Created)
		case "track_latest_release":
			_ = d.Set("track_latest_release", fleet.TrackLatestRelease)
		case "release_id":
			_ = d.Set("release_id", fleet.ReleaseId.ID)
		default:
			return diag.Errorf("unhandled data source attribute: %s", dataSourceAttribute)
		}
	}

	d.SetId(fmt.Sprintf("fleet:%d", fleet.FleetID))
	return nil
}
