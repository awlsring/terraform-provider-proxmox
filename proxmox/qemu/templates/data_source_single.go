package templates

import (
	"context"
	"fmt"

	"github.com/awlsring/terraform-provider-proxmox/internal/service"
	qt "github.com/awlsring/terraform-provider-proxmox/proxmox/qemu/types"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ datasource.DataSource                   = &templateDataSource{}
	_ datasource.DataSourceWithConfigure      = &templateDataSource{}
	_ datasource.DataSourceWithValidateConfig = &templateDataSource{}
)

func DataSourceSingle() datasource.DataSource {
	return &templateDataSource{}
}

type templateDataSource struct {
	client *service.Proxmox
}

func (d *templateDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_template"
}

func (d *templateDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*service.Proxmox)
}

func (d *templateDataSource) ValidateConfig(ctx context.Context, req datasource.ValidateConfigRequest, resp *datasource.ValidateConfigResponse) {
	tflog.Debug(ctx, "validate config template")
	idValue := basetypes.Int64Value{}
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("id"), &idValue)...)

	nameValue := basetypes.StringValue{}
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("name"), &nameValue)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if nameValue.IsNull() && idValue.IsNull() {
		resp.Diagnostics.AddError(
			"Invalid template parameters",
			"Either name or id must be specified",
		)
		return
	}
}

func (d *templateDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: TemplateSingleDataSourceSchema,
	}
}

func (d *templateDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	tflog.Debug(ctx, "Read template")
	nodeValue := basetypes.StringValue{}

	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("node"), &nodeValue)...)
	node := nodeValue.ValueString()

	idValue := basetypes.Int64Value{}
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("id"), &idValue)...)

	nameValue := basetypes.StringValue{}
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("name"), &nameValue)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var vm *service.VirtualMachine
	var err error
	if !idValue.IsNull() {
		tflog.Debug(ctx, fmt.Sprintf("Getting template by id %v", idValue.ValueInt64()))
		id := int(idValue.ValueInt64())

		vm, err = d.client.DescribeTemplate(ctx, node, id)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to get templates",
				"An error was encountered retrieving templates.\n"+
					err.Error(),
			)
			return
		}
	} else if !nameValue.IsNull() {
		tflog.Debug(ctx, fmt.Sprintf("Getting template by name %v", nameValue.ValueString()))
		name := nameValue.ValueString()

		vm, err = d.client.DescribeTemplateFromName(ctx, node, name)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to get templates",
				"An error was encountered retrieving templates.\n"+
					err.Error(),
			)
			return
		}
	} else {
		resp.Diagnostics.AddError(
			"Unable to get templates",
			"An error was encountered retrieving templates.\n"+
				"Either id or name must be set",
		)
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("Converting template %v to model", vm.VmId))
	model := qt.VMToModel(ctx, vm)

	diags := resp.State.Set(ctx, &model)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
