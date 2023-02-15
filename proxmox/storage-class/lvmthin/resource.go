package lvmthin

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
	_ resource.Resource                = &LVMThinResource{}
	_ resource.ResourceWithConfigure   = &LVMThinResource{}
	_ resource.ResourceWithImportState = &LVMThinResource{}
)

var defaultContentTypes = []string{"images", "rootdir"}

func Resource() resource.Resource {
	return &LVMThinResource{}
}

type LVMThinResource struct {
	client *service.Proxmox
}

func (r *LVMThinResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_lvm_thinpool_storage_class"
}

func (r *LVMThinResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resourceSchema
}

func (r *LVMThinResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*service.Proxmox)
}

func (r *LVMThinResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan lvmThinStorageClassModel
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

	request := &service.CreateLVMThinStorageClassInput{
		Id:           plan.ID.ValueString(),
		Nodes:        nodes,
		VolumeGroup:  plan.VolumeGroup.ValueString(),
		Thinpool:     plan.Thinpool.ValueString(),
		ContentTypes: contentTypes,
	}

	err = r.client.CreateLVMThinStorageClass(ctx, request)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating LVM Thinpool storage class",
			"Could not create LVM Thinpool storage class, unexpected error: "+err.Error(),
		)
		return
	}

	pool, err := r.readlvmThinStorageClassModel(ctx, plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading created LVM Thinpool storage class",
			"Could not read created LVM Thinpool storage class, unexpected error: "+err.Error(),
		)
		return
	}
	tflog.Debug(ctx, fmt.Sprintf("created LVM Thinpool storage class with id '%s'", pool.ID.ValueString()))

	diags = resp.State.Set(ctx, &pool)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *LVMThinResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Debug(ctx, "Read LVM Thinpool storage class method")
	var state lvmThinStorageClassModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	lvmThinStorageClassModel, err := r.readlvmThinStorageClassModel(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading LVM Thinpool storage class",
			"Could not read LVM Thinpool storage class, unexpected error: "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("LVM Thinpool storage class '%v'", lvmThinStorageClassModel))

	diags = resp.State.Set(ctx, &lvmThinStorageClassModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *LVMThinResource) readlvmThinStorageClassModel(ctx context.Context, name string) (lvmThinStorageClassModel, error) {
	tflog.Debug(ctx, fmt.Sprintf("Reading LVM Thinpool storage class: %s", name))
	s, err := r.client.GetLVMThinStorageClass(ctx, name)
	if err != nil {
		return lvmThinStorageClassModel{}, err
	}

	return LVMThinStorageClassToModel(s), nil
}

func (r *LVMThinResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	tflog.Debug(ctx, "Update bridge method")
	var plan lvmThinStorageClassModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state lvmThinStorageClassModel
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

	z := lvmThinStorageClassModel{
		ID:           plan.ID,
		Nodes:        utils.UnpackListType(nodes),
		ContentTypes: utils.UnpackListType(contentTypes),
	}

	tflog.Debug(ctx, fmt.Sprintf("state '%v'", state))
	tflog.Debug(ctx, fmt.Sprintf("plan '%v'", plan))

	tflog.Debug(ctx, fmt.Sprintf("Updating LVM Thinpool storage class: '%s'", state.ID.ValueString()))
	tflog.Debug(ctx, fmt.Sprintf("new state: '%v'", z))
	err = r.client.ModifyLVMThinStorageClass(ctx, state.ID.ValueString(), &service.ModifyLVMThinStorageClassInput{
		Nodes:        nodes,
		ContentTypes: contentTypes,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating LVM Thinpool storage class",
			"Could not update LVM Thinpool storage class, unexpected error: "+err.Error(),
		)
		return
	}

	diags = resp.State.Set(ctx, &z)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *LVMThinResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	tflog.Debug(ctx, "Delete LVM Thinpool storage class method")
	var state lvmThinStorageClassModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, fmt.Sprintf("LVM Thinpool state '%v'", state))

	tflog.Debug(ctx, fmt.Sprintf("Deleting LVM Thinpool: '%s'", state.ID.ValueString()))
	err := r.client.DeleteLVMThinSStorageClass(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting LVM Thinpool storage class",
			"Could not delete LVM Thinpool storage class, unexpected error: "+err.Error(),
		)
		return
	}
}

func (r *LVMThinResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)

	model, err := r.readlvmThinStorageClassModel(ctx, req.ID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading LVM Thinpool storage class",
			"Could not read LVM Thinpool storage class, unexpected error: "+err.Error(),
		)
		return
	}

	diags := resp.State.Set(ctx, model)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
