package vms

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/awlsring/proxmox-go/proxmox"
	"github.com/awlsring/terraform-provider-proxmox/internal/service"
	"github.com/awlsring/terraform-provider-proxmox/proxmox/qemu"
	"github.com/hashicorp/terraform-plugin-framework/resource"
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
	client *service.Proxmox
}

func (r *virtualMachineResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_virtual_machine"
}

func (r *virtualMachineResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	tflog.Debug(ctx, "ModifyPlan virtual machine method")
	var plan qemu.VirtualMachineResourceModel
	diags := req.Plan.Get(ctx, &plan)
	if diags.HasError() {
		tflog.Debug(ctx, "Plan is delete, skipping")
		return
	}

	var state qemu.VirtualMachineResourceModel
	diags = req.State.Get(ctx, &state)
	if diags.HasError() {
		tflog.Debug(ctx, "Plan is create, skipping")
		return
	}

	// carry over computed values sets to prevent unnecessary diffs
	amended := plan
	amended.ComputedDisks = state.ComputedDisks
	amended.ComputedNetworkInterfaces = state.ComputedNetworkInterfaces
	amended.ComputedPCIDevices = state.ComputedPCIDevices

	diags = resp.Plan.Set(ctx, &amended)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
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

func (r *virtualMachineResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Debug(ctx, "Create virtual machine method")
	var plan qemu.VirtualMachineResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("plan '%v'", plan.CloudInit))

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

func (r *virtualMachineResource) readModelWithContext(ctx context.Context, node string, id int, state *qemu.VirtualMachineResourceModel) (*qemu.VirtualMachineResourceModel, error) {
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

	tflog.Debug(ctx, fmt.Sprintf("model '%v'", model))
	tflog.Debug(ctx, fmt.Sprintf("ci '%v'", model.CloudInit))
	tflog.Debug(ctx, fmt.Sprintf("dns '%v'", model.CloudInit.DNS))
	tflog.Debug(ctx, fmt.Sprintf("user '%v'", model.CloudInit.User))
	for _, ip := range model.CloudInit.IP.Configs {
		tflog.Debug(ctx, fmt.Sprintf("ip '%v'", ip))
		tflog.Debug(ctx, fmt.Sprintf("position '%v'", ip.Positition))
		tflog.Debug(ctx, fmt.Sprintf("v4 '%v'", ip.V4))
		tflog.Debug(ctx, fmt.Sprintf("v6 '%v'", ip.V6))
	}

	return model, nil
}

func (r *virtualMachineResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Debug(ctx, "Read virtual machine method")
	var state qemu.VirtualMachineResourceModel
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
	tflog.Debug(ctx, fmt.Sprintf("computed disks '%v'", m.ComputedDisks))
	m.ComputedDisks = state.ComputedDisks
	m.ComputedNetworkInterfaces = state.ComputedNetworkInterfaces
	m.ComputedPCIDevices = state.ComputedPCIDevices

	diags = resp.State.Set(ctx, &m)
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

	node := plan.Node.ValueString()
	vmId := int(state.ID.ValueInt64())

	wasRunning := false
	status, err := r.client.GetVirtualMachineStatus(ctx, node, vmId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error getting virtual machine state",
			"Could not configure virtual machine, unexpected error: "+err.Error(),
		)
		return
	}

	if status.Status == proxmox.VIRTUALMACHINESTATUS_RUNNING {
		wasRunning = true
		err = r.stopVm(ctx, node, vmId)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error stopping virtual machine",
				"Could not stop virtual machine, unexpected error: "+err.Error(),
			)
			return
		}
	}

	tflog.Debug(ctx, fmt.Sprintf("vm state '%v'", state))
	tflog.Debug(ctx, fmt.Sprintf("vm plan '%v'", plan))

	tflog.Debug(ctx, "Configuring virtual machine")
	err = r.configureVm(ctx, &plan)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error configuring virtual machine",
			"Could not configure virtual machine, unexpected error: "+err.Error(),
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

	if wasRunning {
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
	var state qemu.VirtualMachineResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.stopVm(ctx, state.Node.ValueString(), int(state.ID.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error stopping virtual machine",
			"Could not stop virtual machine, unexpected error: "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("Deleting vm: '%s' '%v'", state.Node.ValueString(), state.ID.ValueInt64()))
	err = r.client.DeleteVirtualMachine(ctx, state.Node.ValueString(), int(state.ID.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting vm",
			"Could not delete vm, unexpected error: "+err.Error(),
		)
		return
	}
}
