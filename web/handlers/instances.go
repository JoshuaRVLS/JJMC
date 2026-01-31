package handlers

import (
	"jjmc/instances"
)

type InstanceHandler struct {
	Manager *instances.InstanceManager
}

func NewInstanceHandler(im *instances.InstanceManager) *InstanceHandler {
	return &InstanceHandler{Manager: im}
}
