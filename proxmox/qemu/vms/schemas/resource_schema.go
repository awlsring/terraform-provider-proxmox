package schemas

import (
	"github.com/awlsring/terraform-provider-proxmox/proxmox/defaults"
	qs "github.com/awlsring/terraform-provider-proxmox/proxmox/qemu/schemas"
	t "github.com/awlsring/terraform-provider-proxmox/proxmox/qemu/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var ResourceSchema = schema.Schema{
	Attributes: map[string]schema.Attribute{
		// metadata
		"id": schema.Int64Attribute{
			Optional:    true,
			Computed:    true,
			Description: "The identifier of the virtual machine.",
			PlanModifiers: []planmodifier.Int64{
				int64planmodifier.RequiresReplace(),
			},
		},
		"node": schema.StringAttribute{
			Required:    true,
			Description: "The node to create the virtual machine on.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"name": schema.StringAttribute{
			Optional:    true,
			Description: "The name of the virtual machine.",
		},
		"description": schema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Description: "The virtual machine description.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"tags": schema.SetAttribute{
			Optional:    true,
			Computed:    true,
			Description: "The tags of the virtual machine.",
			ElementType: types.StringType,
			Validators: []validator.Set{
				setvalidator.SizeAtLeast(1),
			},
		},
		// creation configuration
		"clone": schema.SingleNestedAttribute{
			Optional: true,
			Validators: []validator.Object{
				objectvalidator.ConflictsWith(path.Expressions{
					path.MatchRoot("iso"),
				}...),
			},
			PlanModifiers: []planmodifier.Object{
				objectplanmodifier.RequiresReplace(),
			},
			Attributes: map[string]schema.Attribute{
				"storage": schema.StringAttribute{
					Optional:    true,
					Computed:    true,
					Description: "The storage to place the clone on.",
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
						defaults.DefaultString("local"),
					},
				},
				"source": schema.Int64Attribute{
					Required:    true,
					Description: "The identifier of the virtual machine or template to clone.",
					PlanModifiers: []planmodifier.Int64{
						int64planmodifier.RequiresReplace(),
					},
					Validators: []validator.Int64{
						int64validator.AtLeast(100),
						int64validator.AtMost(1000000000),
					},
				},
				"full_clone": schema.BoolAttribute{
					Optional:    true,
					Computed:    true,
					Description: "Whether to clone as a full or linked clone.",
					PlanModifiers: []planmodifier.Bool{
						boolplanmodifier.RequiresReplace(),
						defaults.DefaultBool(true),
					},
				},
			},
		},
		"iso": schema.SingleNestedAttribute{
			Optional:    true,
			Description: "The operating system configuration.",
			Attributes: map[string]schema.Attribute{
				"storage": schema.StringAttribute{
					Required:    true,
					Description: "The storage to place install media on.",
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
				"image": schema.StringAttribute{
					Required:    true,
					Description: "The image to use for install media.",
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
			},
			Validators: []validator.Object{
				objectvalidator.ConflictsWith(path.Expressions{
					path.MatchRoot("clone"),
				}...),
			},
			PlanModifiers: []planmodifier.Object{
				objectplanmodifier.RequiresReplace(),
			},
		}, // method for installing from media
		// "cloud_image": schema.SingleNestedAttribute{}, // will require some janky stuff to get working
		// "pxe": schema.SingleNestedAttribute{},
		// configuration
		"agent": schema.SingleNestedAttribute{
			Optional:    true,
			Computed:    true,
			Description: "The agent configuration.",
			PlanModifiers: []planmodifier.Object{
				defaults.DefaultObject(map[string]attr.Value{
					"enabled":    types.BoolValue(true),
					"use_fstrim": types.BoolValue(false),
					"type":       types.StringValue("virtio"),
				}),
			},
			Attributes: map[string]schema.Attribute{
				"enabled": schema.BoolAttribute{
					Optional:    true,
					Computed:    true,
					Description: "Whether the agent is enabled.",
					PlanModifiers: []planmodifier.Bool{
						defaults.DefaultBool(true),
					},
				},
				"use_fstrim": schema.BoolAttribute{
					Optional:    true,
					Computed:    true,
					Description: "Whether to use fstrim.",
					PlanModifiers: []planmodifier.Bool{
						defaults.DefaultBool(true),
					},
				},
				"type": schema.StringAttribute{
					Optional:    true,
					Computed:    true,
					Description: "The guest agent type.",
					Validators: []validator.String{
						stringvalidator.OneOf(
							"virtio",
							"isa",
						),
					},
					PlanModifiers: []planmodifier.String{
						defaults.DefaultString("virtio"),
					},
				},
			},
		},
		"bios": schema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Description: "The BIOS type.",
			Validators: []validator.String{
				stringvalidator.OneOf(
					"seabios",
					"ovmf",
				),
			},
			PlanModifiers: []planmodifier.String{
				defaults.DefaultString("seabios"),
			},
		},
		"cpu": schema.SingleNestedAttribute{
			Optional:    true,
			Computed:    true,
			Description: "The CPU configuration.",
			PlanModifiers: []planmodifier.Object{
				defaults.DefaultObject(map[string]attr.Value{
					"architecture":  types.StringValue("x86_64"),
					"cores":         types.Int64Value(1),
					"sockets":       types.Int64Value(1),
					"emulated_type": types.StringValue("kvm64"),
					"cpu_units":     types.Int64Value(100),
				}),
			},
			// requires root, make computed til this can be warned about
			Attributes: map[string]schema.Attribute{
				"architecture": schema.StringAttribute{
					Computed:    true,
					Optional:    true,
					Description: "The CPU architecture.",
					Validators: []validator.String{
						stringvalidator.OneOf(
							"x86_64",
							"aarch64",
						),
					},
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
						stringplanmodifier.UseStateForUnknown(),
					},
				},
				"cores": schema.Int64Attribute{
					Optional:    true,
					Computed:    true,
					Description: "The number of CPU cores.",
					PlanModifiers: []planmodifier.Int64{
						defaults.DefaultInt64(1),
					},
				},
				"sockets": schema.Int64Attribute{
					Optional:    true,
					Computed:    true,
					Description: "The number of CPU sockets.",
					PlanModifiers: []planmodifier.Int64{
						defaults.DefaultInt64(1),
					},
				},
				"emulated_type": schema.StringAttribute{
					Optional:    true,
					Computed:    true,
					Description: "The emulated CPU type.",
					PlanModifiers: []planmodifier.String{
						defaults.DefaultString("kvm64"),
					},
					Validators: []validator.String{
						stringvalidator.OneOf(
							"486",
							"Broadwell",
							"Broadwell-IBRS",
							"Broadwell-noTSX",
							"Broadwell-noTSX-IBRS",
							"Cascadelake-Server",
							"Conroe",
							"EPYC",
							"EPYC-IBPB",
							"EPYC-Rome",
							"EPYC-Milan",
							"Haswell",
							"Haswell-IBRS",
							"Haswell-noTSX",
							"Haswell-noTSX-IBRS",
							"host",
							"IvyBridge",
							"IvyBridge-IBRS",
							"KnightsMill",
							"Nehalem",
							"Nehalem-IBRS",
							"Opteron_G1",
							"Opteron_G2",
							"Opteron_G3",
							"Opteron_G4",
							"Opteron_G5",
							"Penryn",
							"Skylake-Client",
							"Skylake-Client-IBRS",
							"Skylake-Server",
							"Skylake-Server-IBRS",
							"SandyBridge",
							"SandyBridge-IBRS",
							"Westmere",
							"Westmere-IBRS",
							"athlon",
							"core2duo",
							"coreduo",
							"kvm32",
							"kvm64",
							"max",
							"pentium",
							"pentium2",
							"pentium3",
							"phenom",
							"qemu32",
							"qemu64",
						),
					},
				},
				"cpu_units": schema.Int64Attribute{
					Optional:    true,
					Description: "The CPU units.",
				},
			},
		},
		"disks": schema.SetNestedAttribute{
			Optional:     true,
			Description:  "The terrafrom generated disks attached to the VM.",
			CustomType:   t.NewVirtualMachineDiskSetType(),
			NestedObject: qs.DiskObjectSchema,
			Validators: []validator.Set{
				setvalidator.SizeAtLeast(1),
			},
		},
		"computed_disks": schema.SetNestedAttribute{
			Computed:     true,
			Description:  "The non terrafrom generated disks attached to the VM.",
			CustomType:   t.NewVirtualMachineDiskSetType(),
			NestedObject: qs.DiskObjectSchema,
		},
		"pci_devices": schema.SetNestedAttribute{
			Optional:     true,
			Description:  "PCI devices passed through to the VM.",
			CustomType:   t.NewVirtualMachinePCIDeviceSetType(),
			NestedObject: qs.PCIDeviceObjectSchema,
			Validators: []validator.Set{
				setvalidator.SizeAtLeast(1),
			},
		},
		"computed_pci_devices": schema.SetNestedAttribute{
			Computed:     true,
			Description:  "The non terraform generated PCI devices passed through to the VM.",
			CustomType:   t.NewVirtualMachinePCIDeviceSetType(),
			NestedObject: qs.PCIDeviceObjectSchema,
		},
		"network_interfaces": schema.SetNestedAttribute{
			Optional:     true,
			CustomType:   t.NewVirtualMachineNetworkInterfaceSetType(),
			NestedObject: qs.NetworkInterfaceObjectSchema,
			PlanModifiers: []planmodifier.Set{
				setplanmodifier.UseStateForUnknown(),
			},
			Validators: []validator.Set{
				setvalidator.SizeAtLeast(1),
			},
		},
		"computed_network_interfaces": schema.SetNestedAttribute{
			Computed:     true,
			CustomType:   t.NewVirtualMachineNetworkInterfaceSetType(),
			NestedObject: qs.NetworkInterfaceObjectSchema,
		},
		"memory": schema.SingleNestedAttribute{
			Optional: true,
			Computed: true,
			PlanModifiers: []planmodifier.Object{
				defaults.DefaultObject(map[string]attr.Value{
					"dedicated": types.Int64Value(1024),
					"floating":  types.Int64Null(),
					"shared":    types.Int64Null(),
				}),
			},
			Attributes: map[string]schema.Attribute{
				"dedicated": schema.Int64Attribute{
					Computed:    true,
					Optional:    true,
					Description: "The size of the memory in MB.",
					PlanModifiers: []planmodifier.Int64{
						defaults.DefaultInt64(1024),
					},
				},
				"floating": schema.Int64Attribute{
					Optional:    true,
					Description: "The floating memory in MB.",
					PlanModifiers: []planmodifier.Int64{
						defaults.DefaultInt64Null(),
					},
				},
				"shared": schema.Int64Attribute{
					Optional:    true,
					Description: "The shared memory in MB.",
					PlanModifiers: []planmodifier.Int64{
						defaults.DefaultInt64Null(),
					},
				},
			},
		},
		"machine_type": schema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Description: "The machine type.",
			Validators: []validator.String{
				stringvalidator.OneOf(
					"i440fx",
					"q35",
				),
			},
			PlanModifiers: []planmodifier.String{
				defaults.DefaultString("q35"),
			},
		},
		"kvm_arguments": schema.StringAttribute{
			Optional:    true,
			Description: "The arguments to pass to KVM.",
		},
		"keyboard_layout": schema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Description: "The keyboard layout.",
			PlanModifiers: []planmodifier.String{
				defaults.DefaultString("en-us"),
			},
			Validators: []validator.String{
				stringvalidator.OneOf(
					"da",
					"de",
					"de-ch",
					"en-gb",
					"en-us",
					"es",
					"fi",
					"fr",
					"fr-be",
					"fr-ca",
					"fr-ch",
					"hu",
					"is",
					"it",
					"ja",
					"lt",
					"nl",
					"no",
					"pl",
					"pt",
					"pt-br",
					"sl",
					"sv",
					"tr",
				),
			},
		},
		"cloud_init": schema.SingleNestedAttribute{
			Optional:   true,
			Attributes: t.CloudInitAttributes,
		},
		"type": schema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Description: "The operating system type.",
			PlanModifiers: []planmodifier.String{
				defaults.DefaultString("other"),
			},
			Validators: []validator.String{
				stringvalidator.OneOf(
					"l24",
					"l26",
					"wxp",
					"w2k",
					"w2k3",
					"w2k8",
					"wvista",
					"win7",
					"win8",
					"win10",
					"win11",
					"solaris",
					"other",
				),
			},
		},
		"resource_pool": schema.StringAttribute{
			Optional:    true,
			Description: "The resource pool the virtual machine is in.",
		},
		"start_on_create": schema.BoolAttribute{
			Optional:    true,
			Computed:    true,
			Description: "Whether to start the virtual machine on creation.",
			PlanModifiers: []planmodifier.Bool{
				defaults.DefaultBool(true),
			},
		},
		"start_on_node_boot": schema.BoolAttribute{
			Optional:    true,
			Computed:    true,
			Description: "Whether to start the virtual machine on node boot.",
			PlanModifiers: []planmodifier.Bool{
				defaults.DefaultBool(true),
			},
		},
		"timeouts": schema.SingleNestedAttribute{
			Optional: true,
			Attributes: map[string]schema.Attribute{
				"create": schema.Int64Attribute{
					Optional:    true,
					Description: "The timeout for creating the virtual machine.",
				},
				"delete": schema.Int64Attribute{
					Optional:    true,
					Description: "The timeout for deleting the virtual machine.",
				},
				"stop": schema.Int64Attribute{
					Optional:    true,
					Description: "The timeout for stopping the virtual machine.",
				},
				"start": schema.Int64Attribute{
					Optional:    true,
					Description: "The timeout for starting the virtual machine.",
				},
				"reboot": schema.Int64Attribute{
					Optional:    true,
					Description: "The timeout for rebooting the virtual machine.",
				},
				"shutdown": schema.Int64Attribute{
					Optional:    true,
					Description: "The timeout for shutting down the virtual machine.",
				},
				"clone": schema.Int64Attribute{
					Optional:    true,
					Description: "The timeout for cloning the virtual machine.",
				},
				"configure": schema.Int64Attribute{
					Optional:    true,
					Description: "The timeout for configuring the virtual machine.",
				},
				"resize_disk": schema.Int64Attribute{
					Optional:    true,
					Description: "The timeout for resizing disk the virtual machine.",
				},
			},
		},
	},
}
