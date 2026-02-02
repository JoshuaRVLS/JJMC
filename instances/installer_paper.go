package instances

import (
	"jjmc/installers/paper"
)

func (v *VersionsManager) InstallPaper(version string) error {
	return paper.Install(v.manager.GetWorkDir(), version, v.manager.Broadcast)
}
