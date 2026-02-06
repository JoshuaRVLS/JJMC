package models

type Folder struct {
	ID        string `json:"id" gorm:"primaryKey"`
	Name      string `json:"name"`
	CreatedAt int64  `json:"createdAt"`
}
