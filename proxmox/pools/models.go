package pools

import (
	"github.com/awlsring/terraform-provider-proxmox/proxmox/filters"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type poolsDataSourceModel struct {
	Pools   []poolModel           `tfsdk:"pools"`
	Filters []filters.FilterModel `tfsdk:"filters"`
}

type poolModel struct {
	ID      types.String      `tfsdk:"id"`
	Comment types.String      `tfsdk:"comment"`
	Members []poolMemberModel `tfsdk:"members"`
}

type poolMemberModel struct {
	ID   types.String `tfsdk:"id"`
	Type types.String `tfsdk:"type"`
}
