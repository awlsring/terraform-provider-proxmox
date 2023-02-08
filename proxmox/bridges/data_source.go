package bridges

import (
	"context"
	"fmt"

	"github.com/awlsring/terraform-provider-proxmox/internal/service"
	"github.com/awlsring/terraform-provider-proxmox/proxmox/filters"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
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

type bridgesDataSourceModel struct {
	Bridges []bridgeModel         `tfsdk:"network_bridges"`
	Filters []filters.FilterModel `tfsdk:"filters"`
}

type bridgeModel struct {
	ID         types.String   `tfsdk:"id"`
	Node       types.String   `tfsdk:"node"`
	Name       types.String   `tfsdk:"name"`
	Active     types.Bool     `tfsdk:"active"`
	Autostart  types.Bool     `tfsdk:"autostart"`
	VlanAware  types.Bool     `tfsdk:"vlan_aware"`
	Interfaces []types.String `tfsdk:"interfaces"`
	IPv4       ipAddressModel `tfsdk:"ipv4"`
	IPv6       ipAddressModel `tfsdk:"ipv6"`
}

type ipAddressModel struct {
	Address types.String `tfsdk:"address"`
	Netmask types.String `tfsdk:"netmask"`
	Gateway types.String `tfsdk:"gateway"`
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

var filter = filters.FilterConfig{"node"}

func (d *bridgesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"filters": filter.Schema(),
			"network_bridges": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:    true,
							Description: "The id of the bridge. Formatted as /{node}/{name}.",
						},
						"node": schema.StringAttribute{
							Computed:    true,
							Description: "The node the bridge is on.",
						},
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "The name of the bridge.",
						},
						"active": schema.BoolAttribute{
							Computed:    true,
							Description: "If the bridge is active.",
						},
						"autostart": schema.BoolAttribute{
							Computed:    true,
							Description: "If the bridge is set to autostart.",
						},
						"vlan_aware": schema.BoolAttribute{
							Computed:    true,
							Description: "If the bridge is vlan aware.",
						},
						"interfaces": schema.ListAttribute{
							Computed:    true,
							ElementType: types.StringType,
							Description: "List of interfaces on the bridge.",
						},
						"ipv4": schema.ObjectAttribute{
							Computed:    true,
							Description: "Information of the ipv4 address.",
							AttributeTypes: map[string]attr.Type{
								"address": types.StringType,
								"netmask": types.StringType,
								"gateway": types.StringType,
							},
						},
						"ipv6": schema.ObjectAttribute{
							Computed:    true,
							Description: "Information of the ipv6 address.",
							AttributeTypes: map[string]attr.Type{
								"address": types.StringType,
								"netmask": types.StringType,
								"gateway": types.StringType,
							},
						},
					},
				},
			},
		},
	}
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
			VlanAware:  types.BoolValue(bridge.VLANAware),
			Interfaces: []types.String{},
		}

		if bridge.IPv4 != nil {
			stateBridge.IPv4 = ipAddressModel{
				Address: types.StringValue(bridge.IPv4.Address),
				Netmask: types.StringValue(bridge.IPv4.Netmask),
				Gateway: types.StringValue(bridge.IPv4.Gateway),
			}
		}

		if bridge.IPv6 != nil {
			stateBridge.IPv6 = ipAddressModel{
				Address: types.StringValue(bridge.IPv6.Address),
				Netmask: types.StringValue(bridge.IPv6.Netmask),
				Gateway: types.StringValue(bridge.IPv6.Gateway),
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
