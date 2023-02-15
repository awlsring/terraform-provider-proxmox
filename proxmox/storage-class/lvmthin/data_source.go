package lvmthin

import (
	"context"
	"fmt"

	"github.com/awlsring/terraform-provider-proxmox/internal/service"
	"github.com/awlsring/terraform-provider-proxmox/proxmox/filters"
	"github.com/awlsring/terraform-provider-proxmox/proxmox/utils"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ datasource.DataSource              = &lvmThinStorageClassDataSource{}
	_ datasource.DataSourceWithConfigure = &lvmThinStorageClassDataSource{}
)

func DataSource() datasource.DataSource {
	return &lvmThinStorageClassDataSource{}
}

type lvmThinStorageClassDataSource struct {
	client *service.Proxmox
}

func (d *lvmThinStorageClassDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_lvm_thinpool_storage_classes"
}

func (d *lvmThinStorageClassDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*service.Proxmox)
}

var filter = filters.FilterConfig{"node"}

func (d *lvmThinStorageClassDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dataSourceSchema
}

func (d *lvmThinStorageClassDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state lvmThinStorageClassDataSourceModel
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

	storage, err := d.client.ListLVMThinStorageClasses(context.Background())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to list lvm thinpool storage class",
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
			state.LVMThinpools = append(state.LVMThinpools, LVMThinStorageClassToModel(&s))
		}
	}

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
