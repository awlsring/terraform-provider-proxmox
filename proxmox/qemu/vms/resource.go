package vms

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/awlsring/terraform-provider-proxmox/internal/service"
	"github.com/awlsring/terraform-provider-proxmox/proxmox/qemu"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                   = &virtualMachineResource{}
	_ resource.ResourceWithConfigure      = &virtualMachineResource{}
	_ resource.ResourceWithImportState    = &virtualMachineResource{}
	_ resource.ResourceWithValidateConfig = &virtualMachineResource{}
)

func Resource() resource.Resource {
	return &virtualMachineResource{}
}

type virtualMachineResource struct {
	client *service.Proxmox
}

func (r *virtualMachineResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_virtual_machine"
}

func (r *virtualMachineResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = qemu.ResourceSchema
}

func (r *virtualMachineResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*service.Proxmox)
}

func (r *virtualMachineResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	tflog.Debug(ctx, "Validate Config virtual machine method")
	tflog.Debug(ctx, fmt.Sprintf("virtual machine '%v'", req.Config))
}

func (r *virtualMachineResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Debug(ctx, "Create virtual machine method")
	var plan qemu.VirtualMachineResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("virtual machine '%v'", plan))
	tflog.Debug(ctx, fmt.Sprintf("disks specified '%v'", len(plan.Disks.Elements())))
	tflog.Debug(ctx, fmt.Sprintf("nics specified '%v'", len(plan.NetworkInterfaces.Elements())))

	// create
	tflog.Debug(ctx, "Creating virtual machine")
	err := r.routeCreateVm(ctx, &plan)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating virtual machine",
			"Could not create virtual machine, unexpected error: "+err.Error(),
		)
		return
	}

	// // configure
	tflog.Debug(ctx, "Configuring virtual machine")
	err = r.configureVm(ctx, &plan)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error configuring virtual machine",
			"Could not configure virtual machine, unexpected error: "+err.Error(),
		)
		return
	}

	// read
	tflog.Debug(ctx, "Reading virtual machine")
	m, err := r.readModelWithStateContext(ctx, plan.Node.ValueString(), int(plan.ID.ValueInt64()), &plan)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading virtual machine",
			"Could not read virtual machine, unexpected error: "+err.Error(),
		)
		return
	}

	// launch
	if plan.StartOnCreate.ValueBool() {
		tflog.Debug(ctx, "Starting virtual machine")
		err = r.client.StartVirtualMachine(ctx, plan.Node.ValueString(), int(plan.ID.ValueInt64()))
		if err != nil {
			resp.Diagnostics.AddError(
				"Error starting virtual machine",
				"Could not start virtual machine, unexpected error: "+err.Error(),
			)
			return
		}
	}

	diags = resp.State.Set(ctx, &m)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *virtualMachineResource) readModelWithStateContext(ctx context.Context, node string, id int, state *qemu.VirtualMachineResourceModel) (*qemu.VirtualMachineResourceModel, error) {
	vm, err := r.client.DescribeVirtualMachine(ctx, node, id)
	if err != nil {
		return nil, err
	}

	j, _ := json.Marshal(vm)
	tflog.Debug(ctx, fmt.Sprintf("vm '%v'", string(j)))

	model := qemu.VMToModel(ctx, vm, state)
	model.Clone = state.Clone
	model.ISO = state.ISO
	model.Timeouts = state.Timeouts
	model.StartOnCreate = state.StartOnCreate

	return &model, nil
}

func (r *virtualMachineResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Debug(ctx, "Read virtual machine method")
	var state qemu.VirtualMachineResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	vm, err := r.client.DescribeVirtualMachine(ctx, state.Node.ValueString(), int(state.ID.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading virtual machine",
			"Could not read virtual machine, unexpected error: "+err.Error(),
		)
		return
	}

	model := qemu.VMToModel(ctx, vm, &state)

	if state.Clone != nil {
		model.Clone = state.Clone
	}
	if state.ISO != nil {
		model.ISO = state.ISO
	}

	diags = resp.State.Set(ctx, &model)
	resp.Diagnostics.Append(diags...)

}

func (r *virtualMachineResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	tflog.Debug(ctx, "Update virtual machine method")
	var plan qemu.VirtualMachineResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state qemu.VirtualMachineResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *virtualMachineResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	tflog.Debug(ctx, "Delete virtual machine method")
	var state qemu.VirtualMachineResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *virtualMachineResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
