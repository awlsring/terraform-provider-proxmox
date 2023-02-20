package vm

import "strings"

type VirtualMachineAgent struct {
	Enabled bool
	FsTrim  bool
	Type    *string
}

func DetermineAgentConfig(a *string) *VirtualMachineAgent {
	if a == nil {
		return nil
	}
	agent := VirtualMachineAgent{}
	agentCfg := strings.Split(*a, ",")

	for i, cfg := range agentCfg {
		if i == 0 {
			if cfg == "1" {
				agent.Enabled = true
			} else {
				agent.Enabled = false
			}
		}

		cfg := strings.Split(cfg, "=")
		if len(cfg) != 2 {
			continue
		}

		switch cfg[0] {
		case "fstrim_cloned_disks":
			if cfg[1] == "1" {
				agent.FsTrim = true
			} else {
				agent.FsTrim = false
			}
		case "type":
			agent.Type = &cfg[1]
		}
	}

	return &agent
}
