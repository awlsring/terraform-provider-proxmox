package service

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/awlsring/proxmox-go/proxmox"
	log "github.com/sirupsen/logrus"
)

// doing some janky stuff here to get the disks out of the config
// refator this later to use go routines
func extractDisksFromConfig(cfg *proxmox.VirtualMachineConfigurationSummary) ([]VirtualDisk, error) {
	virtualDisks := []VirtualDisk{}
	
	var cfgMap map[string]interface{}
    inrec, _ := json.Marshal(cfg)
    json.Unmarshal(inrec, &cfgMap)

	virtioDisks, err := loopDiskMap(cfgMap, "virtio", 16)
	if err != nil {
		return nil, err
	}
	virtualDisks = append(virtualDisks, virtioDisks...)

	scsiDisks, err := loopDiskMap(cfgMap, "scsi", 31)
	if err != nil {
		return nil, err
	}
	virtualDisks = append(virtualDisks, scsiDisks...)

	ideDisks, err := loopDiskMap(cfgMap, "ide", 5)
	if err != nil {
		return nil, err
	}
	virtualDisks = append(virtualDisks, ideDisks...)

	sataDisks, err := loopDiskMap(cfgMap, "sata", 6)
	if err != nil {
		return nil, err
	}
	virtualDisks = append(virtualDisks, sataDisks...)

	return virtualDisks, nil
}

func loopDiskMap(m map[string]interface{}, key string, times int) ([]VirtualDisk, error) {
	virtualDisks := []VirtualDisk{}
	for i := 0; i < times; i++ {
		d := fmt.Sprintf("%s%v", key, i)
		if val, ok := m[d]; ok {
			disk, err := parseDiskString(val.(string))
			if err != nil {
				// maybe continue here?
				return nil, err
			}
			disk.Position = d
			virtualDisks = append(virtualDisks, disk)
		} 
	}
	return virtualDisks, nil
}

func parseDiskString(diskString string) (VirtualDisk, error) {
	disk := VirtualDisk{}

	// this is probably fragile, an example of the string is:
	// local-lvm:vm-100-disk-0,size=10G
	// so is
	// storage:the-vm,k1=v1,k2=v2,k3=v3

	splits := strings.Split(diskString, ",")
	storageStr, options := splits[0], splits[1:]

	storage := strings.Split(storageStr, ":")
	if len(storage) != 2 {
		return disk, fmt.Errorf("invalid disk storage string: %s", storageStr)
	}
	disk.Storage = storage[0]

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
		default:
			log.Warnf("unknown disk option: %s", key)
		}
	}

	if disk.Size == 0 {
		return disk, fmt.Errorf("disk size can't be 0")
	}


	return disk, nil
}

func strToBytes(sizeStr string) int64 {
	sizeStr = strings.ToUpper(sizeStr)
	size, _ := strconv.ParseInt(sizeStr[:len(sizeStr)-1], 10, 64)
	unit := sizeStr[len(sizeStr)-1]

	switch unit {
	case 'G':
		return size * 1024 * 1024 * 1024
	case 'M':
		return size * 1024 * 1024
	case 'K':
		return size * 1024
	case 'T':
		return size * 1024 * 1024 * 1024 * 1024
	default:
		return size
	}
}
