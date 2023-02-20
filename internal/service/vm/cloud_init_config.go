package vm

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/awlsring/proxmox-go/proxmox"
	"github.com/awlsring/terraform-provider-proxmox/internal/service/utils"
)

type VirtualMachineCloudInit struct {
	User *VirtualMachineCloudInitUser
	Ip   []VirtualMachineCloudInitIp
	Dns  *VirtualMachineCloudInitDns
}

type VirtualMachineCloudInitUser struct {
	Name       *string
	Password   *string
	PublicKeys []string
}

type VirtualMachineCloudInitIp struct {
	Interface string
	V4        *VirtualMachineCloudInitIpConfig
	V6        *VirtualMachineCloudInitIpConfig
}

type VirtualMachineCloudInitIpConfig struct {
	DHCP    bool
	Address *string
	Netmask *string
	Gateway *string
}

type VirtualMachineCloudInitDns struct {
	Nameserver *string
	Domain     *string
}

func DetermineCloudInitConfiguration(sum proxmox.VirtualMachineConfigurationSummary) *VirtualMachineCloudInit {
	ci := VirtualMachineCloudInit{}

	setUser := false
	ciUser := VirtualMachineCloudInitUser{}
	if sum.HasCiuser() {
		ciUser.Name = sum.Ciuser
		setUser = true
	}

	if sum.HasCipassword() {
		ciUser.Password = sum.Cipassword
		setUser = true
	}

	if sum.HasSshkeys() {
		pubs := utils.StringLinedToSlice(*sum.Sshkeys)
		ciUser.PublicKeys = pubs
		setUser = true
	}

	setNetConfigs(sum, &ci)

	setDns := false
	ciDns := VirtualMachineCloudInitDns{}
	if sum.HasNameserver() {
		ciDns.Nameserver = sum.Nameserver
	}

	if sum.HasSearchdomain() {
		ciDns.Domain = sum.Searchdomain
	}

	if setUser {
		ci.User = &ciUser
	}

	if setDns {
		ci.Dns = &ciDns
	}

	return &ci
}

func setNetConfigs(cfg proxmox.VirtualMachineConfigurationSummary, config *VirtualMachineCloudInit) {
	var cfgMap map[string]interface{}
	inrec, _ := json.Marshal(cfg)
	json.Unmarshal(inrec, &cfgMap)

	for i := 0; i < 7; i++ {
		d := fmt.Sprintf("%s%v", "ipconfig", i)
		if val, ok := cfgMap[d]; ok {
			ipCfg, err := readIpConfigString(val.(string))
			if err != nil {
				continue
			}
			config.Ip = append(config.Ip, ipCfg)
		}
	}
}

func readIpConfigString(str string) (VirtualMachineCloudInitIp, error) {
	// reference string; "ip=10.0.100.101/24,gw=10.0.100.1"
	ip := VirtualMachineCloudInitIp{}

	splits := strings.Split(str, ",")

	updateV4 := false
	updateV6 := false
	v4Config := VirtualMachineCloudInitIpConfig{}
	v6Config := VirtualMachineCloudInitIpConfig{}

	for _, item := range splits {
		itemSplit := strings.Split(item, "=")
		if len(itemSplit) != 2 {
			return VirtualMachineCloudInitIp{}, fmt.Errorf("invalid ip config string: %s", str)
		}
		key, value := itemSplit[0], itemSplit[1]
		switch key {
		case "ip":
			v4Config.DHCP, v4Config.Address, v4Config.Netmask = unpackIpValues(value)
			updateV4 = true
		case "gw":
			v4Config.Gateway = &value
			updateV4 = true
		case "ip6":
			v6Config.DHCP, v6Config.Address, v6Config.Netmask = unpackIpValues(value)
			updateV6 = true
		case "gw6":
			v6Config.Gateway = &value
			updateV6 = true
		default:
			return VirtualMachineCloudInitIp{}, fmt.Errorf("invalid ip config string: %s", str)
		}
	}

	if updateV4 {
		ip.V4 = &v4Config
	}
	if updateV6 {
		ip.V6 = &v6Config
	}

	return ip, nil
}

func unpackIpValues(value string) (bool, *string, *string) {
	if value == "dhcp" {
		return true, nil, nil
	} else {
		iface := strings.Split(value, "/")
		value, mask := iface[0], iface[1]
		return false, &value, &mask
	}
}
