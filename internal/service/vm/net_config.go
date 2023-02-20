package vm

import (
	"fmt"

	"github.com/awlsring/proxmox-go/proxmox"
)

func AllocateNetworkInterfaceConfig(position int, config *string, input *proxmox.ApplyVirtualMachineConfigurationSyncRequestContent) error {
	switch position {
	case 0:
		input.Net0 = config
	case 1:
		input.Net1 = config
	case 2:
		input.Net2 = config
	case 3:
		input.Net3 = config
	case 4:
		input.Net4 = config
	case 5:
		input.Net5 = config
	case 6:
		input.Net6 = config
	case 7:
		input.Net7 = config
	default:
		return fmt.Errorf("invalid position %d", position)
	}
	return nil
}
