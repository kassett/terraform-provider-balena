package balena

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"log"
)

func dataSourceFleet() *schema.Resource {
	return &schema.Resource{
		ReadContext: GetFleet,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:          schema.TypeInt,
				Optional:      true,
				ConflictsWith: []string{"slug"},
				Default:       -1,
			},
			"slug": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"id"},
				Default:       "",
			},
		},
	}
}

func GetFleet(ctx context.Context, d *schema.ResourceData, _ interface{}) diag.Diagnostics {
	if d.Get("slug").(string) != "" && d.Get("id").(int) != -1 {
		return diag.Errorf("only one of id or slug can be specified")
	} else if d.Get("slug").(string) == "" && d.Get("id").(int) == -1 {
		return diag.Errorf("either id or slug must be specified")
	}

	var endpoint string

	if d.Get("slug").(string) != "" {
		endpoint = fmt.Sprintf("v7/application(slug='%s')", d.Get("slug").(string))
		log.Printf("[DEBUG] Retrieving Fleet with SLUG %s", d.Get("slug").(string))
	} else if d.Get("id").(int) != -1 {
		endpoint = fmt.Sprintf("v7/application(id=%d)", d.Get("id").(int))
		log.Printf("[DEBUG] Retrieving Fleet with ID %d", d.Get("id").(int))
	}

	res, err := client.client.R().Get(endpoint)
	if err != nil {
		return diag.FromErr(err)
	}
	if res.StatusCode() != 200 {
		return diag.FromErr(fmt.Errorf("error retrieving Fleet: %s", res.Status()))
	}

	return nil
	//return diag.Errorf("error retrieving Fleet: %s", res.Body())
}
