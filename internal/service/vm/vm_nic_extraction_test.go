package vm

import (
	"testing"

	"github.com/awlsring/proxmox-go/proxmox"
	"github.com/stretchr/testify/assert"
)

func Test_parseNicString_Success(t *testing.T) {
	nicString := "virtio=52:54:00:4A:4B:4C,bridge=vmbr0,firewall=1,tag=10"

	nic, err := parseNicString(nicString)
	assert.Nil(t, err)
	assert.Equal(t, VirtualNetworkDeviceModel(VIRTUAL_NIC_VIRTIO), nic.Model)
	assert.Equal(t, "52:54:00:4A:4B:4C", nic.Mac)
	assert.Equal(t, "vmbr0", nic.Bridge)
	assert.True(t, nic.FirewallEnabled)
	assert.Equal(t, 10, nic.Vlan)
}

func Test_parseNicString_BadModel(t *testing.T) {
	nicString := "bad=52:54:00:4A:4B:4C,bridge=vmbr0,firewall=1,tag=10"
	_, err := parseNicString(nicString)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "invalid model: bad")
}

func Test_ExtractNicsFromConfig_Success(t *testing.T) {
	cfg := &proxmox.VirtualMachineConfigurationSummary{
		Net0: proxmox.PtrString("virtio=52:54:00:4A:4B:4C,bridge=vmbr0,firewall=1,tag=10"),
		Net1: proxmox.PtrString("virtio=52:54:00:4A:4B:3C,bridge=vmbr0,firewall=1,tag=10"),
		Net2: proxmox.PtrString("virtio=52:54:00:4A:4B:2C,bridge=vmbr0,firewall=1,tag=10"),
	}

	parsedNics, err := ExtractNicsFromConfig(cfg)
	assert.Nil(t, err)
	assert.Len(t, parsedNics, 3)

	for _, nic := range parsedNics {
		assert.NotEmpty(t, nic.Mac)
		assert.Equal(t, VirtualNetworkDeviceModel(VIRTUAL_NIC_VIRTIO), nic.Model)
		assert.Equal(t, "vmbr0", nic.Bridge)
		assert.True(t, nic.FirewallEnabled)
		assert.Equal(t, 10, nic.Vlan)
	}
}

func Test_ExtractNicsFromConfig_HandlesGaps(t *testing.T) {
	cfg := &proxmox.VirtualMachineConfigurationSummary{
		Net0: proxmox.PtrString("virtio=52:54:00:4A:4B:4C,bridge=vmbr0,firewall=1,tag=10"),
		Net5: proxmox.PtrString("virtio=52:54:00:4A:4B:3C,bridge=vmbr0,firewall=1,tag=10"),
		Net7: proxmox.PtrString("virtio=52:54:00:4A:4B:2C,bridge=vmbr0,firewall=1,tag=10"),
	}

	parsedNics, err := ExtractNicsFromConfig(cfg)
	assert.Nil(t, err)
	assert.Len(t, parsedNics, 3)

	for _, nic := range parsedNics {
		assert.NotEmpty(t, nic.Mac)
		assert.Equal(t, VirtualNetworkDeviceModel(VIRTUAL_NIC_VIRTIO), nic.Model)
		assert.Equal(t, "vmbr0", nic.Bridge)
		assert.True(t, nic.FirewallEnabled)
		assert.Equal(t, 10, nic.Vlan)
	}
}

func Test_ExtractNicsFromConfig_AllTypes(t *testing.T) {
	cfg := &proxmox.VirtualMachineConfigurationSummary{
		Net0: proxmox.PtrString("virtio=52:54:00:4A:4B:4C,bridge=vmbr0,firewall=1,tag=10"),
		Net2: proxmox.PtrString("e1000=52:54:00:4A:4B:3C,bridge=vmbr0,firewall=1,tag=10"),
		Net3: proxmox.PtrString("rtl8139=52:54:00:4A:4B:2C,bridge=vmbr0,firewall=1,tag=10"),
		Net4: proxmox.PtrString("vmxnet3=52:54:00:4A:4B:1C,bridge=vmbr0,firewall=1,tag=10"),
	}

	parsedNics, err := ExtractNicsFromConfig(cfg)
	assert.Nil(t, err)
	assert.Len(t, parsedNics, 4)

	for _, nic := range parsedNics {
		assert.NotEmpty(t, nic.Mac)
		assert.NotEmpty(t, nic.Model)
		assert.True(t, nic.Model.IsValid())
		assert.Equal(t, "vmbr0", nic.Bridge)
		assert.True(t, nic.FirewallEnabled)
		assert.Equal(t, 10, nic.Vlan)
	}
}