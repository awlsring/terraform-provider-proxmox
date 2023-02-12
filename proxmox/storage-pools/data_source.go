package storage_pools

import (
	"context"

	"github.com/awlsring/terraform-provider-proxmox/internal/service"
	"github.com/awlsring/terraform-provider-proxmox/proxmox/filters"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &storageDataSource{}
	_ datasource.DataSourceWithConfigure = &storageDataSource{}
)

func DataSource() datasource.DataSource {
	return &storageDataSource{}
}

type storageDataSource struct {
	client *service.Proxmox
}

type storageDataSourceModel struct {
	Storage []storageModel        `tfsdk:"storage"`
	Filters []filters.FilterModel `tfsdk:"filters"`
}

type storageModel struct {
	ID          types.String   `tfsdk:"id"`
	SharedNodes []types.String `tfsdk:"shared_nodes"`
	Shared      types.Bool     `tfsdk:"shared"`
	Local       types.Bool     `tfsdk:"local"`
	Size        types.Int64    `tfsdk:"size"`
	Type        types.String   `tfsdk:"type"`
	Content     []types.String `tfsdk:"content"`
	Source      types.String   `tfsdk:"source"`
}

func (d *storageDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_storage"
}

func (d *storageDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*service.Proxmox)
}

var filter = filters.FilterConfig{}

func (d *storageDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"filters": filter.Schema(),
			"storage": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:    true,
							Description: "The id of the bridge. Formatted as /{node}/{name}.",
						},
						"shared_nodes": schema.ListAttribute{
							Computed:    true,
							ElementType: types.StringType,
							Description: "The nodes this storage is shared with.",
						},
						"shared": schema.BoolAttribute{
							Computed:    true,
							Description: "Whether this storage is shared.",
						},
						"local": schema.BoolAttribute{
							Computed:    true,
							Description: "Whether this storage is local.",
						},
						"size": schema.Int64Attribute{
							Computed:    true,
							Description: "The space available on the storage.",
						},
						"type": schema.StringAttribute{
							Computed:    true,
							Description: "The type of the storage.",
						},
						"content": schema.ListAttribute{
							Computed:    true,
							ElementType: types.StringType,
							Description: "The content types on the storage.",
						},
						"source": schema.StringAttribute{
							Computed:    true,
							Description: "The source of the storage.",
						},
					},
				},
			},
		},
	}
}

func (d *storageDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state storageDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	storage := []service.Storage{}
	rs, err := d.client.DescribeStorage(context.Background())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to get storage",
			"An error was encountered retrieving storage.\n"+
				err.Error(),
		)
		return
	}
	storage = append(storage, rs...)

	ls, err := d.client.DescribeLocalStorage(context.Background())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to get storage",
			"An error was encountered retrieving storage.\n"+
				err.Error(),
		)
		return
	}
	storage = append(storage, ls...)

	for _, s := range storage {
		statePool := storageModel{
			ID:          types.StringValue(s.Id),
			SharedNodes: []types.String{},
			Shared:      types.BoolValue(s.Shared),
			Local:       types.BoolValue(s.Local),
			Size:        types.Int64Value(s.Size),
			Type:        types.StringValue(string(s.Type)),
			Content:     []types.String{},
			Source:      types.StringValue(s.Source),
		}

		for _, c := range s.Content {
			statePool.Content = append(statePool.Content, types.StringValue(c))
		}

		for _, n := range s.SharedNodes {
			statePool.SharedNodes = append(statePool.SharedNodes, types.StringValue(n))
		}

		state.Storage = append(state.Storage, statePool)
	}

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
