package resource_pools

import (
	"context"
	"fmt"
	"strconv"

	"github.com/awlsring/proxmox-go/proxmox"
	"github.com/awlsring/terraform-provider-proxmox/internal/service"
	"github.com/awlsring/terraform-provider-proxmox/proxmox/utils"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &poolResource{}
	_ resource.ResourceWithConfigure   = &poolResource{}
	_ resource.ResourceWithImportState = &poolResource{}
)

func Resource() resource.Resource {
	return &poolResource{}
}

type poolResource struct {
	client *service.Proxmox
}

func (r *poolResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_resource_pool"
}

func (r *poolResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:    true,
				Description: "The id of the resource pool.",
			},
			"comment": schema.StringAttribute{
				Optional:    true,
				Description: "Notes on the resource pool.",
			},
			// can be consolidated with data source
			"members": schema.ListNestedAttribute{
				Optional:    true,
				Description: "Resources that are part of the resource pool.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Required:    true,
							Description: "The id of the resource.",
						},
						"type": schema.StringAttribute{
							Required:    true,
							Description: "The type of the resource.",
						},
					},
				},
			},
		},
	}
}

func (r *poolResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*service.Proxmox)
}

func (r *poolResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan poolModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.CreatePool(ctx, &service.CreatePoolInput{
		PoolId:  plan.ID.ValueString(),
		Comment: utils.OptionalToPointerString(plan.Comment.ValueString()),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating resource pool",
			"Could not create resource pool, unexpected error: "+err.Error(),
		)
		return
	}

	if len(plan.Members) != 0 {
		vms := []int{}
		storage := []string{}

		for _, member := range plan.Members {
			switch member.Type.ValueString() {
			case "qemu":
				vmid, err := strconv.Atoi(member.ID.ValueString())
				if err != nil {
					continue
				}
				vms = append(vms, vmid)
			case "storage":
				storage = append(storage, member.ID.ValueString())
			}
		}

		err := r.client.UpdatePool(ctx, &service.UpdatePoolInput{
			PoolId:  plan.ID.ValueString(),
			Vms:     vms,
			Storage: storage,
		})

		if err != nil {
			resp.Diagnostics.AddError(
				"Error creating resource pool",
				"Could not add members to resource pool, unexpected error: "+err.Error(),
			)
			return
		}
	}

	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating resource pool",
			"Could not create resource pool, unexpected error: "+err.Error(),
		)
		return
	}

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *poolResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Debug(ctx, "Read resource pool method")
	var state poolModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	poolModel, err := r.readPoolModel(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading resource pool",
			"Could not read resource pool, unexpected error: "+err.Error(),
		)
		return
	}

	diags = resp.State.Set(ctx, &poolModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *poolResource) readPoolModel(ctx context.Context, id string) (poolModel, error) {
	model := poolModel{
		ID: types.StringValue(id),
	}

	pool, err := r.client.GetPool(ctx, id)
	if err != nil {
		return poolModel{}, err
	}

	model.Comment = types.StringValue(utils.PtrStringToString(pool.Comment))

	model.Members = make([]poolMemberModel, len(pool.Members))
	for i, member := range pool.Members {
		var id string
		t := member.Type

		switch t {
		case proxmox.POOLMEMBERTYPE_QEMU:
			id = strconv.Itoa(int(*member.Vmid))
		case proxmox.POOLMEMBERTYPE_STORAGE:
			id = member.Id
		}

		model.Members[i] = poolMemberModel{
			ID:   types.StringValue(id),
			Type: types.StringValue(string(t)),
		}
	}

	return model, nil
}

func (r *poolResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan poolModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state poolModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	newMembers, removedMembers := determineNewAndRemovedMembers(state.Members, plan.Members)
	tflog.Debug(ctx, fmt.Sprint("New members: ", len(newMembers)))
	tflog.Debug(ctx, fmt.Sprint("Removed members: ", len(removedMembers)))

	var comment *string
	if plan.Comment != state.Comment {
		comment = nil
	}

	// Add members
	if len(newMembers) > 0 {
		tflog.Debug(ctx, "Adding members to resource pool")
		err := r.changePoolMembers(ctx, plan.ID.ValueString(), comment, newMembers, false)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error updating resource pool",
				"Could not update resource pool, unexpected error: "+err.Error(),
			)
			return
		}
	}

	// Remove members
	if len(removedMembers) > 0 {
		tflog.Debug(ctx, "Removing members to resource pool")
		err := r.changePoolMembers(ctx, plan.ID.ValueString(), comment, removedMembers, true)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error updating resource pool",
				"Could not update resource pool, unexpected error: "+err.Error(),
			)
			return
		}
	}

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func determineNewAndRemovedMembers(previous []poolMemberModel, current []poolMemberModel) ([]poolMemberModel, []poolMemberModel) {
	previousMembersMap := make(map[poolMemberModel]bool)
	for _, member := range previous {
		previousMembersMap[member] = true
	}

	currentMembersMap := make(map[poolMemberModel]bool)
	for _, member := range current {
		currentMembersMap[member] = true
	}

	newMembers := []poolMemberModel{}
	removedMembers := []poolMemberModel{}

	for _, member := range current {
		if _, exists := previousMembersMap[member]; !exists {
			newMembers = append(newMembers, member)
		}
	}

	for _, member := range previous {
		if _, exists := currentMembersMap[member]; !exists {
			removedMembers = append(removedMembers, member)
		}
	}

	return newMembers, removedMembers
}

func (r *poolResource) changePoolMembers(ctx context.Context, poolId string, comment *string, members []poolMemberModel, remove bool) error {
	vmMembers := []int{}
	storageMembers := []string{}
	for _, member := range members {
		switch member.Type.ValueString() {
		case string(proxmox.POOLMEMBERTYPE_QEMU):
			vmid, err := strconv.Atoi(member.ID.ValueString())
			if err != nil {
				return err
			}
			vmMembers = append(vmMembers, vmid)
		case string(proxmox.POOLMEMBERTYPE_STORAGE):
			storageMembers = append(storageMembers, member.ID.ValueString())
		}
	}
	err := r.client.UpdatePool(ctx, &service.UpdatePoolInput{
		PoolId:  poolId,
		Comment: comment,
		Delete:  remove,
		Vms:     vmMembers,
		Storage: storageMembers,
	})
	if err != nil {
		return err
	}

	return nil
}

func (r *poolResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state poolModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// remove members from pool
	err := r.changePoolMembers(ctx, state.ID.ValueString(), nil, state.Members, true)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting resource pool",
			"Could not delete resource pool, unexpected error: "+err.Error(),
		)
		return
	}

	// then delete pool
	err = r.client.DeletePool(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting resource pool",
			"Could not delete resource pool, unexpected error: "+err.Error(),
		)
		return
	}
}

func (r *poolResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
	id := req.ID
	model, err := r.readPoolModel(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading resource pool",
			"Could not read resource pool, unexpected error: "+err.Error(),
		)
		return
	}

	diags := resp.State.Set(ctx, model)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
