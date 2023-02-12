package pools

import (
	"context"

	"github.com/awlsring/terraform-provider-proxmox/internal/service"
	"github.com/awlsring/terraform-provider-proxmox/proxmox/filters"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &poolsDataSource{}
	_ datasource.DataSourceWithConfigure = &poolsDataSource{}
)

func DataSource() datasource.DataSource {
	return &poolsDataSource{}
}

type poolsDataSource struct {
	client *service.Proxmox
}

func (d *poolsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_pools"
}

func (d *poolsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*service.Proxmox)
}

var filter = filters.FilterConfig{}

func (d *poolsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"filters": filter.Schema(),
			"pools": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:    true,
							Description: "The id of the pool.",
						},
						"comment": schema.StringAttribute{
							Computed:    true,
							Description: "Notes on the pool.",
						},
						"members": schema.ListNestedAttribute{
							Computed:    true,
							Description: "Resources that are part of the pool.",
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"id": schema.StringAttribute{
										Computed:    true,
										Description: "The id of the resource.",
									},
									"type": schema.StringAttribute{
										Computed:    true,
										Description: "The type of the resource.",
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func (d *poolsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state poolsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	pools, err := d.client.DescribePools(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to get pools",
			"An error was encountered retrieving pools.\n"+
				err.Error(),
		)
		return
	}

	for _, pool := range pools {
		statePool := poolModel{
			ID:      types.StringValue(pool.Id),
			Comment: types.StringValue(pool.Comment),
		}

		for _, member := range pool.Members {
			statePool.Members = append(statePool.Members, poolMemberModel{
				ID:   types.StringValue(member.Id),
				Type: types.StringValue(string(member.Type)),
			})
		}

		state.Pools = append(state.Pools, statePool)
	}

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
