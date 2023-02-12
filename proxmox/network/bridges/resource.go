package bridges

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/awlsring/terraform-provider-proxmox/internal/service"
	"github.com/awlsring/terraform-provider-proxmox/proxmox/network"
	"github.com/awlsring/terraform-provider-proxmox/proxmox/utils"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &bridgeResource{}
	_ resource.ResourceWithConfigure   = &bridgeResource{}
	_ resource.ResourceWithImportState = &bridgeResource{}
)

func Resource() resource.Resource {
	return &bridgeResource{}
}

type bridgeResource struct {
	client *service.Proxmox
}

func (r *bridgeResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_network_bridge"
}

func (r *bridgeResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resourceSchema
}

func (r *bridgeResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*service.Proxmox)
}

func (r *bridgeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Debug(ctx, "Create bridge method")
	var plan bridgeModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Determining bridge name")
	var name string
	var err error
	if plan.Name.IsUnknown() || plan.Name.IsNull() {
		tflog.Debug(ctx, "name is null, assinging next available name")
		name, err = r.generateName(ctx, plan.Node.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Error creating bridge",
				"Could not create bridge, unexpected error: "+err.Error(),
			)
			return
		}
	} else {
		tflog.Debug(ctx, fmt.Sprintf("using given name name '%s'", plan.Name.ValueString()))
		name = plan.Name.ValueString()
	}

	tflog.Debug(ctx, "Forming create bridge request")
	request := &service.CreateNetworkBridgeInput{
		Interfaces:  []string{},
		Name:        name,
		Node:        plan.Node.ValueString(),
		VLANAware:   utils.OptionalToPointerBool(plan.VLANAware.ValueBool()),
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

	j, _ := json.Marshal(request)
	tflog.Debug(ctx, fmt.Sprintf("request: %s", j))

	err = r.client.CreateNetworkBridge(ctx, request)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating bridge",
			"Could not create bridge, unexpected error: "+err.Error(),
		)
		return
	}

	var bridge bridgeModel
	tries := 0
	for {
		bridge, err = r.readBridgeModel(ctx, network.FormId(plan.Node.ValueString(), name))
		if err != nil {
			if tries < 3 {
				tflog.Debug(ctx, "could not read bridge, retrying...")
				tries += 1
				time.Sleep(time.Duration(tries) * time.Second)
				continue
			}
			resp.Diagnostics.AddError(
				"Error creating bridge",
				"Could not read create bridge, unexpected error: "+err.Error(),
			)
			return
		}
		break
	}
	bridge.Active = types.BoolValue(true)
	tflog.Debug(ctx, fmt.Sprintf("created bridge with id '%s'", bridge.ID.ValueString()))

	diags = resp.State.Set(ctx, &bridge)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *bridgeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Debug(ctx, "Read bridge method")
	var state bridgeModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	bridgeModel, err := r.readBridgeModel(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading bridge",
			"Could not read bridge, unexpected error: "+err.Error(),
		)
		return
	}

	diags = resp.State.Set(ctx, &bridgeModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *bridgeResource) generateName(ctx context.Context, node string) (string, error) {
	tflog.Debug(ctx, "Generating name")
	bridges, err := r.client.ListNetworkBridges(ctx, node)
	if err != nil {
		return "", err
	}

	bridgeNames := []string{}
	for _, bridge := range bridges {
		bridgeNames = append(bridgeNames, bridge.Iface)
	}

	name := network.GenerateInterfaceName("vmbr", bridgeNames)
	tflog.Debug(ctx, fmt.Sprintf("Generated name: %s", name))
	return name, nil
}

func (r *bridgeResource) readBridgeModel(ctx context.Context, id string) (bridgeModel, error) {
	tflog.Debug(ctx, fmt.Sprintf("Reading bridge model: %s", id))

	node, name, err := network.UnpackId(id)
	if err != nil {
		return bridgeModel{}, err
	}

	bridge, err := r.client.GetNetworkBridge(ctx, node, name)
	if err != nil {
		return bridgeModel{}, err
	}

	return BridgeToModel(bridge), nil
}

func (r *bridgeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	tflog.Debug(ctx, "Update bridge method")
	var plan bridgeModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state bridgeModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("bridge state '%v'", state))
	tflog.Debug(ctx, fmt.Sprintf("bridge plan '%v'", plan))

	bridge := bridgeModel{
		ID:         state.ID,
		Name:       state.Name,
		Node:       state.Node,
		Interfaces: plan.Interfaces,
		Autostart:  state.Autostart,
		Active:     state.Active,
		VLANAware:  state.Active,
	}

	if !plan.Autostart.IsNull() && !plan.Autostart.IsUnknown() {
		bridge.Autostart = plan.Autostart
	}

	if !plan.Comments.IsNull() && !plan.Comments.IsUnknown() {
		bridge.Comments = plan.Comments
	}

	if !plan.VLANAware.IsNull() && !plan.VLANAware.IsUnknown() {
		bridge.VLANAware = plan.VLANAware
	}

	if plan.IPv4 != nil {
		bridge.IPv4 = &network.IpAddressModel{
			Address: plan.IPv4.Address,
			Netmask: plan.IPv4.Netmask,
		}
	}

	if plan.IPv6 != nil {
		bridge.IPv6 = &network.IpAddressModel{
			Address: plan.IPv6.Address,
			Netmask: plan.IPv6.Netmask,
		}
	}

	if !plan.IPv4Gateway.IsNull() && !plan.IPv4Gateway.IsUnknown() {
		bridge.IPv4Gateway = plan.IPv4Gateway
	}

	if !plan.IPv6Gateway.IsNull() && !plan.IPv6Gateway.IsUnknown() {
		bridge.IPv6Gateway = plan.IPv6Gateway
	}

	request := &service.UpdateNetworkBridgeInput{
		Interfaces:  []string{},
		Name:        bridge.Name.ValueString(),
		Node:        bridge.Node.ValueString(),
		VLANAware:   utils.OptionalToPointerBool(bridge.VLANAware.ValueBool()),
		AutoStart:   utils.OptionalToPointerBool(bridge.Autostart.ValueBool()),
		Comments:    utils.OptionalToPointerString(bridge.Comments.ValueString()),
		IPv4Gateway: utils.OptionalToPointerString(bridge.IPv4Gateway.ValueString()),
		IPv6Gateway: utils.OptionalToPointerString(bridge.IPv6Gateway.ValueString()),
	}

	if bridge.IPv4 != nil {
		request.IPv4 = &service.IP{
			Address: plan.IPv4.Address.ValueString(),
			Netmask: plan.IPv4.Netmask.ValueString(),
		}
	} else {
		if plan.IPv4 == nil && state.IPv4 != nil {
			request.IPv4 = &service.IP{
				Address: "",
				Netmask: "",
			}
		}
	}

	if bridge.IPv6 != nil {
		request.IPv6 = &service.IP{
			Address: plan.IPv6.Address.ValueString(),
			Netmask: plan.IPv6.Netmask.ValueString(),
		}
	}

	for _, i := range bridge.Interfaces {
		request.Interfaces = append(request.Interfaces, i.ValueString())
	}

	err := r.client.UpdateNetworkBridge(ctx, request)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating bridge",
			"Could not create bridge, unexpected error: "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("Updating bridge: '%s' '%s'", state.Node.ValueString(), state.Name.ValueString()))

	diags = resp.State.Set(ctx, &bridge)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *bridgeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	tflog.Debug(ctx, "Delete bridge method")
	var state bridgeModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("Deleting bridge: '%s' '%s'", state.Node.ValueString(), state.Name.ValueString()))
	err := r.client.DeleteNetworkInterface(ctx, state.Node.ValueString(), state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting bridge",
			"Could not delete bridge, unexpected error: "+err.Error(),
		)
		return
	}
}

func (r *bridgeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)

	model, err := r.readBridgeModel(ctx, req.ID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading bridge",
			"Could not read bridge, unexpected error: "+err.Error(),
		)
		return
	}

	diags := resp.State.Set(ctx, model)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
