package bonds

import (
	"context"
	"fmt"
	"time"

	"github.com/awlsring/proxmox-go/proxmox"
	"github.com/awlsring/terraform-provider-proxmox/internal/service"
	"github.com/awlsring/terraform-provider-proxmox/proxmox/network"
	"github.com/awlsring/terraform-provider-proxmox/proxmox/utils"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &bondResource{}
	_ resource.ResourceWithConfigure   = &bondResource{}
	_ resource.ResourceWithImportState = &bondResource{}
)

func NewResource() resource.Resource {
	return &bondResource{}
}

type bondResource struct {
	client *service.Proxmox
}

func (r *bondResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_network_bond"
}

func (r *bondResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resourceSchema
}

func (r *bondResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*service.Proxmox)
}

func (r *bondResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan bondModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var name string
	var err error
	if plan.Name.IsUnknown() || plan.Name.IsNull() {
		tflog.Debug(ctx, "name is null, assinging next available name")
		name, err = r.generateName(ctx, plan.Node.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Error creating bond",
				"Could not create bond, unexpected error: "+err.Error(),
			)
			return
		}
	} else {
		tflog.Debug(ctx, fmt.Sprintf("using given name name '%s'", plan.Name.ValueString()))
		name = plan.Name.ValueString()
	}

	request := &service.CreateNetworkBondInput{
		Interfaces:  []string{},
		Name:        name,
		Node:        plan.Node.ValueString(),
		Mode:        proxmox.NetworkInterfaceBondMode(plan.Mode.ValueString()),
		BondPrimary: utils.OptionalToPointerString(plan.BondPrimary.ValueString()),
		AutoStart:   utils.OptionalToPointerBool(plan.Autostart.ValueBool()),
		Comments:    utils.OptionalToPointerString(plan.Comments.ValueString()),
		IPv4Gateway: utils.OptionalToPointerString(plan.IPv4Gateway.ValueString()),
		IPv6Gateway: utils.OptionalToPointerString(plan.IPv6Gateway.ValueString()),
	}

	if plan.IPv4 != nil {
		request.IPv4 = &service.IP{
			Address: plan.IPv4.Address.ValueString(),
			Netmask: plan.IPv4.Netmask.ValueString(),
		}
	}

	if plan.IPv6 != nil {
		request.IPv6 = &service.IP{
			Address: plan.IPv6.Address.ValueString(),
			Netmask: plan.IPv6.Netmask.ValueString(),
		}
	}

	for _, i := range plan.Interfaces {
		request.Interfaces = append(request.Interfaces, i.ValueString())
	}

	if !plan.HashPolicy.IsNull() {
		h := proxmox.NetworkInterfaceBondHashPolicy(plan.HashPolicy.ValueString())
		request.HashPolicy = &h
	}

	err = r.client.CreateNetworkBond(ctx, request)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating bond",
			"Could not create bond, unexpected error: "+err.Error(),
		)
		return
	}

	var bond bondModel
	tries := 0
	for {
		bond, err = r.readBondModel(ctx, network.FormId(plan.Node.ValueString(), name))
		if err != nil {
			if tries < 3 {
				tflog.Debug(ctx, "could not read bond, retrying...")
				tries += 1
				time.Sleep(time.Duration(tries) * time.Second)
				continue
			}
			resp.Diagnostics.AddError(
				"Error creating bond",
				"Could not read create bond, unexpected error: "+err.Error(),
			)
			return
		}
		break
	}
	bond.Active = types.BoolValue(true)
	tflog.Debug(ctx, fmt.Sprintf("created bond with id '%s'", bond.ID.ValueString()))

	diags = resp.State.Set(ctx, &bond)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *bondResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Debug(ctx, "Read bond method")
	var state bondModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	bondModel, err := r.readBondModel(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading bond",
			"Could not read bond, unexpected error: "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("Bond model '%v'", bondModel))

	diags = resp.State.Set(ctx, &bondModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *bondResource) generateName(ctx context.Context, node string) (string, error) {
	tflog.Debug(ctx, "Generating name")
	bonds, err := r.client.ListNetworkBonds(ctx, node)
	if err != nil {
		return "", err
	}

	bondsNames := []string{}
	for _, bond := range bonds {
		bondsNames = append(bondsNames, bond.Iface)
	}

	name := network.GenerateInterfaceName("bond", bondsNames)
	tflog.Debug(ctx, fmt.Sprintf("Generated name: %s", name))
	return name, nil
}

func (r *bondResource) readBondModel(ctx context.Context, id string) (bondModel, error) {
	tflog.Debug(ctx, fmt.Sprintf("Reading bond model: %s", id))

	node, name, err := network.UnpackId(id)
	if err != nil {
		return bondModel{}, err
	}

	bond, err := r.client.GetNetworkBond(ctx, node, name)
	if err != nil {
		return bondModel{}, err
	}

	return BondToModel(bond), nil
}

func (r *bondResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	tflog.Debug(ctx, "Update bond method")
	var plan bondModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state bondModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("bond state '%v'", state))
	tflog.Debug(ctx, fmt.Sprintf("bond plan '%v'", plan))

	bond := bondModel{
		ID:         state.ID,
		Name:       state.Name,
		Node:       state.Node,
		Interfaces: plan.Interfaces,
		Mode:       plan.Mode,
		Autostart:  plan.Active,
		Active:     state.Active,
		MiiMon:     plan.MiiMon,
	}

	if !plan.HashPolicy.IsNull() {
		bond.HashPolicy = plan.HashPolicy
	}

	if !plan.BondPrimary.IsNull() {
		bond.BondPrimary = plan.BondPrimary
	}

	if !plan.Autostart.IsNull() {
		bond.Autostart = plan.Autostart
	}

	if !plan.Comments.IsNull() {
		bond.Comments = plan.Comments
	}

	if plan.IPv4 != nil {
		bond.IPv4 = &network.IpAddressModel{
			Address: plan.IPv4.Address,
			Netmask: plan.IPv4.Netmask,
		}
	}

	if plan.IPv6 != nil {
		bond.IPv6 = &network.IpAddressModel{
			Address: plan.IPv6.Address,
			Netmask: plan.IPv6.Netmask,
		}
	}

	if !plan.IPv4Gateway.IsNull() && !plan.IPv4Gateway.IsUnknown() {
		bond.IPv4Gateway = plan.IPv4Gateway
	}

	if !plan.IPv6Gateway.IsNull() && !plan.IPv6Gateway.IsUnknown() {
		bond.IPv6Gateway = plan.IPv6Gateway
	}

	request := &service.UpdateNetworkBondInput{
		Interfaces:  []string{},
		Name:        bond.Name.ValueString(),
		Node:        bond.Node.ValueString(),
		Mode:        proxmox.NetworkInterfaceBondMode(bond.Mode.ValueString()),
		BondPrimary: utils.OptionalToPointerString(bond.BondPrimary.ValueString()),
		AutoStart:   utils.OptionalToPointerBool(bond.Autostart.ValueBool()),
		Comments:    utils.OptionalToPointerString(bond.Comments.ValueString()),
		IPv4Gateway: utils.OptionalToPointerString(bond.IPv4Gateway.ValueString()),
		IPv6Gateway: utils.OptionalToPointerString(bond.IPv6Gateway.ValueString()),
	}

	if bond.IPv4 != nil {
		request.IPv4 = &service.IP{
			Address: plan.IPv4.Address.ValueString(),
			Netmask: plan.IPv4.Netmask.ValueString(),
		}
	}

	if bond.IPv6 != nil {
		request.IPv6 = &service.IP{
			Address: plan.IPv6.Address.ValueString(),
			Netmask: plan.IPv6.Netmask.ValueString(),
		}
	}

	for _, i := range bond.Interfaces {
		request.Interfaces = append(request.Interfaces, i.ValueString())
	}

	if !plan.HashPolicy.IsNull() {
		h := proxmox.NetworkInterfaceBondHashPolicy(plan.HashPolicy.ValueString())
		bond.HashPolicy = types.StringValue(string(h))
		request.HashPolicy = &h
	}

	err := r.client.UpdateNetworkBond(ctx, request)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating bond",
			"Could not create bond, unexpected error: "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("Updating bond: '%s' '%s'", state.Node.ValueString(), state.Name.ValueString()))

	diags = resp.State.Set(ctx, &bond)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *bondResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	tflog.Debug(ctx, "Delete bond method")
	var state bondModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("bond state '%v'", state))

	tflog.Debug(ctx, fmt.Sprintf("Deleting bond: '%s' '%s'", state.Node.ValueString(), state.Name.ValueString()))
	err := r.client.DeleteNetworkInterface(ctx, state.Node.ValueString(), state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting bond",
			"Could not delete bond, unexpected error: "+err.Error(),
		)
		return
	}
}

func (r *bondResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)

	model, err := r.readBondModel(ctx, req.ID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading bond",
			"Could not read bond, unexpected error: "+err.Error(),
		)
		return
	}

	diags := resp.State.Set(ctx, model)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
