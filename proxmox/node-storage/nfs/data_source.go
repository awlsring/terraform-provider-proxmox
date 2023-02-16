package nfs

import (
	"context"
	"fmt"

	"github.com/awlsring/terraform-provider-proxmox/internal/service"
	"github.com/awlsring/terraform-provider-proxmox/proxmox/filters"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ datasource.DataSource              = &nodeStorageNfsDataSource{}
	_ datasource.DataSourceWithConfigure = &nodeStorageNfsDataSource{}
)

func DataSource() datasource.DataSource {
	return &nodeStorageNfsDataSource{}
}

type nodeStorageNfsDataSource struct {
	client *service.Proxmox
}

func (d *nodeStorageNfsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_node_storage_nfs"
}

func (d *nodeStorageNfsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*service.Proxmox)
}

func (d *nodeStorageNfsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dataSourceSchema
}

func (d *nodeStorageNfsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state nodeStorageNfsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	nodes := filters.DetermineNode(d.client, state.Filters)

	for _, node := range nodes {
		tflog.Debug(ctx, fmt.Sprintf("Listing nfs storage on node %s", node))
		s, err := d.client.DescribeNFSNodeStorage(context.Background(), node)
		if err != nil {
			resp.Diagnostics.AddError(
				fmt.Sprintf("Unable to list nfs storage on node %s", node),
				"An error was encountered retrieving storage.\n"+
					err.Error(),
			)
			return
		}

		for _, n := range s {
			state.NFS = append(state.NFS, NFSToModel(n))
		}
	}

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
