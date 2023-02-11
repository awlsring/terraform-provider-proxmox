package bonds

import (
	"context"

	"github.com/awlsring/terraform-provider-proxmox/internal/service"
	"github.com/awlsring/terraform-provider-proxmox/proxmox/filters"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &bondsDataSource{}
	_ datasource.DataSourceWithConfigure = &bondsDataSource{}
)

func NewDataSource() datasource.DataSource {
	return &bondsDataSource{}
}

type bondsDataSource struct {
	client *service.Proxmox
}

func (d *bondsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_network_bonds"
}

func (d *bondsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*service.Proxmox)
}

var filter = filters.FilterConfig{"node"}

func (d *bondsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"filters": filter.Schema(),
			"network_bonds": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:    true,
							Description: "The id of the bond. Formatted as `/{node}/{name}`.",
						},
						"node": schema.StringAttribute{
							Computed:    true,
							Description: "The node the bond is on.",
						},
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "The name of the bond.",
						},
						"active": schema.BoolAttribute{
							Computed:    true,
							Description: "If the bond is active.",
						},
						"autostart": schema.BoolAttribute{
							Computed:    true,
							Description: "If the bond is set to autostart.",
						},
						"hash_policy": schema.StringAttribute{
							Computed:    true,
							Description: "Hash policy used on the bond.",
						},
						"bond_primary": schema.StringAttribute{
							Computed:    true,
							Description: "Primary interface on the bond.",
						},
						"comments": schema.StringAttribute{
							Computed:    true,
							Description: "Comments on the bond.",
						},
						"mode": schema.StringAttribute{
							Computed:    true,
							Description: "Mode of the bond.",
						},
						"mii_mon": schema.StringAttribute{
							Computed:    true,
							Description: "Miimon of the bond.",
						},
						"interfaces": schema.ListAttribute{
							Computed:    true,
							ElementType: types.StringType,
							Description: "List of interfaces on the bond.",
						},
					},
				},
			},
		},
	}
}

func (d *bondsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state bondsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	nodes := filters.DetermineNode(d.client, state.Filters)

	bonds := []service.NetworkBond{}
	for _, node := range nodes {
		b, err := d.client.DescribeNetworkBonds(ctx, node)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to get network bonds",
				"An error was encountered retrieving network bonds.\n"+
					err.Error(),
			)
			return
		}
		bonds = append(bonds, b...)
	}

	for _, bond := range bonds {
		state.Bonds = append(state.Bonds, BondToModel(&bond))
	}

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
