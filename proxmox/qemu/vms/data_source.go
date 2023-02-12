package vms

import (
	"context"
	"math/big"

	"github.com/awlsring/terraform-provider-proxmox/internal/service"
	"github.com/awlsring/terraform-provider-proxmox/internal/service/vm"
	"github.com/awlsring/terraform-provider-proxmox/proxmox/filters"
	"github.com/awlsring/terraform-provider-proxmox/proxmox/qemu"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
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
	VirtualMachines []qemu.VirtualMachineModel `tfsdk:"virtual_machines"`
	Filters         []filters.FilterModel      `tfsdk:"filters"`
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
			"virtual_machines": qemu.Schema,
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

	virtualMachines := []vm.VirtualMachine{}
	for _, node := range nodes {
		vm, err := d.client.DescribeVirtualMachines(ctx, node)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to get virtual machines",
				"An error was encountered retrieving virtual machines.\n"+
					err.Error(),
			)
			return
		}
		virtualMachines = append(virtualMachines, vm...)
	}

	for _, vm := range virtualMachines {
		stateTemplate := qemu.VirtualMachineModel{
			ID:         types.NumberValue(big.NewFloat(float64(vm.Id))),
			Node:       types.StringValue(vm.Node),
			Name:       types.StringValue(vm.Name),
			Cores:      types.NumberValue(big.NewFloat(float64(vm.Cores))),
			Memory:     types.Int64Value(vm.Memory),
			Agent:      types.BoolValue(vm.Agent),
			Tags:       []types.String{},
			Disks:      []qemu.VirtualDiskModel{},
			Interfaces: []qemu.VirtualInterfaceModel{},
		}

		for _, tag := range vm.Tags {
			stateTemplate.Tags = append(stateTemplate.Tags, types.StringValue(tag))
		}

		for _, disk := range vm.VirtualDisks {
			stateDisk := qemu.VirtualDiskModel{
				Storage:  types.StringValue(disk.Storage),
				Size:     types.Int64Value(disk.Size),
				Type:     types.StringValue(string(disk.Type)),
				Position: types.StringValue(disk.Position),
				Discard:  types.BoolValue(disk.Discard),
			}
			stateTemplate.Disks = append(stateTemplate.Disks, stateDisk)
		}

		for _, iface := range vm.VirtualNetworkDevices {
			stateInterface := qemu.VirtualInterfaceModel{
				Bridge:     types.StringValue(iface.Bridge),
				Firewall:   types.BoolValue(iface.FirewallEnabled),
				Model:      types.StringValue(string(iface.Model)),
				MacAddress: types.StringValue(iface.Mac),
				Vlan:       types.NumberValue(big.NewFloat(float64(iface.Vlan))),
				Position:   types.StringValue(iface.Position),
			}
			stateTemplate.Interfaces = append(stateTemplate.Interfaces, stateInterface)
		}

		state.VirtualMachines = append(state.VirtualMachines, stateTemplate)
	}

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
