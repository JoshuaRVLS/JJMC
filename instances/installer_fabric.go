package instances

import (
	"jjmc/installers/fabric"
	"jjmc/installers/quilt"
)

func (v *VersionsManager) InstallFabric(version string) error {
	return fabric.Install(v.manager.GetWorkDir(), version, v.manager.Broadcast)
}

func (v *VersionsManager) InstallQuilt(version string) error {
	return quilt.Install(v.manager.GetWorkDir(), version, v.manager.Broadcast)
}
