package instances

import (
	"jjmc/manager"
	"jjmc/models"
)

type Instance struct {
	*models.Instance
	Manager *manager.Manager `json:"-"`
}
