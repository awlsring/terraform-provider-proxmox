package nfs

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
	_ resource.Resource                = &NFSResource{}
	_ resource.ResourceWithConfigure   = &NFSResource{}
	_ resource.ResourceWithImportState = &NFSResource{}
)

var defaultContentTypes = []string{"iso", "images", "rootdir"}

func Resource() resource.Resource {
	return &NFSResource{}
}

type NFSResource struct {
	client *service.Proxmox
}

func (r *NFSResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_nfs_storage_class"
}

func (r *NFSResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resourceSchema
}

func (r *NFSResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*service.Proxmox)
}

func (r *NFSResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan nfsStorageClassModel
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

	request := &service.CreateNFSStorageClassInput{
		Id:           plan.ID.ValueString(),
		Server:       plan.Server.ValueString(),
		Export:       plan.Export.ValueString(),
		Nodes:        nodes,
		ContentTypes: contentTypes,
	}

	err = r.client.CreateNFSStorageClass(ctx, request)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating NFS storage class",
			"Could not create NFS storage class, unexpected error: "+err.Error(),
		)
		return
	}

	pool, err := r.readnfsStorageClassModel(ctx, plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading created NFS storage class",
			"Could not read created NFS storage class, unexpected error: "+err.Error(),
		)
		return
	}
	tflog.Debug(ctx, fmt.Sprintf("created NFS storage class with id '%s'", pool.ID.ValueString()))

	diags = resp.State.Set(ctx, &pool)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *NFSResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Debug(ctx, "Read NFS storage class method")
	var state nfsStorageClassModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	nfsStorageClassModel, err := r.readnfsStorageClassModel(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading NFS storage class",
			"Could not read NFS storage class, unexpected error: "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("NFS storage class '%v'", nfsStorageClassModel))

	diags = resp.State.Set(ctx, &nfsStorageClassModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *NFSResource) readnfsStorageClassModel(ctx context.Context, name string) (nfsStorageClassModel, error) {
	tflog.Debug(ctx, fmt.Sprintf("Reading NFS storage class: %s", name))
	s, err := r.client.GetNFSStorageClass(ctx, name)
	if err != nil {
		return nfsStorageClassModel{}, err
	}

	return NFSStorageClassToModel(s), nil
}

func (r *NFSResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	tflog.Debug(ctx, "Update bridge method")
	var plan nfsStorageClassModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state nfsStorageClassModel
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

	z := nfsStorageClassModel{
		ID:           plan.ID,
		Server:       plan.Server,
		Nodes:        utils.UnpackListType(nodes),
		ContentTypes: utils.UnpackListType(contentTypes),
		Mount:        plan.Mount,
		Export:       plan.Export,
	}

	tflog.Debug(ctx, fmt.Sprintf("state '%v'", state))
	tflog.Debug(ctx, fmt.Sprintf("plan '%v'", plan))

	tflog.Debug(ctx, fmt.Sprintf("Updating NFS storage class: '%s'", state.ID.ValueString()))
	tflog.Debug(ctx, fmt.Sprintf("new state: '%v'", z))
	err = r.client.ModifyNFSStorageClass(ctx, state.ID.ValueString(), &service.ModifyNFSStorageClassInput{
		Nodes:        nodes,
		ContentTypes: contentTypes,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating NFS storage class",
			"Could not update NFS storage class, unexpected error: "+err.Error(),
		)
		return
	}

	diags = resp.State.Set(ctx, &z)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *NFSResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	tflog.Debug(ctx, "Delete NFS storage class method")
	var state nfsStorageClassModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, fmt.Sprintf("NFS state '%v'", state))

	tflog.Debug(ctx, fmt.Sprintf("Deleting NFS: '%s'", state.ID.ValueString()))
	err := r.client.DeleteNFSSStorageClass(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting NFS storage class",
			"Could not delete NFS storage class, unexpected error: "+err.Error(),
		)
		return
	}
}

func (r *NFSResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)

	model, err := r.readnfsStorageClassModel(ctx, req.ID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading NFS storage class",
			"Could not read NFS storage class, unexpected error: "+err.Error(),
		)
		return
	}

	diags := resp.State.Set(ctx, model)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
