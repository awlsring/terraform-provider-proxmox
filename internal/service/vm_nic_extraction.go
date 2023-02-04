package service

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/awlsring/proxmox-go/proxmox"
	log "github.com/sirupsen/logrus"
)


func extractNicsFromConfig(cfg *proxmox.VirtualMachineConfigurationSummary) ([]VirtualNetworkDevice, error) {
	var cfgMap map[string]interface{}
    inrec, _ := json.Marshal(cfg)
    json.Unmarshal(inrec, &cfgMap)

	nics := []VirtualNetworkDevice{}
	for i := 0; i < 8; i++ {
		n := fmt.Sprintf("%s%v", "net", i)
		if val, ok := cfgMap[n]; ok {
			nic, err := parseNicString(val.(string))
			if err != nil {
				return nil, err
			}
			nic.Position = n
			nics = append(nics, nic)
		} 
	}

	return nics, nil
}

func parseNicString(nicString string) (VirtualNetworkDevice, error) {

	// this is probably fragile, an example of the string is:
	// virtio=52:54:00:4A:4B:4C,bridge=vmbr0,firewall=1
	// so is
	// model=mac,k1=v1,k2=v2,k3=v3

	splits := strings.Split(nicString, ",")
	modelAndMac, options := splits[0], splits[1:]
	
	model, mac, err := geModelAndMacFromField(modelAndMac)
	if err != nil {
		return VirtualNetworkDevice{}, err
	}

	nic := VirtualNetworkDevice{
		Model: model,
		Mac: mac,
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
				return VirtualNetworkDevice{}, err
			}
			nic.Vlan = vlan
		case "firewall":
			firewall, err := strconv.ParseBool(value)
			if err != nil {
				return VirtualNetworkDevice{}, err
			}
			nic.FirewallEnabled = firewall
		default:
			log.Warnf("unknown nic option: %s", key)
		}
	}

	return nic, nil
}

func geModelAndMacFromField(s string) (VirtualNetworkDeviceModel, string, error) {
	ts := strings.Split(s, "=")
	if len(ts) != 2 {
		return "", "", fmt.Errorf("invalid model string: %s", s)
	}
	model := VirtualNetworkDeviceModel(ts[0])
	if !model.IsValid() {
		return "", "", fmt.Errorf("invalid model: %s", model)
	}
	return model, ts[1], nil
}