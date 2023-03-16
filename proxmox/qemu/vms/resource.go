package vms

import (
	"context"
	"fmt"

	"github.com/awlsring/proxmox-go/proxmox"
	"github.com/awlsring/terraform-provider-proxmox/internal/service"
	"github.com/awlsring/terraform-provider-proxmox/proxmox/qemu/vms/schemas"
	vt "github.com/awlsring/terraform-provider-proxmox/proxmox/qemu/vms/types"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource               = &virtualMachineResource{}
	_ resource.ResourceWithConfigure  = &virtualMachineResource{}
	_ resource.ResourceWithModifyPlan = &virtualMachineResource{}
)

func Resource() resource.Resource {
	return &virtualMachineResource{}
}

type virtualMachineResource struct {
	client   *service.Proxmox
	timeouts *VirtualMachineTimeouts
}

type ConfigureMode string

const (
	ConfigureModeCreate ConfigureMode = "create"
	ConfigureModeUpdate ConfigureMode = "update"
	ConfigureModeDelete ConfigureMode = "delete"
)

func (r *virtualMachineResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_virtual_machine"
}

func (r *virtualMachineResource) createPlanModifiers(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse, plan *vt.VirtualMachineResourceModel) {
	authCreateValidator(ctx, r.client.IsRoot, plan, resp)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *virtualMachineResource) updatePlanModifiers(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse, state *vt.VirtualMachineResourceModel, plan *vt.VirtualMachineResourceModel) {
	node := plan.Node.ValueString()
	vmId := int(state.ID.ValueInt64())

	status, err := r.client.GetVirtualMachineStatus(ctx, node, vmId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading virtual machine",
			"Could not read virtual machine, unexpected error: "+err.Error(),
		)
		return
	}
	if status.Status != proxmox.VIRTUALMACHINESTATUS_RUNNING {
		powerOffValidator(ctx, r.client, state, plan, resp)
	}

	authUpdateValidator(ctx, r.client.IsRoot, plan, resp)

	changeValidatorDiskSize(ctx, state, plan, resp)
	changeValidatorDiskStorage(ctx, state, plan, resp)
	changeValidatorDiskRemoved(ctx, state, plan, resp)
	// carry over computed values sets to prevent unnecessary diffs
	amended := plan
	amended.ComputedDisks = state.ComputedDisks
	amended.ComputedNetworkInterfaces = state.ComputedNetworkInterfaces
	amended.ComputedPCIDevices = state.ComputedPCIDevices

	diags := resp.Plan.Set(ctx, &amended)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *virtualMachineResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	tflog.Debug(ctx, "ModifyPlan virtual machine method")
	configureMode := ConfigureModeUpdate

	var plan vt.VirtualMachineResourceModel
	diags := req.Plan.Get(ctx, &plan)
	if diags.HasError() {
		configureMode = ConfigureModeDelete
	}

	var state vt.VirtualMachineResourceModel
	diags = req.State.Get(ctx, &state)
	if diags.HasError() {
		configureMode = ConfigureModeCreate
	}

	tflog.Info(ctx, fmt.Sprintf("configureMode '%v'", configureMode))

	switch configureMode {
	case ConfigureModeCreate:
		r.createPlanModifiers(ctx, req, resp, &plan)
		return
	case ConfigureModeDelete:
		return
	default:
		r.updatePlanModifiers(ctx, req, resp, &state, &plan)
	}
}

func (r *virtualMachineResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schemas.ResourceSchema
}

func (r *virtualMachineResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*service.Proxmox)
	r.timeouts = &timeoutDefaults
}

func (r *virtualMachineResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Debug(ctx, "Create virtual machine method")
	var plan vt.VirtualMachineResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, fmt.Sprintf("plan '%v'", plan))
	r.timeouts = loadTimeouts(ctx, plan.Timeouts)

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

	node := plan.Node.ValueString()
	vmId := int(plan.ID.ValueInt64())
	vm, err := r.client.DescribeVirtualMachine(ctx, node, vmId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating virtual machine",
			"Could not create virtual machine, unexpected error: "+err.Error(),
		)
		return
	}
	currentModel := vt.VMToResourceModel(ctx, vm, &plan)

	// configure
	tflog.Debug(ctx, "Configuring virtual machine")
	err = r.determineVmConfigurations(ctx, currentModel, &plan)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error configuring virtual machine",
			"Could not configure virtual machine, unexpected error: "+err.Error(),
		)
		return
	}

	// read
	tflog.Debug(ctx, "Reading virtual machine")
	m, err := r.readModelWithContext(ctx, plan.Node.ValueString(), int(plan.ID.ValueInt64()), &plan)
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

func (r *virtualMachineResource) readModelWithContext(ctx context.Context, node string, id int, state *vt.VirtualMachineResourceModel) (*vt.VirtualMachineResourceModel, error) {
	vm, err := r.client.DescribeVirtualMachine(ctx, node, id)
	if err != nil {
		return nil, err
	}
	model := vt.VMToResourceModel(ctx, vm, state)

	if !state.ResourcePool.IsNull() {
		tflog.Debug(ctx, "Determining resource pool")
		in, pool, err := r.client.DetermineVirtualMachineResourcePool(ctx, id)
		if err != nil {
			tflog.Error(ctx, "Error determining resource pool:"+err.Error())
			return nil, err
		}
		if in {
			tflog.Debug(ctx, fmt.Sprintf("Resource pool '%v'", pool))
			model.ResourcePool = types.StringValue(pool)
		}
	}

	// pass resource tf metadata
	model.Clone = state.Clone
	model.ISO = state.ISO
	model.Timeouts = state.Timeouts
	model.StartOnCreate = state.StartOnCreate

	return model, nil
}

func (r *virtualMachineResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Debug(ctx, "Read virtual machine method")
	var state vt.VirtualMachineResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// read
	tflog.Debug(ctx, "Reading virtual machine")
	m, err := r.readModelWithContext(ctx, state.Node.ValueString(), int(state.ID.ValueInt64()), &state)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading virtual machine",
			"Could not read virtual machine, unexpected error: "+err.Error(),
		)
		return
	}
	m.ComputedDisks = state.ComputedDisks
	m.ComputedNetworkInterfaces = state.ComputedNetworkInterfaces
	m.ComputedPCIDevices = state.ComputedPCIDevices

	diags = resp.State.Set(ctx, &m)
	resp.Diagnostics.Append(diags...)
}

func (r *virtualMachineResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	tflog.Debug(ctx, "Update virtual machine method")
	var plan vt.VirtualMachineResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state vt.VirtualMachineResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	node := plan.Node.ValueString()
	vmId := int(state.ID.ValueInt64())
	r.timeouts = loadTimeouts(ctx, plan.Timeouts)

	stopped, err := r.stopIfSensitivePropertyChanged(ctx, &state, &plan)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error getting virtual machine state",
			"Could not configure virtual machine, unexpected error: "+err.Error(),
		)
		return
	}
	tflog.Debug(ctx, fmt.Sprintf("vm was stopped: '%v'", stopped))

	tflog.Debug(ctx, fmt.Sprintf("vm state '%v'", state))
	tflog.Debug(ctx, fmt.Sprintf("vm plan '%v'", plan))

	tflog.Debug(ctx, "Configuring virtual machine")
	err = r.determineVmConfigurations(ctx, &state, &plan)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error configuring virtual machine",
			"Could not configure virtual machine, unexpected error: "+err.Error(),
		)
		return
	}

	err = r.modifyResourcePool(ctx, vmId, state.ResourcePool, plan.ResourcePool)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error modifying virtual machine resource pool",
			"Could not modify virtual machine resource pool, unexpected error: "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "Reading virtual machine")
	m, err := r.readModelWithContext(ctx, node, vmId, &plan)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading virtual machine",
			"Could not read virtual machine, unexpected error: "+err.Error(),
		)
		return
	}

	if stopped {
		err = r.startVm(ctx, node, vmId)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error configuring virtual machine",
				"Could not configure virtual machine, unexpected error: "+err.Error(),
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

func (r *virtualMachineResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	tflog.Debug(ctx, "Delete virtual machine method")
	var state vt.VirtualMachineResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	node := state.Node.ValueString()
	vmId := int(state.ID.ValueInt64())
	r.timeouts = loadTimeouts(ctx, state.Timeouts)

	err := r.stopVm(ctx, node, vmId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error stopping virtual machine",
			"Could not stop virtual machine, unexpected error: "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("Deleting vm: '%s' '%v'", node, vmId))
	err = r.deleteVm(ctx, node, vmId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting vm",
			"Could not delete vm, unexpected error: "+err.Error(),
		)
		return
	}
}
