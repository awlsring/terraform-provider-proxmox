package vm

import (
	"fmt"

	"github.com/awlsring/proxmox-go/proxmox"
)

func AllocateDiskConfig(t string, position int, config *string, input *proxmox.ApplyVirtualMachineConfigurationSyncRequestContent) error {
	switch t {
	case "scsi":
		switch position {
		case 0:
			input.Scsi0 = config
		case 1:
			input.Scsi1 = config
		case 2:
			input.Scsi2 = config
		case 3:
			input.Scsi3 = config
		case 4:
			input.Scsi4 = config
		case 5:
			input.Scsi5 = config
		case 6:
			input.Scsi6 = config
		case 7:
			input.Scsi7 = config
		case 8:
			input.Scsi8 = config
		case 9:
			input.Scsi9 = config
		case 10:
			input.Scsi10 = config
		case 11:
			input.Scsi11 = config
		case 12:
			input.Scsi12 = config
		case 13:
			input.Scsi13 = config
		case 14:
			input.Scsi14 = config
		case 15:
			input.Scsi15 = config
		case 16:
			input.Scsi16 = config
		case 17:
			input.Scsi17 = config
		case 18:
			input.Scsi18 = config
		case 19:
			input.Scsi19 = config
		case 20:
			input.Scsi20 = config
		case 21:
			input.Scsi21 = config
		case 22:
			input.Scsi22 = config
		case 23:
			input.Scsi23 = config
		case 24:
			input.Scsi24 = config
		case 25:
			input.Scsi25 = config
		case 26:
			input.Scsi26 = config
		case 27:
			input.Scsi27 = config
		case 28:
			input.Scsi28 = config
		case 29:
			input.Scsi29 = config
		case 30:
			input.Scsi30 = config
		default:
			return fmt.Errorf("invalid scsi position")
		}
	case "ide":
		switch position {
		case 0:
			input.Ide0 = config
		case 1:
			input.Ide1 = config
		case 2:
			input.Ide2 = config
		case 3:
			input.Ide3 = config
		default:
			return fmt.Errorf("invalid ide position")
		}
	case "virtio":
		switch position {
		case 0:
			input.Virtio0 = config
		case 1:
			input.Virtio1 = config
		case 2:
			input.Virtio2 = config
		case 3:
			input.Virtio3 = config
		case 4:
			input.Virtio4 = config
		case 5:
			input.Virtio5 = config
		case 6:
			input.Virtio6 = config
		case 7:
			input.Virtio7 = config
		case 8:
			input.Virtio8 = config
		case 9:
			input.Virtio9 = config
		case 10:
			input.Virtio10 = config
		case 11:
			input.Virtio11 = config
		case 12:
			input.Virtio12 = config
		case 13:
			input.Virtio13 = config
		case 14:
			input.Virtio14 = config
		case 15:
			input.Virtio15 = config
		default:
			return fmt.Errorf("invalid virtio position")
		}
	case "sata":
		switch position {
		case 0:
			input.Sata0 = config
		case 1:
			input.Sata1 = config
		case 2:
			input.Sata2 = config
		case 3:
			input.Sata3 = config
		case 4:
			input.Sata4 = config
		case 5:
			input.Sata5 = config
		default:
			return fmt.Errorf("invalid sata position")
		}
	default:
		return fmt.Errorf("invalid disk type")
	}
	return nil
}
