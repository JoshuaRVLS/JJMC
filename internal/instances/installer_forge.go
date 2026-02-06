package instances

import (
	"jjmc/internal/installers/forge"
)

func (v *VersionsManager) InstallForge(version string) error {
	return forge.Install(v.manager.GetWorkDir(), version, v.manager.Broadcast)
}

func (v *VersionsManager) InstallNeoForge(version string) error {
	return forge.InstallNeo(v.manager.GetWorkDir(), version, v.manager.Broadcast)
}
