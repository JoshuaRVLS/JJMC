package models

type Schedule struct {
	ID             string `json:"id" gorm:"primaryKey"`
	InstanceID     string `json:"instanceId"`
	Name           string `json:"name"`
	CronExpression string `json:"cronExpression"`
	Type           string `json:"type"` // "command", "restart", "backup"
	Payload        string `json:"payload"`
	Enabled        bool   `json:"enabled"`
	LastRun        int64  `json:"lastRun"`
}
