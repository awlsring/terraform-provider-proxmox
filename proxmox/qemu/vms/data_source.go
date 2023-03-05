package vms

import (
	"context"
	"fmt"

	"github.com/awlsring/terraform-provider-proxmox/internal/service"
	"github.com/awlsring/terraform-provider-proxmox/proxmox/filters"
	"github.com/awlsring/terraform-provider-proxmox/proxmox/qemu/schemas"
	qt "github.com/awlsring/terraform-provider-proxmox/proxmox/qemu/types"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ datasource.DataSource              = &virtualMachinesDataSource{}
	_ datasource.DataSourceWithConfigure = &virtualMachinesDataSource{}
)

func DataSource() datasource.DataSource {
	return &virtualMachinesDataSource{}
}

type virtualMachinesDataSource struct {
	client *service.Proxmox
}

type virtualMachinesDataSourceModel struct {
	VirtualMachines qt.VirtualMachineDataSourceSetValue `tfsdk:"virtual_machines"`
	Filters         []filters.FilterModel               `tfsdk:"filters"`
}

func (d *virtualMachinesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_virtual_machines"
}

func (d *virtualMachinesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*service.Proxmox)
}

var filter = filters.FilterConfig{"node"}

func (d *virtualMachinesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"filters":          filter.Schema(),
			"virtual_machines": schemas.DataSourceSchema,
		},
	}
}

func (d *virtualMachinesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state virtualMachinesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	nodes := filters.DetermineNode(d.client, state.Filters)

	for _, node := range nodes {
		vms, err := d.client.DescribeVirtualMachines(ctx, node)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to get virtual machines",
				"An error was encountered retrieving virtual machines.\n"+
					err.Error(),
			)
			return
		}

		vmModels := []qt.VirtualMachineDataSourceModel{}
		for _, vm := range vms {
			tflog.Debug(ctx, fmt.Sprintf("Converting VM %v to model", vm.VmId))
			model := qt.VMToModel(ctx, vm)
			vmModels = append(vmModels, *model)
		}

		state.VirtualMachines, err = qt.VirtualMachineDataSourceSetValueFrom(ctx, vmModels)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to convert virtual machines to state",
				"An error was encountered converting virtual machines to state.\n"+
					err.Error(),
			)
			return
		}
	}

	tflog.Debug(ctx, fmt.Sprintf("Found %v virtual machines, assigning to state", len(state.VirtualMachines.Elements())))
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "Error setting state")
		return
	}
	tflog.Debug(ctx, "Successfully read virtual machines")
}
