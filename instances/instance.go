package instances

import (
	"jjmc/manager"
	"jjmc/models"
)

type Instance struct {
	*models.Instance
	Manager *manager.Manager `json:"-"`
	Tunnel  *TunnelManager   `json:"-"`
}

func NewInstance(base *models.Instance, mgr *manager.Manager) *Instance {
	return &Instance{
		Instance: base,
		Manager:  mgr,
		Tunnel:   NewTunnelManager(base.Directory),
	}
}

func (i *Instance) IsRunning() bool {
	return i.Status == "Online" || i.Status == "Starting" || i.Status == "Stopping"
}
