package proxmox

import (
	"context"
	"fmt"

	"github.com/awlsring/terraform-provider-proxmox/internal/service"
	"github.com/awlsring/terraform-provider-proxmox/proxmox/bridges"
	"github.com/awlsring/terraform-provider-proxmox/proxmox/nodes"
	"github.com/awlsring/terraform-provider-proxmox/proxmox/templates"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Provider() *schema.Provider {
    return &schema.Provider{
        Schema: map[string]*schema.Schema{
			"username": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("PROXMOX_USERNAME", nil),
				Description: "Username for proxmox. Ex. awlsring@pam",
			},
			"password": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("PROXMOX_PASSWORD", nil),
				Description: "Password for specified user",
				Sensitive:   true,
			},
			"api_key": {
				Type:        schema.TypeString,
				Description: "Proxmox API key",
				Optional:    true,
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc("PROXMOX_API_KEY", nil),
			},
			"endpoint": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Proxmox endpoint to connect with. Ex. https://10.0.0.2:8006",
				DefaultFunc: schema.EnvDefaultFunc("PROXMOX_ENDPOINT", nil),
			},
			"insecure": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Skip TLS verification. Defaults to true.",
				Default:    true,
			},
		},
		DataSourcesMap: map[string]*schema.Resource{
			"proxmox_nodes": nodes.DataSource(),
			"proxmox_templates": templates.DataSource(),
			"proxmox_network_bridges": bridges.DataSource(),
		},
        ConfigureContextFunc: providerConfigure,
    }
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {	
	cfg, err := formConfig(d)
	if err != nil {
		return nil, diag.FromErr(err)
	}

	service, err := service.New(cfg)
	if err != nil {
		return nil, diag.FromErr(err)
	}
	return service, nil
}

func formConfig(d *schema.ResourceData) (service.ClientConfig, error){
	username := d.Get("username").(string)
	password := d.Get("password").(string)
	apiKey := d.Get("api_key").(string)
	endpoint := d.Get("endpoint").(string)
	skipVerify := d.Get("insecure").(bool)

	if endpoint == "" {
		return service.ClientConfig{}, fmt.Errorf("endpoint is required")
	}

	return service.ClientConfig{
		Username: username,
		Password: password,
		Token: apiKey,
		Endpoint: endpoint,
		SkipVerify: skipVerify,
	}, nil
}