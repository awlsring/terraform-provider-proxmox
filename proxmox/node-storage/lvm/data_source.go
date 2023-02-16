package lvmthin

import (
	"context"
	"fmt"

	"github.com/awlsring/terraform-provider-proxmox/internal/service"
	"github.com/awlsring/terraform-provider-proxmox/proxmox/filters"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ datasource.DataSource              = &nodeStorageLVMDataSource{}
	_ datasource.DataSourceWithConfigure = &nodeStorageLVMDataSource{}
)

func DataSource() datasource.DataSource {
	return &nodeStorageLVMDataSource{}
}

type nodeStorageLVMDataSource struct {
	client *service.Proxmox
}

func (d *nodeStorageLVMDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_node_storage_lvms"
}

func (d *nodeStorageLVMDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*service.Proxmox)
}

func (d *nodeStorageLVMDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dataSourceSchema
}

func (d *nodeStorageLVMDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state nodeStorageLVMDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	nodes := filters.DetermineNode(d.client, state.Filters)

	for _, node := range nodes {
		tflog.Debug(ctx, fmt.Sprintf("Listing lvm storage on node %s", node))
		s, err := d.client.DescribeLVMNodeStorage(context.Background(), node)
		if err != nil {
			resp.Diagnostics.AddError(
				fmt.Sprintf("Unable to list lvm storage on node %s", node),
				"An error was encountered retrieving storage.\n"+
					err.Error(),
			)
			return
		}

		for _, n := range s {
			state.LVMs = append(state.LVMs, LVMToModel(n))
		}
	}

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
