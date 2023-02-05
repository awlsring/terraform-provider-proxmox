package vm

import (
	"testing"

	"github.com/awlsring/proxmox-go/proxmox"
	"github.com/stretchr/testify/assert"
)

func Test_ExtractDisksFromConfig_Success(t *testing.T) {
	cfg := &proxmox.VirtualMachineConfigurationSummary{
		Scsi0: proxmox.PtrString("local-lvm:vm-100-disk-0,size=10G"),
	}

	parsedDisks, err := ExtractDisksFromConfig(cfg)
	assert.Nil(t, err)
	assert.Len(t, parsedDisks, 1)

	for _, disk := range parsedDisks {
		assert.Equal(t, "local-lvm", disk.Storage)
		assert.Equal(t, int64(10737418240), disk.Size)
	}
}

func Test_ExtractDisksFromConfig_HandlesGaps(t *testing.T) {
	cfg := &proxmox.VirtualMachineConfigurationSummary{
		Scsi0: proxmox.PtrString("local-lvm:vm-100-disk-0,size=10G"),
		Scsi5: proxmox.PtrString("local-lvm:vm-100-disk-1,size=10G"),
		Scsi7: proxmox.PtrString("local-lvm:vm-100-disk-2,size=10G"),
	}

	parsedDisks, err := ExtractDisksFromConfig(cfg)
	assert.Nil(t, err)
	assert.Len(t, parsedDisks, 3)

	for _, disk := range parsedDisks {	
		assert.Equal(t, "local-lvm", disk.Storage)
		assert.Equal(t, int64(10737418240), disk.Size)
	}
}

func Test_ExtractDisksFromConfig_MissingStorage(t *testing.T) {
	cfg := &proxmox.VirtualMachineConfigurationSummary{
		Scsi0: proxmox.PtrString("vm-100-disk-0,size=10G"),
	}
	_, err := ExtractDisksFromConfig(cfg)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "invalid disk storage string")
}
