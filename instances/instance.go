package instances

import (
	"jjmc/manager"
	"jjmc/models"
)

type Instance struct {
	*models.Instance
	Manager *manager.Manager `json:"-"`
}

func (i *Instance) IsRunning() bool {
	return i.Status == "Online" || i.Status == "Starting" || i.Status == "Stopping"
}
