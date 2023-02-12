package bonds

import (
	"context"

	"github.com/awlsring/terraform-provider-proxmox/internal/service"
	"github.com/awlsring/terraform-provider-proxmox/proxmox/filters"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
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

func (d *bondsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dataSourceSchema
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
