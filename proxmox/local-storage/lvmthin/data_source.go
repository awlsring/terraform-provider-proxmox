package lvmthin

import (
	"context"

	"github.com/awlsring/terraform-provider-proxmox/internal/service"
	"github.com/awlsring/terraform-provider-proxmox/proxmox/filters"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

var (
	_ datasource.DataSource              = &lvmThinpoolsDataSource{}
	_ datasource.DataSourceWithConfigure = &lvmThinpoolsDataSource{}
)

func DataSource() datasource.DataSource {
	return &lvmThinpoolsDataSource{}
}

type lvmThinpoolsDataSource struct {
	client *service.Proxmox
}

func (d *lvmThinpoolsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_lvm_thinpools"
}

func (d *lvmThinpoolsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*service.Proxmox)
}

func (d *lvmThinpoolsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dataSourceSchema
}

func (d *lvmThinpoolsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state lvmThinpoolDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	nodes := filters.DetermineNode(d.client, state.Filters)

	for _, node := range nodes {
		pools, err := d.client.ListLVMThinpools(ctx, node)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to get lvm thinpools",
				"An error was encountered retrieving lvm thinpools.\n"+
					err.Error(),
			)
			return
		}

		for _, p := range pools {
			state.LVMThinpools = append(state.LVMThinpools, LVMThinpoolToModel(&p))
		}
	}

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
