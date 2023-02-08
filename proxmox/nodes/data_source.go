package nodes

import (
	"context"

	"github.com/awlsring/terraform-provider-proxmox/internal/service"
	"github.com/awlsring/terraform-provider-proxmox/proxmox/filters"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &nodesDataSource{}
	_ datasource.DataSourceWithConfigure = &nodesDataSource{}
)

func NewDataSource() datasource.DataSource {
	return &nodesDataSource{}
}

type nodesDataSource struct {
	client *service.Proxmox
}

type nodesDataSourceModel struct {
	Nodes   []nodeModel           `tfsdk:"nodes"`
	Filters []filters.FilterModel `tfsdk:"filters"`
}

type nodeModel struct {
	ID             types.String   `tfsdk:"id"`
	Node           types.String   `tfsdk:"node"`
	Cores          types.Int64    `tfsdk:"cores"`
	SslFingerprint types.String   `tfsdk:"ssl_fingerprint"`
	Memory         types.Int64    `tfsdk:"memory"`
	Disks          []diskModel    `tfsdk:"disks"`
	Interfaces     []types.String `tfsdk:"network_interfaces"`
}

type diskModel struct {
	Device types.String `tfsdk:"device"`
	Size   types.Int64  `tfsdk:"size"`
	Model  types.String `tfsdk:"model"`
	Serial types.String `tfsdk:"serial"`
	Vendor types.String `tfsdk:"vendor"`
	Type   types.String `tfsdk:"type"`
}

func (d *nodesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_nodes"
}

func (d *nodesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*service.Proxmox)
}

var filter = filters.FilterConfig{"node"}

func (d *nodesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"filters": filter.Schema(),
			"nodes": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:    true,
							Description: "The id of the bridge. Formatted as /{node}/{name}.",
						},
						"node": schema.StringAttribute{
							Computed:    true,
							Description: "The node name.",
						},
						"cores": schema.Int64Attribute{
							Computed:    true,
							Description: "Amount of CPU cores on the machine",
						},
						"ssl_fingerprint": schema.StringAttribute{
							Computed:    true,
							Description: "The SSL fingerprint of the node",
						},
						"memory": schema.Int64Attribute{
							Computed:    true,
							Description: "Amount of memory on the machine",
						},
						"disks": schema.ListNestedAttribute{
							Computed:    true,
							Description: "List of physical disks on the machine",
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"device": schema.StringAttribute{
										Computed:    true,
										Description: "The device name of the disk",
									},
									"size": schema.Int64Attribute{
										Computed:    true,
										Description: "The size of the disk in bytes",
									},
									"model": schema.StringAttribute{
										Computed:    true,
										Description: "The model of the disk",
									},
									"serial": schema.StringAttribute{
										Computed:    true,
										Description: "The serial of the disk",
									},
									"vendor": schema.StringAttribute{
										Computed:    true,
										Description: "The vendor of the disk",
									},
									"type": schema.StringAttribute{
										Computed:    true,
										Description: "The type of the disk",
									},
								},
							},
						},
						"network_interfaces": schema.ListAttribute{
							Computed:    true,
							ElementType: types.StringType,
							Description: "List of physical network interfaces on the machine.",
						},
					},
				},
			},
		},
	}
}

func (d *nodesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state nodesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	nodes := filters.DetermineNode(d.client, state.Filters)

	nodeList := []service.Node{}
	for _, node := range nodes {
		n, err := d.client.DescribeNode(ctx, node)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to get nodes",
				"An error was encountered retrieving nodes.\n"+
					err.Error(),
			)
			return
		}
		nodeList = append(nodeList, n)
	}

	for _, node := range nodeList {
		stateNode := nodeModel{
			ID:             types.StringValue(node.Id),
			Node:           types.StringValue(node.Node),
			Cores:          types.Int64Value(int64(node.Cores)),
			SslFingerprint: types.StringValue(node.SslFingerprint),
			Memory:         types.Int64Value(int64(node.Memory)),
		}

		for _, iface := range node.NetworkInterfaces {
			stateNode.Interfaces = append(stateNode.Interfaces, types.StringValue(iface.Name))
		}

		for _, disk := range node.Disks {
			stateDisk := diskModel{
				Device: types.StringValue(disk.Device),
				Size:   types.Int64Value(int64(disk.Size)),
				Model:  types.StringValue(disk.Model),
				Serial: types.StringValue(disk.Serial),
				Vendor: types.StringValue(disk.Vendor),
				Type:   types.StringValue(string(disk.Type)),
			}
			stateNode.Disks = append(stateNode.Disks, stateDisk)
		}

		state.Nodes = append(state.Nodes, stateNode)
	}

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
