package lvm

import (
	"context"

	"github.com/awlsring/terraform-provider-proxmox/internal/service"
	"github.com/awlsring/terraform-provider-proxmox/proxmox/filters"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

var (
	_ datasource.DataSource              = &lvmDataSource{}
	_ datasource.DataSourceWithConfigure = &lvmDataSource{}
)

func DataSource() datasource.DataSource {
	return &lvmDataSource{}
}

type lvmDataSource struct {
	client *service.Proxmox
}

func (d *lvmDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_lvms"
}

func (d *lvmDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*service.Proxmox)
}

func (d *lvmDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dataSourceSchema
}

func (d *lvmDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state lvmDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	nodes := filters.DetermineNode(d.client, state.Filters)

	for _, node := range nodes {
		lvms, err := d.client.ListLVMs(ctx, node)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to get lvms",
				"An error was encountered retrieving lvms.\n"+
					err.Error(),
			)
			return
		}

		for _, p := range lvms {
			state.LVMs = append(state.LVMs, LVMToModel(&p))
		}
	}

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
