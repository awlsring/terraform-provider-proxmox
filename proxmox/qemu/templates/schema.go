package templates

import (
	qs "github.com/awlsring/terraform-provider-proxmox/proxmox/qemu/schemas"
	qt "github.com/awlsring/terraform-provider-proxmox/proxmox/qemu/types"

	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var TemplateMultiDataSourceSchema = schema.SetNestedAttribute{
	Computed:     true,
	CustomType:   qt.NewVirtualMachineDataSourceType(),
	NestedObject: TemplateMultiDataSourceObject,
}

var TemplateSingleDataSourceSchema = map[string]schema.Attribute{
	// metadata
	"id": schema.Int64Attribute{
		Computed:    true,
		Optional:    true,
		Description: "The identifier of the template.",
	},
	"node": schema.StringAttribute{
		Required:    true,
		Description: "The node to create the template on.",
	},
	"name": schema.StringAttribute{
		Computed:    true,
		Optional:    true,
		Description: "The name of the template.",
	},
	"description": schema.StringAttribute{
		Computed:    true,
		Description: "The template description.",
	},
	"tags": schema.SetAttribute{
		Computed:    true,
		Description: "The tags of the template.",
		ElementType: types.StringType,
	},
	"agent": schema.SingleNestedAttribute{
		Computed:    true,
		Description: "The agent configuration.",
		Attributes: map[string]schema.Attribute{
			"enabled": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the agent is enabled.",
			},
			"use_fstrim": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether to use fstrim.",
			},
			"type": schema.StringAttribute{
				Computed:    true,
				Description: "The guest agent type.",
			},
		},
	},
	"bios": schema.StringAttribute{
		Computed:    true,
		Description: "The BIOS type.",
	},
	"cpu": schema.SingleNestedAttribute{
		Computed:    true,
		Description: "The CPU configuration.",
		Attributes: map[string]schema.Attribute{
			"architecture": schema.StringAttribute{
				Computed:    true,
				Description: "The CPU architecture.",
			},
			"cores": schema.Int64Attribute{
				Computed:    true,
				Description: "The number of CPU cores.",
			},
			"sockets": schema.Int64Attribute{
				Computed:    true,
				Description: "The number of CPU sockets.",
			},
			"emulated_type": schema.StringAttribute{
				Computed:    true,
				Description: "The emulated CPU type.",
			},
			"cpu_units": schema.Int64Attribute{
				Computed:    true,
				Description: "The CPU units.",
			},
		},
	},
	"disks": schema.SetNestedAttribute{
		Computed:     true,
		Description:  "The terrafrom generated disks attached to the VM.",
		CustomType:   qt.NewVirtualMachineDiskSetType(),
		NestedObject: qs.DiskObjectDataSourceSchema,
	},
	"pci_devices": schema.SetNestedAttribute{
		Computed:     true,
		Description:  "PCI devices passed through to the VM.",
		CustomType:   qt.NewVirtualMachinePCIDeviceSetType(),
		NestedObject: qs.PCIDeviceObjectDataSourceSchema,
	},
	"network_interfaces": schema.SetNestedAttribute{
		Computed:     true,
		CustomType:   qt.NewVirtualMachineNetworkInterfaceSetType(),
		NestedObject: qs.NetworkInterfaceObjectDataSourceSchema,
	},
	"memory": schema.SingleNestedAttribute{
		Computed: true,
		Attributes: map[string]schema.Attribute{
			"dedicated": schema.Int64Attribute{
				Computed:    true,
				Description: "The size of the memory in MB.",
			},
			"floating": schema.Int64Attribute{
				Computed:    true,
				Description: "The floating memory in MB.",
			},
			"shared": schema.Int64Attribute{
				Computed:    true,
				Description: "The shared memory in MB.",
			},
		},
	},
	"machine_type": schema.StringAttribute{
		Computed:    true,
		Description: "The machine type.",
	},
	"kvm_arguments": schema.StringAttribute{
		Computed:    true,
		Description: "The arguments to pass to KVM.",
	},
	"keyboard_layout": schema.StringAttribute{
		Computed:    true,
		Description: "The keyboard layout.",
	},
	"cloud_init": schema.SingleNestedAttribute{
		Computed:   true,
		Attributes: qt.CloudInitDataSourceAttributes,
	},
	"type": schema.StringAttribute{
		Computed:    true,
		Description: "The operating system type.",
	},
	"resource_pool": schema.StringAttribute{
		Computed:    true,
		Description: "The resource pool the template is in.",
	},
	"start_on_node_boot": schema.BoolAttribute{
		Computed:    true,
		Description: "Whether to start the template on node boot.",
	},
}

var TemplateMultiDataSourceObject = schema.NestedAttributeObject{
	Attributes: map[string]schema.Attribute{
		// metadata
		"id": schema.Int64Attribute{
			Computed:    true,
			Description: "The identifier of the template.",
		},
		"node": schema.StringAttribute{
			Computed:    true,
			Description: "The node to create the template on.",
		},
		"name": schema.StringAttribute{
			Computed:    true,
			Description: "The name of the template.",
		},
		"description": schema.StringAttribute{
			Computed:    true,
			Description: "The template description.",
		},
		"tags": schema.SetAttribute{
			Computed:    true,
			Description: "The tags of the template.",
			ElementType: types.StringType,
		},
		"agent": schema.SingleNestedAttribute{
			Computed:    true,
			Description: "The agent configuration.",
			Attributes: map[string]schema.Attribute{
				"enabled": schema.BoolAttribute{
					Computed:    true,
					Description: "Whether the agent is enabled.",
				},
				"use_fstrim": schema.BoolAttribute{
					Computed:    true,
					Description: "Whether to use fstrim.",
				},
				"type": schema.StringAttribute{
					Computed:    true,
					Description: "The guest agent type.",
				},
			},
		},
		"bios": schema.StringAttribute{
			Computed:    true,
			Description: "The BIOS type.",
		},
		"cpu": schema.SingleNestedAttribute{
			Computed:    true,
			Description: "The CPU configuration.",
			Attributes: map[string]schema.Attribute{
				"architecture": schema.StringAttribute{
					Computed:    true,
					Description: "The CPU architecture.",
				},
				"cores": schema.Int64Attribute{
					Computed:    true,
					Description: "The number of CPU cores.",
				},
				"sockets": schema.Int64Attribute{
					Computed:    true,
					Description: "The number of CPU sockets.",
				},
				"emulated_type": schema.StringAttribute{
					Computed:    true,
					Description: "The emulated CPU type.",
				},
				"cpu_units": schema.Int64Attribute{
					Computed:    true,
					Description: "The CPU units.",
				},
			},
		},
		"disks": schema.SetNestedAttribute{
			Computed:     true,
			Description:  "The terrafrom generated disks attached to the VM.",
			CustomType:   qt.NewVirtualMachineDiskSetType(),
			NestedObject: qs.DiskObjectDataSourceSchema,
		},
		"pci_devices": schema.SetNestedAttribute{
			Computed:     true,
			Description:  "PCI devices passed through to the VM.",
			CustomType:   qt.NewVirtualMachinePCIDeviceSetType(),
			NestedObject: qs.PCIDeviceObjectDataSourceSchema,
		},
		"network_interfaces": schema.SetNestedAttribute{
			Computed:     true,
			CustomType:   qt.NewVirtualMachineNetworkInterfaceSetType(),
			NestedObject: qs.NetworkInterfaceObjectDataSourceSchema,
		},
		"memory": schema.SingleNestedAttribute{
			Computed: true,
			Attributes: map[string]schema.Attribute{
				"dedicated": schema.Int64Attribute{
					Computed:    true,
					Description: "The size of the memory in MB.",
				},
				"floating": schema.Int64Attribute{
					Computed:    true,
					Description: "The floating memory in MB.",
				},
				"shared": schema.Int64Attribute{
					Computed:    true,
					Description: "The shared memory in MB.",
				},
			},
		},
		"machine_type": schema.StringAttribute{
			Computed:    true,
			Description: "The machine type.",
		},
		"kvm_arguments": schema.StringAttribute{
			Computed:    true,
			Description: "The arguments to pass to KVM.",
		},
		"keyboard_layout": schema.StringAttribute{
			Computed:    true,
			Description: "The keyboard layout.",
		},
		"cloud_init": schema.SingleNestedAttribute{
			Computed:   true,
			Attributes: qt.CloudInitDataSourceAttributes,
		},
		"type": schema.StringAttribute{
			Computed:    true,
			Description: "The operating system type.",
		},
		"resource_pool": schema.StringAttribute{
			Computed:    true,
			Description: "The resource pool the template is in.",
		},
		"start_on_node_boot": schema.BoolAttribute{
			Computed:    true,
			Description: "Whether to start the template on node boot.",
		},
	},
}
