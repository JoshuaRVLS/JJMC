package models

type Template struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Environment map[string]interface{} `json:"environment"`
	Install     []InstallStep          `json:"install"`
	Run         RunConfig              `json:"run"`
}

type InstallStep struct {
	Type    string            `json:"type"`
	Options map[string]string `json:"options"`
}

type RunConfig struct {
	Command   string            `json:"command"`
	Stop      string            `json:"stop"`
	Arguments []string          `json:"arguments"`
	Env       map[string]string `json:"env"`
}
