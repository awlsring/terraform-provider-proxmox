package qemu

import (
	"regexp"

	"github.com/awlsring/terraform-provider-proxmox/proxmox/defaults"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
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
			Description: "The CPU description.",
		},
		"tags": schema.ListAttribute{
			Optional:    true,
			Description: "The tags of the virtual machine.",
			ElementType: types.StringType,
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
			Attributes: map[string]schema.Attribute{
				"architecture": schema.StringAttribute{
					Optional:    true,
					Computed:    true,
					Description: "The CPU architecture.",
					Validators: []validator.String{
						stringvalidator.OneOf(
							"x86_64",
							"aarch64",
						),
					},
					PlanModifiers: []planmodifier.String{
						defaults.DefaultString("x86_64"),
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
					Computed:    true,
					Description: "The CPU units.",
					PlanModifiers: []planmodifier.Int64{
						defaults.DefaultInt64(100),
					},
				},
			},
		},
		"disks": schema.ListNestedAttribute{
			Optional: true,
			Computed: true,
			PlanModifiers: []planmodifier.List{
				defaults.DefaultList([]attr.Value{}),
			},
			NestedObject: schema.NestedAttributeObject{
				Attributes: map[string]schema.Attribute{
					"storage": schema.StringAttribute{
						Required:    true,
						Description: "The storage the disk is on.",
					},
					"file_format": schema.StringAttribute{
						Computed:    true,
						Optional:    true,
						Description: "The file format of the disk.",
						Validators: []validator.String{
							stringvalidator.OneOf(
								"raw",
								"qcow2",
								"vmdk",
							),
						},
						PlanModifiers: []planmodifier.String{
							defaults.DefaultString("qcow2"),
						},
					},
					"size": schema.Int64Attribute{
						Required:    true,
						Description: "The size of the disk in bytes.",
					},
					"use_iothread": schema.BoolAttribute{
						Optional:    true,
						Computed:    true,
						Description: "Whether to use an iothread for the disk.",
						PlanModifiers: []planmodifier.Bool{
							defaults.DefaultBool(false),
						},
					},
					"speed_limits": schema.SingleNestedAttribute{
						Optional:    true,
						Description: "The speed limits of the disk. If not set, no speed limitations are applied.",
						Attributes: map[string]schema.Attribute{
							"read": schema.Int64Attribute{
								Optional:    true,
								Description: "The read speed limit in bytes per second.",
							},
							"write": schema.Int64Attribute{
								Optional:    true,
								Description: "The write speed limit in bytes per second.",
							},
							"write_burstable": schema.Int64Attribute{
								Optional:    true,
								Description: "The write burstable speed limit in bytes per second.",
							},
							"read_burstable": schema.Int64Attribute{
								Optional:    true,
								Description: "The read burstable speed limit in bytes per second.",
							},
						},
					},
					"interface_type": schema.StringAttribute{
						Required:    true,
						Description: "The type of the disk.",
						Validators: []validator.String{
							stringvalidator.OneOf(
								"scsi",
								"sata",
								"virtio",
							),
						},
					},
					"ssd_emulation": schema.BoolAttribute{
						Computed:    true,
						Optional:    true,
						Description: "Whether to use SSD emulation. conflicts with virtio disk type.",
						PlanModifiers: []planmodifier.Bool{
							defaults.DefaultBool(false),
						},
					},
					"position": schema.Int64Attribute{
						Required:    true,
						Description: "The position of the disk. (0, 1, 2, etc.) This is combined with the `interface_type` to determine the disk name.",
					},
					"discard": schema.BoolAttribute{
						Computed:    true,
						Optional:    true,
						Description: "Whether the disk has discard enabled.",
						PlanModifiers: []planmodifier.Bool{
							defaults.DefaultBool(true),
						},
					},
				},
			},
		},
		"pci_devices": schema.ListNestedAttribute{
			Optional:    true,
			Description: "PCI devices passed through to the VM.",
			NestedObject: schema.NestedAttributeObject{
				Attributes: map[string]schema.Attribute{
					"device_name": schema.StringAttribute{
						Required:    true,
						Description: "The device name of the PCI device.",
					},
					"device_id": schema.StringAttribute{
						Required:    true,
						Description: "The device ID of the PCI device.",
					},
					"pcie": schema.BoolAttribute{
						Optional:    true,
						Computed:    true,
						Description: "Whether the PCI device is PCIe.",
						PlanModifiers: []planmodifier.Bool{
							defaults.DefaultBool(false),
						},
					},
					"mdev": schema.StringAttribute{
						Optional:    true,
						Description: "The mediated device name.",
					},
				},
			},
		},
		"network_interfaces": schema.ListNestedAttribute{
			Optional: true,
			Computed: true,
			PlanModifiers: []planmodifier.List{
				defaults.DefaultList([]attr.Value{}),
			},
			NestedObject: schema.NestedAttributeObject{
				Attributes: map[string]schema.Attribute{
					"bridge": schema.StringAttribute{
						Computed:    true,
						Optional:    true,
						Description: "The bridge the network interface is on.",
						Validators: []validator.String{
							stringvalidator.RegexMatches(regexp.MustCompile("vmbr[0-9]$"), "name must follow scheme `vmbr<n>`"),
						},
					},
					"enabled": schema.BoolAttribute{
						Optional:    true,
						Computed:    true,
						Description: "Whether the network interface is enabled.",
						PlanModifiers: []planmodifier.Bool{
							defaults.DefaultBool(true),
						},
					},
					"mac_address": schema.StringAttribute{
						Computed:    true,
						Optional:    true,
						Description: "The MAC address of the network interface.",
						Validators: []validator.String{
							stringvalidator.RegexMatches(regexp.MustCompile("^([0-9A-Fa-f]{2}[:-]){5}([0-9A-Fa-f]{2})$"), "must be a valid MAC address"),
						},
					},
					"model": schema.StringAttribute{
						Computed:    true,
						Optional:    true,
						Description: "The model of the network interface.",
						PlanModifiers: []planmodifier.String{
							defaults.DefaultString("virtio"),
						},
						Validators: []validator.String{
							stringvalidator.OneOf(
								"virtio",
								"e1000",
								"rtl8139",
								"vmxnet3",
							),
						},
					},
					"rate_limit": schema.Int64Attribute{
						Optional:    true,
						Description: "The rate limit of the network interface in megabytes per second.",
					},
					"vlan": schema.NumberAttribute{
						Optional:    true,
						Description: "The VLAN tag of the network interface.",
					},
					"mtu": schema.Int64Attribute{
						Optional:    true,
						Description: "The MTU of the network interface. Only valid for virtio.",
					},
				},
			},
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
				},
				"shared": schema.Int64Attribute{
					Optional:    true,
					Description: "The shared memory in MB.",
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
			Optional: true,
			Attributes: map[string]schema.Attribute{
				"user": schema.SingleNestedAttribute{
					Required: true,
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Required:    true,
							Description: "The name of the user.",
						},
						"password": schema.StringAttribute{
							Optional:    true,
							Description: "The password of the user.",
						},
						"public_keys": schema.ListAttribute{
							Optional:    true,
							Description: "The public ssh keys of the user.",
							ElementType: types.StringType,
						},
					},
				},
				"ip": schema.SingleNestedAttribute{
					Optional: true,
					Attributes: map[string]schema.Attribute{
						"v4": schema.SingleNestedAttribute{
							Optional: true,
							Attributes: map[string]schema.Attribute{
								"dhcp": schema.BoolAttribute{
									Optional:    true,
									Description: "Whether to use DHCP to get the IP address.",
								},
								"address": schema.StringAttribute{
									Optional:    true,
									Description: "The IP address to use for the machine.",
								},
								"gateway": schema.StringAttribute{
									Optional:    true,
									Description: "The gateway to use for the machine.",
								},
							},
						},
						"v6": schema.SingleNestedAttribute{
							Optional: true,
							Attributes: map[string]schema.Attribute{
								"dhcp": schema.BoolAttribute{
									Optional:    true,
									Description: "Whether to use DHCP to get the IP address.",
								},
								"address": schema.StringAttribute{
									Optional:    true,
									Description: "The IP address to use for the machine.",
								},
								"gateway": schema.StringAttribute{
									Optional:    true,
									Description: "The gateway to use for the machine.",
								},
							},
						},
					},
				},
				"dns": schema.SingleNestedAttribute{
					Optional: true,
					Attributes: map[string]schema.Attribute{
						"nameserver": schema.StringAttribute{
							Optional:    true,
							Description: "The nameserver to use for the machine.",
						},
						"domain": schema.StringAttribute{
							Optional:    true,
							Description: "The domain to use for the machine.",
						},
					},
				},
			},
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
				"move_disk": schema.Int64Attribute{
					Optional:    true,
					Description: "The timeout for moving the virtual machine disk.",
				},
			},
		},
	},
}