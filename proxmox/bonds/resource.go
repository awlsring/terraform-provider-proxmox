package bonds

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/awlsring/proxmox-go/proxmox"
	"github.com/awlsring/terraform-provider-proxmox/internal/service"
	"github.com/awlsring/terraform-provider-proxmox/proxmox/utils"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &bondResource{}
	_ resource.ResourceWithConfigure   = &bondResource{}
	_ resource.ResourceWithImportState = &bondResource{}
)

func NewResource() resource.Resource {
	return &bondResource{}
}

type bondResource struct {
	client *service.Proxmox
}

func formId(node string, name string) string {
	return fmt.Sprintf("%s/%s", node, name)
}

func unpackId(id string) (string, string, error) {
	s := strings.Split(id, "/")
	if len(s) != 2 {
		return "", "", fmt.Errorf("invalid id %s", id)
	}
	return s[0], s[1], nil
}

func (r *bondResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_network_bond"
}

func (r *bondResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The id of the bond. Formatted as `/{node}/{name}`.",
			},
			"node": schema.StringAttribute{
				Required:    true,
				Description: "The node the bond is on.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The name of the bond.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile("bond[0-9]$"), "name must follow scheme `bond<n>`"),
				},
			},
			"active": schema.BoolAttribute{
				Computed:    true,
				Description: "If the bond is active.",
			},
			"autostart": schema.BoolAttribute{
				Optional:    true,
				Description: "If the bond is set to autostart.",
			},
			"hash_policy": schema.StringAttribute{
				Optional:    true,
				Description: "Hash policy used on the bond.",
				Validators: []validator.String{
					stringvalidator.OneOf("layer2", "layer2+3", "layer3+4"),
				},
			},
			"bond_primary": schema.StringAttribute{
				Optional:    true,
				Description: "Primary interface on the bond.",
			},
			"mode": schema.StringAttribute{
				Required:    true,
				Description: "Mode of the bond.",
				Validators: []validator.String{
					stringvalidator.OneOf(
						"balance-rr",
						"active-backup",
						"balance-xor",
						"broadcast",
						"802.3ad",
						"balance-tlb",
						"balance-alb",
						"balance-slb",
						"lacp-balance-slb",
						"lacp-balance-tcp",
					),
				},
			},
			"comments": schema.StringAttribute{
				Optional:    true,
				Description: "Comment in the bond.",
			},
			"mii_mon": schema.StringAttribute{
				Computed:    true,
				Description: "Miimon of the bond.",
			},
			"interfaces": schema.ListAttribute{
				Required:    true,
				ElementType: types.StringType,
				Description: "List of interfaces on the bond.",
			},
		},
	}
}

func (r *bondResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*service.Proxmox)
}

func (r *bondResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan bondModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var name string
	var err error
	if plan.Name.IsUnknown() || plan.Name.IsNull() {
		tflog.Debug(ctx, "name is null, assinging next available name")
		name, err = r.generateName(ctx, plan.Node.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Error creating bond",
				"Could not create bond, unexpected error: "+err.Error(),
			)
			return
		}
	} else {
		tflog.Debug(ctx, fmt.Sprintf("using given name name '%s'", plan.Name.ValueString()))
		name = plan.Name.ValueString()
	}

	request := &service.CreateNetworkBondInput{
		Interfaces:  []string{},
		Name:        name,
		Node:        plan.Node.ValueString(),
		Mode:        proxmox.NetworkInterfaceBondMode(plan.Mode.ValueString()),
		BondPrimary: utils.OptionalToPointerString(plan.BondPrimary.ValueString()),
		AutoStart:   utils.OptionalToPointerBool(plan.Autostart.ValueBool()),
		Comments:    utils.OptionalToPointerString(plan.Comments.ValueString()),
	}

	for _, i := range plan.Interfaces {
		request.Interfaces = append(request.Interfaces, i.ValueString())
	}

	if !plan.HashPolicy.IsNull() {
		h := proxmox.NetworkInterfaceBondHashPolicy(plan.HashPolicy.ValueString())
		request.HashPolicy = &h
	}

	err = r.client.CreateNetworkBond(ctx, request)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating bond",
			"Could not create bond, unexpected error: "+err.Error(),
		)
		return
	}

	var bond bondModel
	tries := 0
	for {
		bond, err = r.readBondModel(ctx, formId(plan.Node.ValueString(), name))
		if err != nil {
			if tries < 3 {
				tflog.Debug(ctx, "could not read bond, retrying...")
				tries += 1
				time.Sleep(time.Duration(tries) * time.Second)
				continue
			}
			resp.Diagnostics.AddError(
				"Error creating bond",
				"Could not read create bond, unexpected error: "+err.Error(),
			)
			return
		}
		break
	}
	bond.Active = types.BoolValue(true)
	tflog.Debug(ctx, fmt.Sprintf("created bond with id '%s'", bond.ID.ValueString()))

	diags = resp.State.Set(ctx, &bond)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *bondResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Debug(ctx, "Read pool method")
	var state bondModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	bondModel, err := r.readBondModel(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading pool",
			"Could not read pool, unexpected error: "+err.Error(),
		)
		return
	}

	diags = resp.State.Set(ctx, &bondModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *bondResource) generateName(ctx context.Context, node string) (string, error) {
	tflog.Debug(ctx, "Generating name")
	bonds, err := r.client.ListNetworkBonds(ctx, node)
	if err != nil {
		return "", err
	}

	bondsNames := []string{}
	for _, bond := range bonds {
		bondsNames = append(bondsNames, bond.Iface)
	}

	if len(bonds) == 0 {
		return "bond0", err
	}

	var bondNumbers []int
	re := regexp.MustCompile(`\d+`)
	for _, bond := range bondsNames {
		bondNumber, _ := strconv.Atoi(re.FindString(bond))
		bondNumbers = append(bondNumbers, bondNumber)
	}
	sort.Ints(bondNumbers)

	for i := 0; i < bondNumbers[len(bondNumbers)-1]+1; i++ {
		if !utils.Contains(bondNumbers, i) {
			return fmt.Sprintf("bond%d", i), nil
		}
	}
	n := bondNumbers[len(bondNumbers)-1] + 1
	return fmt.Sprintf("bond%d", n), nil
}

func (r *bondResource) readBondModel(ctx context.Context, id string) (bondModel, error) {
	tflog.Debug(ctx, fmt.Sprintf("Reading bond model: %s", id))

	node, name, err := unpackId(id)
	if err != nil {
		return bondModel{}, err
	}

	bond, err := r.client.GetNetworkBond(ctx, node, name)
	if err != nil {
		return bondModel{}, err
	}

	return BondToModel(bond), nil
}

func (r *bondResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan bondModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state bondModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	bond := bondModel{
		ID:         state.ID,
		Name:       state.Name,
		Node:       state.Node,
		Interfaces: plan.Interfaces,
		Mode:       plan.Mode,
		Autostart:  types.BoolValue(true),
		Active:     plan.Active,
		MiiMon:     plan.MiiMon,
	}

	if !plan.HashPolicy.IsNull() {
		bond.HashPolicy = plan.HashPolicy
	}

	if !plan.BondPrimary.IsNull() {
		bond.BondPrimary = plan.BondPrimary
	}

	if !plan.Autostart.IsNull() {
		bond.Autostart = plan.Autostart
	}

	if !plan.Comments.IsNull() {
		bond.Comments = plan.Comments
	}

	request := &service.UpdateNetworkBondInput{
		Interfaces:  []string{},
		Name:        bond.Name.ValueString(),
		Node:        bond.Node.ValueString(),
		Mode:        proxmox.NetworkInterfaceBondMode(bond.Mode.ValueString()),
		BondPrimary: utils.OptionalToPointerString(bond.BondPrimary.ValueString()),
		AutoStart:   utils.OptionalToPointerBool(bond.Autostart.ValueBool()),
		Comments:    utils.OptionalToPointerString(bond.Comments.ValueString()),
	}

	for _, i := range bond.Interfaces {
		request.Interfaces = append(request.Interfaces, i.ValueString())
	}

	if !plan.HashPolicy.IsNull() {
		h := proxmox.NetworkInterfaceBondHashPolicy(plan.HashPolicy.ValueString())
		bond.HashPolicy = types.StringValue(string(h))
		request.HashPolicy = &h
	}

	err := r.client.UpdateNetworkBond(ctx, request)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating bond",
			"Could not create bond, unexpected error: "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("Updating bond: '%s' '%s'", state.Node.ValueString(), state.Name.ValueString()))

	diags = resp.State.Set(ctx, &bond)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *bondResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	tflog.Debug(ctx, "Delete bond method")
	var state bondModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("Deleting bond: '%s' '%s'", state.Node.ValueString(), state.Name.ValueString()))
	err := r.client.DeleteNetworkBond(ctx, state.Node.ValueString(), state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting pool",
			"Could not delete pool, unexpected error: "+err.Error(),
		)
		return
	}
}

func (r *bondResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)

	model, err := r.readBondModel(ctx, req.ID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading pool",
			"Could not read pool, unexpected error: "+err.Error(),
		)
		return
	}

	diags := resp.State.Set(ctx, model)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
