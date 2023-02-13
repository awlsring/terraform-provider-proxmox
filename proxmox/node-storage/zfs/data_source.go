package zfs

import (
	"context"
	"fmt"

	"github.com/awlsring/terraform-provider-proxmox/internal/service"
	"github.com/awlsring/terraform-provider-proxmox/proxmox/filters"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ datasource.DataSource              = &nodeStorageZfsDataSource{}
	_ datasource.DataSourceWithConfigure = &nodeStorageZfsDataSource{}
)

func DataSource() datasource.DataSource {
	return &nodeStorageZfsDataSource{}
}

type nodeStorageZfsDataSource struct {
	client *service.Proxmox
}

func (d *nodeStorageZfsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_node_storage_zfs"
}

func (d *nodeStorageZfsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*service.Proxmox)
}

func (d *nodeStorageZfsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dataSourceSchema
}

func (d *nodeStorageZfsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state nodeStorageZfsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	nodes := filters.DetermineNode(d.client, state.Filters)

	for _, node := range nodes {
		tflog.Debug(ctx, fmt.Sprintf("Listing zfs storage on node %s", node))
		s, err := d.client.DescribeZFSNodeStorage(context.Background(), node)
		if err != nil {
			resp.Diagnostics.AddError(
				fmt.Sprintf("Unable to list zfs storage on node %s", node),
				"An error was encountered retrieving storage.\n"+
					err.Error(),
			)
			return
		}

		for _, z := range s {
			state.ZFS = append(state.ZFS, ZFSToModel(z))
		}
	}

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
