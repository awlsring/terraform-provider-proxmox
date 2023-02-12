package bridges

import (
	"context"
	"fmt"

	"github.com/awlsring/terraform-provider-proxmox/internal/service"
	"github.com/awlsring/terraform-provider-proxmox/proxmox/filters"
	"github.com/awlsring/terraform-provider-proxmox/proxmox/network"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &bridgesDataSource{}
	_ datasource.DataSourceWithConfigure = &bridgesDataSource{}
)

func NewDataSource() datasource.DataSource {
	return &bridgesDataSource{}
}

type bridgesDataSource struct {
	client *service.Proxmox
}

func (d *bridgesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_network_bridges"
}

func (d *bridgesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*service.Proxmox)
}

func (d *bridgesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dataSourceSchema
}

func (d *bridgesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state bridgesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	nodes := filters.DetermineNode(d.client, state.Filters)

	bridges := []service.NetworkBridge{}
	for _, node := range nodes {
		b, err := d.client.DescribeNetworkBridges(ctx, node)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to get network bridges",
				"An error was encountered retrieving network bridges.\n"+
					err.Error(),
			)
			return
		}
		bridges = append(bridges, b...)
	}

	for _, bridge := range bridges {
		stateBridge := bridgeModel{
			ID:         types.StringValue(fmt.Sprintf("%s/%s", bridge.Node, bridge.Name)),
			Node:       types.StringValue(bridge.Node),
			Name:       types.StringValue(bridge.Name),
			Active:     types.BoolValue(bridge.Active),
			Autostart:  types.BoolValue(bridge.Autostart),
			VLANAware:  types.BoolValue(bridge.VLANAware),
			Interfaces: []types.String{},
		}

		if bridge.IPv4 != nil {
			stateBridge.IPv4 = &network.IpAddressModel{
				Address: types.StringValue(bridge.IPv4.Address),
				Netmask: types.StringValue(bridge.IPv4.Netmask),
			}
		}

		if bridge.IPv6 != nil {
			stateBridge.IPv6 = &network.IpAddressModel{
				Address: types.StringValue(bridge.IPv6.Address),
				Netmask: types.StringValue(bridge.IPv6.Netmask),
			}
		}

		for _, iface := range bridge.Interfaces {
			stateBridge.Interfaces = append(stateBridge.Interfaces, types.StringValue(iface))
		}

		state.Bridges = append(state.Bridges, stateBridge)
	}

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
