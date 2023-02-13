package zfs

import (
	"context"
	"fmt"

	"github.com/awlsring/terraform-provider-proxmox/internal/service"
	"github.com/awlsring/terraform-provider-proxmox/proxmox/filters"
	"github.com/awlsring/terraform-provider-proxmox/proxmox/utils"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ datasource.DataSource              = &storageClassZfsDataSource{}
	_ datasource.DataSourceWithConfigure = &storageClassZfsDataSource{}
)

func DataSource() datasource.DataSource {
	return &storageClassZfsDataSource{}
}

type storageClassZfsDataSource struct {
	client *service.Proxmox
}

func (d *storageClassZfsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_storage_class_zfs"
}

func (d *storageClassZfsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*service.Proxmox)
}

var filter = filters.FilterConfig{"node"}

func (d *storageClassZfsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dataSourceSchema
}

func (d *storageClassZfsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state storageClassZfsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	nodes := []string{}
	for _, filter := range state.Filters {
		if filter.Name.ValueString() == "node" {
			for _, v := range filter.Values {
				nodes = append(nodes, v.ValueString())
			}
		}
	}
	tflog.Debug(ctx, fmt.Sprintf("Filtering nodes %s", nodes))

	storage, err := d.client.ListZFSStorageClasses(context.Background())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to list zfs storage class",
			"An error was encountered retrieving storage.\n"+
				err.Error(),
		)
		return
	}

	for _, s := range storage {
		add := true
		if len(nodes) > 0 {
			found := false
			for _, node := range nodes {
				tflog.Debug(ctx, fmt.Sprintf("Checking if node %s is in list", node))
				if utils.ListContains(s.Nodes, node) {
					tflog.Debug(ctx, fmt.Sprintf("Node %s is in list", node))
					found = true
					break
				}
			}

			if !found {
				add = false
			}
		}

		if add {
			zfs := storageClassZfsModel{
				ID:           types.StringValue(s.Id),
				Pool:         types.StringValue(s.ZFSPool),
				ContentTypes: utils.UnpackList(s.Content),
				Nodes:        utils.UnpackList(s.Nodes),
				Mount:        types.StringValue(s.Mount),
			}
			state.ZFS = append(state.ZFS, zfs)
		}
	}

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
