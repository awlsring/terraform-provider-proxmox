package vm

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/awlsring/proxmox-go/proxmox"
	log "github.com/sirupsen/logrus"
)

type VirtualMachineNetworkInterface struct {
	Bridge    string
	Enabled   bool
	Firewall  bool
	MAC       string
	Model     string
	RateLimit *int64
	VLAN      *int
	MTU       *int64
	Position  int
}

func DetermineNetworkDevicesFromConfig(cfg *proxmox.VirtualMachineConfigurationSummary) ([]VirtualMachineNetworkInterface, error) {
	var cfgMap map[string]interface{}
	inrec, err := json.Marshal(cfg)
	if err != nil {
		return nil, err
	}
	json.Unmarshal(inrec, &cfgMap)

	nics := []VirtualMachineNetworkInterface{}
	for i := 0; i < 8; i++ {
		n := fmt.Sprintf("%s%v", "net", i)
		if val, ok := cfgMap[n]; ok {
			nic, err := readNicString(val.(string))
			if err != nil {
				return nil, err
			}
			nic.Position = i
			nics = append(nics, nic)
		}
	}

	return nics, nil
}

func readNicString(nicString string) (VirtualMachineNetworkInterface, error) {

	// this is probably fragile, an example of the string is:
	// virtio=52:54:00:4A:4B:4C,bridge=vmbr0,firewall=1
	// so is
	// model=mac,k1=v1,k2=v2,k3=v3

	splits := strings.Split(nicString, ",")
	modelAndMac, options := splits[0], splits[1:]
	model, mac, err := readModelAndMacFromField(modelAndMac)
	if err != nil {
		return VirtualMachineNetworkInterface{}, err
	}

	nic := VirtualMachineNetworkInterface{
		Model: model,
		MAC:   mac,
	}

	for _, option := range options {
		values := strings.Split(option, "=")
		key, value := values[0], values[1]
		switch key {
		case "bridge":
			nic.Bridge = value
		case "tag":
			vlan, err := strconv.Atoi(value)
			if err != nil {
				return VirtualMachineNetworkInterface{}, err
			}
			nic.VLAN = &vlan
		case "firewall":
			firewall, err := strconv.ParseBool(value)
			if err != nil {
				return VirtualMachineNetworkInterface{}, err
			}
			nic.Firewall = firewall
		case "link_down":
			enabled, err := strconv.ParseBool(value)
			if err != nil {
				return VirtualMachineNetworkInterface{}, err
			}
			nic.Enabled = !enabled
		case "rate":
			rate, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return VirtualMachineNetworkInterface{}, err
			}
			nic.RateLimit = &rate
		case "mtu":
			mtu, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return VirtualMachineNetworkInterface{}, err
			}
			nic.MTU = &mtu
		default:
			log.Warnf("unknown nic option: %s", key)
		}
	}

	return nic, nil
}

func readModelAndMacFromField(s string) (string, string, error) {
	ts := strings.Split(s, "=")
	if len(ts) != 2 {
		return "", "", fmt.Errorf("invalid model string: %s", s)
	}
	model := ts[0]
	return model, ts[1], nil
}
