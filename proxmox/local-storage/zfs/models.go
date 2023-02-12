package zfs

import (
	"github.com/awlsring/terraform-provider-proxmox/internal/service"
	"github.com/awlsring/terraform-provider-proxmox/proxmox/filters"
	"github.com/awlsring/terraform-provider-proxmox/proxmox/utils"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type zfsDataSourceModel struct {
	ZFSPools []zfsModel            `tfsdk:"zfs_pools"`
	Filters  []filters.FilterModel `tfsdk:"filters"`
}

type zfsModel struct {
	ID        types.String   `tfsdk:"id"`
	Node      types.String   `tfsdk:"node"`
	Name      types.String   `tfsdk:"name"`
	Size      types.Int64    `tfsdk:"size"`
	Health    types.String   `tfsdk:"health"`
	RaidLevel types.String   `tfsdk:"raid_level"`
	Disks     []types.String `tfsdk:"disks"`
}

func ZFSToModel(zfs *service.ZFSPool) zfsModel {
	m := zfsModel{
		ID:        types.StringValue(utils.FormId(zfs.Node, zfs.Name)),
		Node:      types.StringValue(zfs.Node),
		Name:      types.StringValue(zfs.Name),
		Size:      types.Int64Value(zfs.Size),
		Health:    types.StringValue(zfs.Health),
		RaidLevel: types.StringValue(string(zfs.RaidLevel)),
	}

	for _, disk := range zfs.Disks {
		m.Disks = append(m.Disks, types.StringValue(disk))
	}

	return m
}
