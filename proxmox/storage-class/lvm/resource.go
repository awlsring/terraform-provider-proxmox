package lvm

import (
	"context"
	"fmt"

	"github.com/awlsring/terraform-provider-proxmox/internal/service"
	"github.com/awlsring/terraform-provider-proxmox/proxmox/storage-class"
	"github.com/awlsring/terraform-provider-proxmox/proxmox/utils"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &LVMResource{}
	_ resource.ResourceWithConfigure   = &LVMResource{}
	_ resource.ResourceWithImportState = &LVMResource{}
)

var defaultContentTypes = []string{"images", "rootdir"}

func Resource() resource.Resource {
	return &LVMResource{}
}

type LVMResource struct {
	client *service.Proxmox
}

func (r *LVMResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_lvm_storage_class"
}

func (r *LVMResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resourceSchema
}

func (r *LVMResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*service.Proxmox)
}

func (r *LVMResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan lvmStorageClassModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	nodes, err := storage.DetermineNodes(ctx, r.client, plan.Nodes)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error determining nodes",
			"Could not determine nodes, unexpected error: "+err.Error(),
		)
		return
	}
	tflog.Debug(ctx, fmt.Sprintf("nodes '%v'", nodes))

	contentTypes, err := storage.DetermineContentTypes(ctx, plan.ContentTypes, defaultContentTypes)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error determining content types",
			"Could not determine content types, unexpected error: "+err.Error(),
		)
		return
	}

	request := &service.CreateLVMStorageClassInput{
		Id:           plan.ID.ValueString(),
		Nodes:        nodes,
		VolumeGroup:  plan.VolumeGroup.ValueString(),
		ContentTypes: contentTypes,
	}

	err = r.client.CreateLVMStorageClass(ctx, request)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating LVM storage class",
			"Could not create LVM storage class, unexpected error: "+err.Error(),
		)
		return
	}

	pool, err := r.readlvmStorageClassModel(ctx, plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading created LVM storage class",
			"Could not read created LVM storage class, unexpected error: "+err.Error(),
		)
		return
	}
	tflog.Debug(ctx, fmt.Sprintf("created LVM storage class with id '%s'", pool.ID.ValueString()))

	diags = resp.State.Set(ctx, &pool)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *LVMResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Debug(ctx, "Read LVM storage class method")
	var state lvmStorageClassModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	lvmStorageClassModel, err := r.readlvmStorageClassModel(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading LVM storage class",
			"Could not read LVM storage class, unexpected error: "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("LVM storage class '%v'", lvmStorageClassModel))

	diags = resp.State.Set(ctx, &lvmStorageClassModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *LVMResource) readlvmStorageClassModel(ctx context.Context, name string) (lvmStorageClassModel, error) {
	tflog.Debug(ctx, fmt.Sprintf("Reading LVM storage class: %s", name))
	s, err := r.client.GetLVMStorageClass(ctx, name)
	if err != nil {
		return lvmStorageClassModel{}, err
	}

	return LVMStorageClassToModel(s), nil
}

func (r *LVMResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	tflog.Debug(ctx, "Update bridge method")
	var plan lvmStorageClassModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state lvmStorageClassModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	nodes, err := storage.DetermineNodes(ctx, r.client, plan.Nodes)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error determining nodes",
			"Could not determine nodes, unexpected error: "+err.Error(),
		)
		return
	}
	tflog.Debug(ctx, fmt.Sprintf("nodes '%v'", nodes))

	contentTypes, err := storage.DetermineContentTypes(ctx, plan.ContentTypes, defaultContentTypes)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error determining content types",
			"Could not determine content types, unexpected error: "+err.Error(),
		)
		return
	}
	tflog.Debug(ctx, fmt.Sprintf("content types '%v'", contentTypes))

	z := lvmStorageClassModel{
		ID:           plan.ID,
		Nodes:        utils.UnpackListType(nodes),
		ContentTypes: utils.UnpackListType(contentTypes),
	}

	tflog.Debug(ctx, fmt.Sprintf("state '%v'", state))
	tflog.Debug(ctx, fmt.Sprintf("plan '%v'", plan))

	tflog.Debug(ctx, fmt.Sprintf("Updating LVM storage class: '%s'", state.ID.ValueString()))
	tflog.Debug(ctx, fmt.Sprintf("new state: '%v'", z))
	err = r.client.ModifyLVMStorageClass(ctx, state.ID.ValueString(), &service.ModifyLVMStorageClassInput{
		Nodes:        nodes,
		ContentTypes: contentTypes,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating LVM storage class",
			"Could not update LVM storage class, unexpected error: "+err.Error(),
		)
		return
	}

	diags = resp.State.Set(ctx, &z)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *LVMResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	tflog.Debug(ctx, "Delete LVM storage class method")
	var state lvmStorageClassModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, fmt.Sprintf("LVM state '%v'", state))

	tflog.Debug(ctx, fmt.Sprintf("Deleting LVM: '%s'", state.ID.ValueString()))
	err := r.client.DeleteLVMSStorageClass(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting LVM storage class",
			"Could not delete LVM storage class, unexpected error: "+err.Error(),
		)
		return
	}
}

func (r *LVMResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)

	model, err := r.readlvmStorageClassModel(ctx, req.ID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading LVM storage class",
			"Could not read LVM storage class, unexpected error: "+err.Error(),
		)
		return
	}

	diags := resp.State.Set(ctx, model)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
