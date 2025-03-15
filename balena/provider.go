package balena

import (
	"context"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"os"
	"path/filepath"
	"strings"
)

var (
	client APIClient
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
				Default:     false,
				DefaultFunc: schema.EnvDefaultFunc("BALENA_API_KEY", nil),
			},
		},

		DataSourcesMap: map[string]*schema.Resource{
			"balena_fleet": dataSourceFleet(),
		},
		ConfigureContextFunc: providerConfigure,
	}
}

func providerConfigure(_ context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	var balenaUrl = d.Get("balena_url").(string)
	var balenaUseEnvVar = d.Get("use_env_var").(bool)
	var balenaTokenPath = d.Get("balena_token_path").(string)
	fmt.Print(balenaUrl)
	fmt.Print(balenaTokenPath)

	var token string

	if balenaUseEnvVar {
		apiKey, exists := os.LookupEnv("BALENA_API_KEY")
		if !exists {
			return nil, diag.Errorf("`BALENA_API_KEY` environment variable not set")
		}
		token = apiKey
	} else {
		contents, err := os.ReadFile(balenaTokenPath)
		if err != nil {
			return nil, diag.Errorf("failed to read balena token file: %s", err)
		}
		token = string(contents)
	}

	NewAPIClient(balenaUrl, "Authorization", fmt.Sprintf("Bearer %s", token))
	return nil, nil
}
