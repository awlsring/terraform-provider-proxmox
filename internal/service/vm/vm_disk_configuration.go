package vm

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/awlsring/proxmox-go/proxmox"
	log "github.com/sirupsen/logrus"
)

type VirtualMachineDisk struct {
	Storage       string
	FileFormat    *string
	Size          int64
	UseIOThreads  bool
	Position      int
	InterfaceType string
	SpeedLimits   *VirtualMachineDiskSpeedLimits
	SSDEmulation  bool
	Discard       bool
}

type VirtualMachineDiskSpeedLimits struct {
	Read           *int64
	ReadBurstable  *int64
	Write          *int64
	WriteBurstable *int64
}

// take 2 of the vm_disk_extraction from previous data source, I'll need to consolidate these at some point
func DetermineDiskConfiguration(cfg *proxmox.VirtualMachineConfigurationSummary) ([]VirtualMachineDisk, error) {
	virtualDisks := []VirtualMachineDisk{}

	var cfgMap map[string]interface{}
	inrec, _ := json.Marshal(cfg)
	json.Unmarshal(inrec, &cfgMap)

	virtioDisks, err := readDiskMap(cfgMap, "virtio", 16)
	if err != nil {
		return nil, err
	}
	virtualDisks = append(virtualDisks, virtioDisks...)

	scsiDisks, err := readDiskMap(cfgMap, "scsi", 31)
	if err != nil {
		return nil, err
	}
	virtualDisks = append(virtualDisks, scsiDisks...)

	ideDisks, err := readDiskMap(cfgMap, "ide", 5)
	if err != nil {
		return nil, err
	}
	virtualDisks = append(virtualDisks, ideDisks...)

	sataDisks, err := readDiskMap(cfgMap, "sata", 6)
	if err != nil {
		return nil, err
	}
	virtualDisks = append(virtualDisks, sataDisks...)

	unusedDisks, err := readDiskMap(cfgMap, "unused", 7)
	if err != nil {
		return nil, err
	}
	virtualDisks = append(virtualDisks, unusedDisks...)

	return virtualDisks, nil
}

func readDiskMap(m map[string]interface{}, key string, times int) ([]VirtualMachineDisk, error) {
	virtualDisks := []VirtualMachineDisk{}
	for i := 0; i < times; i++ {
		d := fmt.Sprintf("%s%v", key, i)
		if val, ok := m[d]; ok {
			disk, err := readDiskString(val.(string))
			if err != nil {
				// maybe continue here?
				return nil, err
			}
			disk.Position = i
			disk.InterfaceType = key
			virtualDisks = append(virtualDisks, disk)
		}
	}
	return virtualDisks, nil
}

func readDiskString(diskString string) (VirtualMachineDisk, error) {
	disk := VirtualMachineDisk{}

	// this is probably fragile, an example of the string is:
	// local-lvm:vm-100-disk-0,size=10G
	// so is
	// storage:the-vm,k1=v1,k2=v2,k3=v3

	splits := strings.Split(diskString, ",")
	storageStr, options := splits[0], splits[1:]

	storage := strings.Split(storageStr, ":")
	if len(storage) != 2 {
		if len(storage) == 1 {
			if storage[0] == "none" {
				disk.Storage = storage[0]
				return disk, nil
			}
		}
		return disk, fmt.Errorf("invalid disk storage string: %s", storageStr)
	}
	disk.Storage = storage[0]

	diskSpeedLimits := VirtualMachineDiskSpeedLimits{}
	for _, option := range options {
		values := strings.Split(option, "=")
		key, value := values[0], values[1]
		switch key {
		case "size":
			disk.Size = strToBytes(value)
		case "discard":
			if value == "on" {
				disk.Discard = true
			} else {
				disk.Discard = false
			}
		case "ssd":
			if value == "on" {
				disk.SSDEmulation = true
			} else {
				disk.SSDEmulation = false
			}
		case "format":
			disk.FileFormat = &value
		case "iothread":
			if value == "1" {
				disk.UseIOThreads = true
			} else {
				disk.UseIOThreads = false
			}
		case "mpbs_rd":
			n, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return disk, err
			}
			diskSpeedLimits.Read = &n
		case "mpbs_rd_max":
			n, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return disk, err
			}
			diskSpeedLimits.ReadBurstable = &n
		case "mpbs_wr":
			n, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return disk, err
			}
			diskSpeedLimits.Write = &n
		case "mpbs_wr_max":
			n, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return disk, err
			}
			diskSpeedLimits.WriteBurstable = &n
		default:
			log.Warnf("unknown disk option: %s", key)
		}
	}
	disk.SpeedLimits = &diskSpeedLimits
	return disk, nil
}
