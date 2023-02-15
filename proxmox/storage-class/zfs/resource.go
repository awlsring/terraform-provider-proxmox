package zfs

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
	_ resource.Resource                = &zfsResource{}
	_ resource.ResourceWithConfigure   = &zfsResource{}
	_ resource.ResourceWithImportState = &zfsResource{}
)

var defaultContentTypes = []string{"images", "rootdir"}

func Resource() resource.Resource {
	return &zfsResource{}
}

type zfsResource struct {
	client *service.Proxmox
}

func (r *zfsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_zfs_storage_class"
}

func (r *zfsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resourceSchema
}

func (r *zfsResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*service.Proxmox)
}

func (r *zfsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan storageClassZfsModel
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

	request := &service.CreateZFSStorageClassInput{
		Id:           plan.ID.ValueString(),
		Pool:         plan.Pool.ValueString(),
		Nodes:        nodes,
		ContentTypes: contentTypes,
	}

	err = r.client.CreateZFSStorageClass(ctx, request)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating ZFS storage class",
			"Could not create ZFS storage class, unexpected error: "+err.Error(),
		)
		return
	}

	pool, err := r.readstorageClassZfsModel(ctx, plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading created zfs storage class",
			"Could not read created zfs storage class, unexpected error: "+err.Error(),
		)
		return
	}
	tflog.Debug(ctx, fmt.Sprintf("created zfs storage class with id '%s'", pool.ID.ValueString()))

	diags = resp.State.Set(ctx, &pool)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *zfsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Debug(ctx, "Read zfs storage class method")
	var state storageClassZfsModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	storageClassZfsModel, err := r.readstorageClassZfsModel(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading zfs storage class",
			"Could not read zfs storage class, unexpected error: "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("zfs storage class '%v'", storageClassZfsModel))

	diags = resp.State.Set(ctx, &storageClassZfsModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *zfsResource) readstorageClassZfsModel(ctx context.Context, name string) (storageClassZfsModel, error) {
	tflog.Debug(ctx, fmt.Sprintf("Reading zfs storage class: %s", name))
	s, err := r.client.GetZFSStorageClass(ctx, name)
	if err != nil {
		return storageClassZfsModel{}, err
	}

	return ZFSStorageClassToModel(s), nil
}

func (r *zfsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	tflog.Debug(ctx, "Update bridge method")
	var plan storageClassZfsModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state storageClassZfsModel
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

	z := storageClassZfsModel{
		ID:           plan.ID,
		Pool:         plan.Pool,
		Mount:        plan.Mount,
		Nodes:        utils.UnpackListType(nodes),
		ContentTypes: utils.UnpackListType(contentTypes),
	}

	tflog.Debug(ctx, fmt.Sprintf("state '%v'", state))
	tflog.Debug(ctx, fmt.Sprintf("plan '%v'", plan))

	tflog.Debug(ctx, fmt.Sprintf("Updating zfs storage class: '%s'", state.ID.ValueString()))
	tflog.Debug(ctx, fmt.Sprintf("new state: '%v'", z))
	err = r.client.ModifyZFSStorageClass(ctx, state.ID.ValueString(), &service.ModifyZFSStorageClassInput{
		Nodes:        nodes,
		ContentTypes: contentTypes,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating zfs storage class",
			"Could not update zfs storage class, unexpected error: "+err.Error(),
		)
		return
	}

	diags = resp.State.Set(ctx, &z)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *zfsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	tflog.Debug(ctx, "Delete zfs storage class method")
	var state storageClassZfsModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, fmt.Sprintf("zfs state '%v'", state))

	tflog.Debug(ctx, fmt.Sprintf("Deleting zfs: '%s'", state.ID.ValueString()))
	err := r.client.DeleteZFSSStorageClass(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting zfs storage class",
			"Could not delete zfs storage class, unexpected error: "+err.Error(),
		)
		return
	}
}

func (r *zfsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)

	model, err := r.readstorageClassZfsModel(ctx, req.ID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading zfs storage class",
			"Could not read zfs storage class, unexpected error: "+err.Error(),
		)
		return
	}

	diags := resp.State.Set(ctx, model)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
