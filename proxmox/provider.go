package proxmox

import (
	"context"
	"os"

	"github.com/awlsring/terraform-provider-proxmox/internal/service"
	"github.com/awlsring/terraform-provider-proxmox/proxmox/network/bonds"
	"github.com/awlsring/terraform-provider-proxmox/proxmox/network/bridges"
	"github.com/awlsring/terraform-provider-proxmox/proxmox/nodes"
	"github.com/awlsring/terraform-provider-proxmox/proxmox/pools"
	"github.com/awlsring/terraform-provider-proxmox/proxmox/qemu/templates"
	"github.com/awlsring/terraform-provider-proxmox/proxmox/qemu/vms"
	"github.com/awlsring/terraform-provider-proxmox/proxmox/storage"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ provider.Provider = &ProxmoxProvider{}
)

type ProxmoxProvider struct{}

type ProxmoxProviderConfig struct {
	User     types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
	ApiKey   types.String `tfsdk:"api_key"`
	Endpoint types.String `tfsdk:"endpoint"`
	Insecure types.Bool   `tfsdk:"insecure"`
}

func New() provider.Provider {
	return &ProxmoxProvider{}
}

// Metadata returns the provider type name.
func (p *ProxmoxProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "proxmox"
}

func (p *ProxmoxProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"username": schema.StringAttribute{
				Optional:    true,
				Description: "The username to use for authentication.",
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.Expressions{
						path.MatchRoot("api_key"),
					}...),
					stringvalidator.AlsoRequires(path.Expressions{
						path.MatchRoot("password"),
					}...),
				},
			},
			"password": schema.StringAttribute{
				Optional:    true,
				Description: "Password for specified user.",
				Sensitive:   true,
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.Expressions{
						path.MatchRoot("api_key"),
					}...),
					stringvalidator.AlsoRequires(path.Expressions{
						path.MatchRoot("password"),
					}...),
				},
			},
			"api_key": schema.StringAttribute{
				Optional:    true,
				Description: "A proxmox api key.",
				Sensitive:   true,
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.Expressions{
						path.MatchRoot("username"),
					}...),
				},
			},
			"endpoint": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Proxmox endpoint to connect with. **Ex `https://10.0.0.2:8006`**",
			},
			"insecure": schema.BoolAttribute{
				Optional:    true,
				Description: "Skip TLS verification. Defaults to true.",
			},
		},
	}
}

func (p *ProxmoxProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	tflog.Info(ctx, "Configuring Prxoxmox client")

	var cfg ProxmoxProviderConfig
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpoint := os.Getenv("PROXMOX_ENDPOINT")
	user := os.Getenv("PROXMOX_USERNAME")
	password := os.Getenv("PROXMOX_PASSWORD")
	apiKey := os.Getenv("PROXMOX_API_KEY")
	insecure := true

	if !cfg.Endpoint.IsNull() {
		endpoint = cfg.Endpoint.ValueString()
	}

	if !cfg.User.IsNull() {
		user = cfg.User.ValueString()
	}

	if !cfg.Password.IsNull() {
		password = cfg.Password.ValueString()
	}

	if !cfg.ApiKey.IsNull() {
		apiKey = cfg.ApiKey.ValueString()
	}

	if !cfg.Insecure.IsNull() {
		insecure = true
	}

	if endpoint == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("endpoint"),
			"Proxmox endpoint",
			"The provider cannot create the Proxmox API client as there is a missing or empty value for the endpoint. "+
				"Set the endpoint value in the configuration or use the PROXMOX_ENDPOINT environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	scfg := service.ClientConfig{
		Username:   user,
		Password:   password,
		Token:      apiKey,
		Endpoint:   endpoint,
		SkipVerify: insecure,
	}

	ctx = tflog.SetField(ctx, "proxmox_endpoint", endpoint)
	ctx = tflog.SetField(ctx, "proxmox_username", user)
	ctx = tflog.MaskFieldValuesWithFieldKeys(ctx, "proxmox_password")
	ctx = tflog.MaskFieldValuesWithFieldKeys(ctx, "proxmox_api_key")

	tflog.Debug(ctx, "Creating Proxmox client")
	service, err := service.New(scfg)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Proxmox API Client",
			"An unexpected error occurred when creating the Proxmox API client. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"Proxmox Client Error: "+err.Error(),
		)
		return
	}

	resp.DataSourceData = service
	resp.ResourceData = service

	tflog.Info(ctx, "Configured Proxmox client", map[string]any{"success": true})
}

func (p *ProxmoxProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		pools.Resource,
		bonds.Resource,
		bridges.Resource,
	}
}

func (p *ProxmoxProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		nodes.DataSource,
		bonds.DataSource,
		bridges.DataSource,
		pools.DataSource,
		storage.DataSource,
		vms.DataSource,
		templates.DataSource,
	}

}
