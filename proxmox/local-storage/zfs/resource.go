package zfs

import (
	"context"
	"fmt"
	"time"

	"github.com/awlsring/proxmox-go/proxmox"
	"github.com/awlsring/terraform-provider-proxmox/internal/service"
	"github.com/awlsring/terraform-provider-proxmox/proxmox/network"
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

func Resource() resource.Resource {
	return &zfsResource{}
}

type zfsResource struct {
	client *service.Proxmox
}

func (r *zfsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_zfs_pool"
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
	var plan zfsModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	request := &service.CreateZFSPoolInput{
		Name:      plan.Name.ValueString(),
		Node:      plan.Node.ValueString(),
		RaidLevel: proxmox.ZFSRaidLevel(plan.RaidLevel.ValueString()),
	}

	for _, d := range plan.Disks {
		request.Disks = append(request.Disks, d.ValueString())
	}

	err := r.client.CreateZFSPool(ctx, request)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating ZFS pool",
			"Could not create ZFS pool, unexpected error: "+err.Error(),
		)
		return
	}

	var pool zfsModel
	tries := 0
	for {
		pool, err = r.readZFSModel(ctx, network.FormId(plan.Node.ValueString(), plan.Name.ValueString()))
		if err != nil {
			if tries < 10 {
				tries += 1
				tflog.Warn(ctx, fmt.Sprintf("Attempt %d caught error. Waiting %d second then retrying", tries, tries))
				time.Sleep(time.Duration(tries) * time.Second)
				continue
			}
			resp.Diagnostics.AddError(
				"Error reading created ZFS pool",
				"Could not read created ZFS pool, unexpected error: "+err.Error(),
			)
			return
		}
		break
	}
	pool.Disks = plan.Disks

	tflog.Debug(ctx, fmt.Sprintf("created ZFS pool with id '%s'", pool.ID.ValueString()))

	diags = resp.State.Set(ctx, &pool)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *zfsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Debug(ctx, "Read bond method")
	var state zfsModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	zfsModel, err := r.readZFSModel(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading ZFS pool",
			"Could not read ZFS pool, unexpected error: "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("ZFS model '%v'", zfsModel))

	diags = resp.State.Set(ctx, &zfsModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *zfsResource) readZFSModel(ctx context.Context, id string) (zfsModel, error) {
	tflog.Debug(ctx, fmt.Sprintf("Reading ZFS model: %s", id))

	node, name, err := utils.UnpackId(id)
	if err != nil {
		return zfsModel{}, err
	}

	pool, err := r.client.DescribeZFSPool(ctx, node, name)
	if err != nil {
		return zfsModel{}, err
	}

	return ZFSToModel(pool), nil
}

func (r *zfsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Unexpected update!",
		"Unexpected update request, resource cannot update.",
	)
}

func (r *zfsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	tflog.Debug(ctx, "Delete ZFS Pool method")
	var state zfsModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, fmt.Sprintf("zfs state '%v'", state))

	tflog.Debug(ctx, fmt.Sprintf("Deleting zfs: '%s' '%s'", state.Node.ValueString(), state.Name.ValueString()))
	err := r.client.DeleteZFSPool(ctx, state.Node.ValueString(), state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting ZFS pool",
			"Could not delete ZFS pool, unexpected error: "+err.Error(),
		)
		return
	}
}

func (r *zfsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)

	model, err := r.readZFSModel(ctx, req.ID)
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
