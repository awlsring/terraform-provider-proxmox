package zfs

import (
	"context"

	"github.com/awlsring/terraform-provider-proxmox/internal/service"
	"github.com/awlsring/terraform-provider-proxmox/proxmox/filters"
	"github.com/awlsring/terraform-provider-proxmox/proxmox/utils"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &zfsDataSource{}
	_ datasource.DataSourceWithConfigure = &zfsDataSource{}
)

func DataSource() datasource.DataSource {
	return &zfsDataSource{}
}

type zfsDataSource struct {
	client *service.Proxmox
}

func (d *zfsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_zfs_pools"
}

func (d *zfsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*service.Proxmox)
}

func (d *zfsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dataSourceSchema
}

func (d *zfsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state zfsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	nodes := filters.DetermineNode(d.client, state.Filters)

	for _, node := range nodes {

		zfsPools, err := d.client.DescribeZFSPools(ctx, node)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to get zfs pools",
				"An error was encountered retrieving zfs pools.\n"+
					err.Error(),
			)
			return
		}

		for _, p := range zfsPools {
			pool := zfsModel{
				ID:     types.StringValue(utils.FormId(node, p.Name)),
				Node:   types.StringValue(node),
				Name:   types.StringValue(p.Name),
				Health: types.StringValue(p.Health),
				Size:   types.Int64Value(p.Size),
			}

			for _, v := range p.Disks {
				pool.Disks = append(pool.Disks, types.StringValue(v))
			}

			state.ZFSPools = append(state.ZFSPools, pool)
		}
	}

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
