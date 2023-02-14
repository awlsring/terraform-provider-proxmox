package lvmthin

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
	_ resource.Resource                = &lvmThinpoolResource{}
	_ resource.ResourceWithConfigure   = &lvmThinpoolResource{}
	_ resource.ResourceWithImportState = &lvmThinpoolResource{}
)

func Resource() resource.Resource {
	return &lvmThinpoolResource{}
}

type lvmThinpoolResource struct {
	client *service.Proxmox
}

func (r *lvmThinpoolResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_lvm_thinpool"
}

func (r *lvmThinpoolResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resourceSchema
}

func (r *lvmThinpoolResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*service.Proxmox)
}

func (r *lvmThinpoolResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan lvmThinpoolModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	request := &service.CreateLVMThinpoolInput{
		Name:   plan.Name.ValueString(),
		Node:   plan.Node.ValueString(),
		Device: plan.Device.ValueString(),
	}

	err := r.client.CreateLVMThinpool(ctx, request)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating LVM thinpool",
			"Could not create LVM thinpool, unexpected error: "+err.Error(),
		)
		return
	}

	var pool lvmThinpoolModel
	tries := 0
	for {
		pool, err = r.readLVMThinpoolModel(ctx, network.FormId(plan.Node.ValueString(), plan.Name.ValueString()))
		if err != nil {
			// this can take a while to create, so we'll retry a LOT for now
			if tries < 50 {
				tries += 1
				tflog.Warn(ctx, fmt.Sprintf("Attempt %d caught error. Waiting %d second then retrying", tries, tries))
				time.Sleep(time.Duration(tries) * time.Second)
				continue
			}
			resp.Diagnostics.AddError(
				"Error reading created LVM thinpool",
				"Could not read created LVM thipool, unexpected error: "+err.Error(),
			)
			return
		}
		break
	}
	tflog.Debug(ctx, fmt.Sprintf("created LVM thinpool with id '%s'", pool.ID.ValueString()))

	diags = resp.State.Set(ctx, &pool)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *lvmThinpoolResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Debug(ctx, "Read LVM thinpool method")
	var state lvmThinpoolModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	lvmThinpoolModel, err := r.readLVMThinpoolModel(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading LVM thinpool",
			"Could not read LVM thinpool, unexpected error: "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("LVM thinpool model '%v'", lvmThinpoolModel))

	diags = resp.State.Set(ctx, &lvmThinpoolModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *lvmThinpoolResource) readLVMThinpoolModel(ctx context.Context, id string) (lvmThinpoolModel, error) {
	tflog.Debug(ctx, fmt.Sprintf("Reading LVM thinpool model: %s", id))

	node, name, err := utils.UnpackId(id)
	if err != nil {
		return lvmThinpoolModel{}, err
	}

	pool, err := r.client.GetLVMThinpool(ctx, node, name)
	if err != nil {
		return lvmThinpoolModel{}, err
	}

	return LVMThinpoolToModel(pool), nil
}

func (r *lvmThinpoolResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Unexpected update!",
		"Unexpected update request, resource cannot update.",
	)
}

func (r *lvmThinpoolResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	tflog.Debug(ctx, "Delete LVM thinpool method")
	var state lvmThinpoolModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, fmt.Sprintf("lvm thinpool state '%v'", state))

	_, vg, err := utils.UnpackId(state.VolumeGroup.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting LVM thinpool",
			"Could not delete LVM thinpool, unexpected error: "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("Deleting lvm thinpool: '%s' '%s'", state.Node.ValueString(), state.Name.ValueString()))
	err = r.client.DeleteLVMThinpool(ctx, state.Node.ValueString(), state.Name.ValueString(), vg)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting LVM thinpool",
			"Could not delete LVM thinpool, unexpected error: "+err.Error(),
		)
		return
	}
}

func (r *lvmThinpoolResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)

	model, err := r.readLVMThinpoolModel(ctx, req.ID)
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
