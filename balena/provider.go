package balena

import (
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"os"
	"path/filepath"
	"strings"
)

func getBalenaTokenDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".balena", "token")
}

type APIClient struct {
	client *resty.Client
}

// NewAPIClient initializes a new API client with a global header
func NewAPIClient(baseURL, headerKey, headerValue string) *APIClient {
	client := resty.New().
		SetBaseURL(baseURL).
		SetHeader(headerKey, headerValue) // Set global header

	return &APIClient{client: client}
}

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"balena_token_path": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  getBalenaTokenDir(),
				ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
					path := v.(string)
					if path == "" {
						errors = append(errors, fmt.Errorf("`balena_token_path` must not be an empty string"))
					}
					_, err := os.Stat(path)
					if err != nil {
						errors = append(errors, fmt.Errorf("the `balena_token_path` %s does not exist", path))
					}
					return
				},
			},
			"balena_url": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "https://api.balena-cloud.com/",
				ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
					url := v.(string)
					if url == "" {
						errors = append(errors, fmt.Errorf("`balena_url` must not be an empty string"))
					}
					if !strings.HasPrefix(url, "https://") {
						errors = append(errors, fmt.Errorf("`balena_url` must start with `https://`"))
					}
					return
				},
			},
			"use_env_var": {
				Type:        schema.TypeBool,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("BALENA_API_KEY", nil),
			},
		},
	}
}
