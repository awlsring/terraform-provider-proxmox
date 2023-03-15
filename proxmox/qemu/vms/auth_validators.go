package vms

import (
	"context"

	vt "github.com/awlsring/terraform-provider-proxmox/proxmox/qemu/vms/types"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func authCreateValidator(ctx context.Context, isRoot bool, model *vt.VirtualMachineResourceModel, resp *resource.ModifyPlanResponse) {
	tflog.Debug(ctx, "authCreateValidator")
	if isRoot {
		tflog.Debug(ctx, "user is root, all allowed")
		return
	}
	if model.CloudInit != nil {
		resp.Diagnostics.AddError("Root only property set", "A current bug prevents non-root users from using cloud-init https://lists.proxmox.com/pipermail/pve-devel/2023-March/056155.html")
	}
	if !model.CPU.Architecture.IsNull() && !model.CPU.Architecture.IsUnknown() {
		resp.Diagnostics.AddError("Root only property set", "The field cpu.architecture is only allowed to be set by root users")
	}
}

func authUpdateValidator(ctx context.Context, isRoot bool, model *vt.VirtualMachineResourceModel, resp *resource.ModifyPlanResponse) {
	tflog.Debug(ctx, "authCreateValidator")
	if isRoot {
		tflog.Debug(ctx, "user is root, all allowed")
		return
	}
	if model.CloudInit != nil {
		resp.Diagnostics.AddError("Root only property set", "A current bug prevents non-root users from using cloud-init https://lists.proxmox.com/pipermail/pve-devel/2023-March/056155.html")
	}
}
