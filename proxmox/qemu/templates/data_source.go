package templates

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
	_ datasource.DataSource              = &templatesDataSource{}
	_ datasource.DataSourceWithConfigure = &templatesDataSource{}
)

func DataSource() datasource.DataSource {
	return &templatesDataSource{}
}

type templatesDataSource struct {
	client *service.Proxmox
}

type templatesDataSourceModel struct {
	Templates []qemu.VirtualMachineModel `tfsdk:"templates"`
	Filters   []filters.FilterModel      `tfsdk:"filters"`
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
			"templates": qemu.Schema,
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

	templates := []vm.VirtualMachine{}
	for _, node := range nodes {
		t, err := d.client.DescribeTemplates(ctx, node)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to get templates",
				"An error was encountered retrieving templates.\n"+
					err.Error(),
			)
			return
		}
		templates = append(templates, t...)
	}

	for _, t := range templates {
		stateTemplate := qemu.VirtualMachineModel{
			ID:         types.NumberValue(big.NewFloat(float64(t.Id))),
			Node:       types.StringValue(t.Node),
			Name:       types.StringValue(t.Name),
			Cores:      types.NumberValue(big.NewFloat(float64(t.Cores))),
			Memory:     types.Int64Value(t.Memory),
			Agent:      types.BoolValue(t.Agent),
			Tags:       []types.String{},
			Disks:      []qemu.VirtualDiskModel{},
			Interfaces: []qemu.VirtualInterfaceModel{},
		}

		for _, tag := range t.Tags {
			stateTemplate.Tags = append(stateTemplate.Tags, types.StringValue(tag))
		}

		for _, disk := range t.VirtualDisks {
			stateDisk := qemu.VirtualDiskModel{
				Storage:  types.StringValue(disk.Storage),
				Size:     types.Int64Value(disk.Size),
				Type:     types.StringValue(string(disk.Type)),
				Position: types.StringValue(disk.Position),
				Discard:  types.BoolValue(disk.Discard),
			}
			stateTemplate.Disks = append(stateTemplate.Disks, stateDisk)
		}

		for _, iface := range t.VirtualNetworkDevices {
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

		state.Templates = append(state.Templates, stateTemplate)
	}

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
