package templates

import (
	"context"
	"fmt"

	"github.com/awlsring/terraform-provider-proxmox/internal/service"
	"github.com/awlsring/terraform-provider-proxmox/proxmox/filters"
	qt "github.com/awlsring/terraform-provider-proxmox/proxmox/qemu/types"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ datasource.DataSource              = &templatesDataSource{}
	_ datasource.DataSourceWithConfigure = &templatesDataSource{}
)

func DataSourceMulti() datasource.DataSource {
	return &templatesDataSource{}
}

type templatesDataSource struct {
	client *service.Proxmox
}

type templatesDataSourceModel struct {
	Templates qt.VirtualMachineDataSourceSetValue `tfsdk:"templates"`
	Filters   []filters.FilterModel               `tfsdk:"filters"`
}

func (d *templatesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_templates"
}

func (d *templatesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*service.Proxmox)
}

var filter = filters.FilterConfig{"node"}

func (d *templatesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"filters":   filter.Schema(),
			"templates": TemplateMultiDataSourceSchema,
		},
	}
}

func (d *templatesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state templatesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	nodes := filters.DetermineNode(d.client, state.Filters)

	for _, node := range nodes {
		vms, err := d.client.DescribeTemplates(ctx, node)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to get templates",
				"An error was encountered retrieving templates.\n"+
					err.Error(),
			)
			return
		}

		vmModels := []qt.VirtualMachineDataSourceModel{}
		for _, vm := range vms {
			tflog.Debug(ctx, fmt.Sprintf("Converting template %v to model", vm.VmId))
			model := qt.VMToModel(ctx, vm)
			vmModels = append(vmModels, *model)
		}

		state.Templates, err = qt.VirtualMachineDataSourceSetValueFrom(ctx, vmModels)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to convert virtual machines to state",
				"An error was encountered converting virtual machines to state.\n"+
					err.Error(),
			)
			return
		}
	}

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
