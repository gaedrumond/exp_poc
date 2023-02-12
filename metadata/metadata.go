package metadata

import (
	"exp_poc/metadata/macaddr"
	pcname "exp_poc/metadata/pc_name"
)

type MachineMetadata struct {
	MacAdd string
	PCName string
}

func GetMetadata() (meta MachineMetadata) {
	meta.MacAdd = macaddr.GetMacAddr()
	meta.PCName = pcname.GetPCName()
	return
}
