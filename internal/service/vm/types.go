package vm

type VirtualMachine struct {
	Id                    int
	Agent                 bool
	Node                  string
	Name                  string
	Memory                int64
	Cores                 int
	Tags                  []string
	VirtualDisks          []VirtualDisk
	VirtualNetworkDevices []VirtualNetworkDevice
}

type VirtualBios string

const (
	VIRTUAL_BIOS_SEABIOS VirtualBios = "seabios"
	VIRTUAL_BIOS_OVMF    VirtualBios = "ovmf"
)

func (e *VirtualBios) IsValid() bool {
	switch *e {
	case VIRTUAL_BIOS_SEABIOS, VIRTUAL_BIOS_OVMF:
		return true
	}
	return false
}

type VirtualNetworkDevice struct {
	Bridge          string
	Vlan            int
	Model           VirtualNetworkDeviceModel
	Mac             string
	Position        string
	FirewallEnabled bool
}

type VirtualNetworkDeviceModel string

const (
	VIRTUAL_NIC_INTEL_E1000     VirtualNetworkDeviceModel = "e1000"
	VIRTUAL_NIC_VIRTIO          VirtualNetworkDeviceModel = "virtio"
	VIRTUAL_NIC_REALTEK_RTL8139 VirtualNetworkDeviceModel = "rtl8139"
	VIRTUAL_NIC_VMWARE_VMXNET3  VirtualNetworkDeviceModel = "vmxnet3"
)

func (e *VirtualNetworkDeviceModel) IsValid() bool {
	switch *e {
	case VIRTUAL_NIC_INTEL_E1000, VIRTUAL_NIC_VIRTIO, VIRTUAL_NIC_REALTEK_RTL8139, VIRTUAL_NIC_VMWARE_VMXNET3:
		return true
	}
	return false
}

type VirtualDisk struct {
	Storage  string
	Type     VirtualDiskType
	Position string //virtio0, virtio1, scsi0, scsi1, ide0, ide1, sata0, sata1, etc
	Size     int64
	Discard  bool
}

type VirtualDiskType string

const (
	VIRTUAL_DISK_SCSI     VirtualDiskType = "scsi"
	VIRTUAL_DISK_VIRTIO_D VirtualDiskType = "virtio"
	VIRTUAL_DISK_SATA     VirtualDiskType = "sata"
	VIRTUAL_DISK_IDE      VirtualDiskType = "ide"
)

func (e *VirtualDiskType) IsValid() bool {
	switch *e {
	case VIRTUAL_DISK_SCSI, VIRTUAL_DISK_VIRTIO_D, VIRTUAL_DISK_SATA, VIRTUAL_DISK_IDE:
		return true
	}
	return false
}
