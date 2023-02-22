package vm

type VirtualMachinePCIDevice struct {
	Name       string
	ID         string
	PCIE       bool
	MDEV       *string
	ROMBAR     bool
	ROMFile    *string
	PrimaryGPU bool
}
