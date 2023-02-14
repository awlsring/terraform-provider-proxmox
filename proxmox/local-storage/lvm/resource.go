package lvm

import (
	"context"
	"fmt"
	"time"

	"github.com/awlsring/terraform-provider-proxmox/internal/service"
	"github.com/awlsring/terraform-provider-proxmox/proxmox/network"
	"github.com/awlsring/terraform-provider-proxmox/proxmox/utils"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &lvmResource{}
	_ resource.ResourceWithConfigure   = &lvmResource{}
	_ resource.ResourceWithImportState = &lvmResource{}
)

func Resource() resource.Resource {
	return &lvmResource{}
}

type lvmResource struct {
	client *service.Proxmox
}

func (r *lvmResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_lvm"
}

func (r *lvmResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resourceSchema
}

func (r *lvmResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*service.Proxmox)
}

func (r *lvmResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan lvmModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	request := &service.CreateLVMInput{
		Name:   plan.Name.ValueString(),
		Node:   plan.Node.ValueString(),
		Device: plan.Device.ValueString(),
	}

	err := r.client.CreateLVM(ctx, request)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating LVM",
			"Could not create LVM, unexpected error: "+err.Error(),
		)
		return
	}

	var pool lvmModel
	tries := 0
	for {
		pool, err = r.readLVMModel(ctx, network.FormId(plan.Node.ValueString(), plan.Name.ValueString()))
		if err != nil {
			if tries < 10 {
				tries += 1
				tflog.Warn(ctx, fmt.Sprintf("Attempt %d caught error. Waiting %d second then retrying", tries, tries))
				time.Sleep(time.Duration(tries) * time.Second)
				continue
			}
			resp.Diagnostics.AddError(
				"Error reading created LVM",
				"Could not read created LVM, unexpected error: "+err.Error(),
			)
			return
		}
		break
	}
	tflog.Debug(ctx, fmt.Sprintf("created LVM with id '%s'", pool.ID.ValueString()))

	diags = resp.State.Set(ctx, &pool)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *lvmResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Debug(ctx, "Read LVM method")
	var state lvmModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	lvmModel, err := r.readLVMModel(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading LVM",
			"Could not read LVM, unexpected error: "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("LVM model '%v'", lvmModel))

	diags = resp.State.Set(ctx, &lvmModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *lvmResource) readLVMModel(ctx context.Context, id string) (lvmModel, error) {
	tflog.Debug(ctx, fmt.Sprintf("Reading LVM model: %s", id))

	node, name, err := utils.UnpackId(id)
	if err != nil {
		return lvmModel{}, err
	}

	lvm, err := r.client.GetLVM(ctx, node, name)
	if err != nil {
		return lvmModel{}, err
	}

	return LVMToModel(lvm), nil
}

func (r *lvmResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Unexpected update!",
		"Unexpected update request, resource cannot update.",
	)
}

func (r *lvmResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	tflog.Debug(ctx, "Delete LVM method")
	var state lvmModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, fmt.Sprintf("lvm state '%v'", state))

	tflog.Debug(ctx, fmt.Sprintf("Deleting lvm: '%s' '%s'", state.Node.ValueString(), state.Name.ValueString()))
	err := r.client.DeleteLVM(ctx, state.Node.ValueString(), state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting LVM",
			"Could not delete LVM, unexpected error: "+err.Error(),
		)
		return
	}
}

func (r *lvmResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)

	model, err := r.readLVMModel(ctx, req.ID)
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
