package balena

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"log"
)

func dataSourceFleet() *schema.Resource {
	return &schema.Resource{
		ReadContext: GetFleet,
		Schema: map[string]*schema.Schema{
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
		},
	}
}

func GetFleet(_ context.Context, d *schema.ResourceData, _ interface{}) diag.Diagnostics {
	if d.Get("slug").(string) != "" && d.Get("fleet_id").(int) != -1 {
		return diag.Errorf("only one of id or slug can be specified")
	} else if d.Get("slug").(string) == "" && d.Get("fleet_id").(int) == -1 {
		return diag.Errorf("either id or slug must be specified")
	}

	var endpoint string

	if d.Get("slug").(string) != "" {
		endpoint = fmt.Sprintf("/v7/application(slug='%s')", d.Get("slug").(string))
		log.Printf("[DEBUG] Retrieving Fleet with SLUG %s", d.Get("slug").(string))
	} else if d.Get("fleet_id").(int) != -1 {
		endpoint = fmt.Sprintf("/v7/application(id=%d)", d.Get("fleet_id").(int))
		log.Printf("[DEBUG] Retrieving Fleet with ID %d", d.Get("fleet_id").(int))
	}

	res, err := client.client.R().Get(endpoint)
	if err != nil {
		return diag.FromErr(err)
	}
	if res.StatusCode() != 200 {
		return diag.FromErr(fmt.Errorf("error retrieving Fleet: %s", res.Status()))
	}

	var rawResponse map[string][]map[string]interface{}
	if err := json.Unmarshal(res.Body(), &rawResponse); err != nil {
		return diag.FromErr(fmt.Errorf("failed to unmarshal response: %w", err))
	}
	responseWithoutD := rawResponse["d"]
	numberOfMatchingFleets := len(responseWithoutD)
	if numberOfMatchingFleets > 1 {
		return diag.Errorf("multiple fleets found")
	} else if numberOfMatchingFleets == 0 {
		return diag.Errorf("no matching fleets found")
	}

	flattenedResponse := flattenJSON(responseWithoutD[0])

	_ = d.Set("fleet_id", int(flattenedResponse["id"].(float64)))
	_ = d.Set("slug", flattenedResponse["slug"])
	_ = d.Set("organization_id", flattenedResponse["organization.__id"])
	_ = d.Set("app_name", flattenedResponse["app_name"])
	_ = d.Set("device_type_id", flattenedResponse["is_for__device_type.__id"])
	_ = d.Set("public", flattenedResponse["is_public"])
	_ = d.Set("host", flattenedResponse["is_host"])
	_ = d.Set("archived", flattenedResponse["is_archived"])
	_ = d.Set("created", flattenedResponse["created_at"])
	_ = d.Set("uuid", flattenedResponse["uuid"])
	_ = d.Set("track_latest_release", flattenedResponse["should_track_latest_release"])
	_ = d.Set("release_id", flattenedResponse["should_be_running__release.__id"])

	d.SetId(fmt.Sprintf("fleet:%d", int(flattenedResponse["id"].(float64))))
	return nil
}
