package model

import "time"

type Moment struct {
	ID         uint64    `gorm:"primaryKey" json:"id"`
	UserID     uint64    `gorm:"not null;index" json:"userId"`
	Content    string    `gorm:"type:text;not null" json:"content"`
	ImagesJSON string    `gorm:"type:text;not null;column:images_json" json:"-"`
	CreatedAt  time.Time `json:"createdAt"`
}

func (Moment) TableName() string {
	return "moments"
}
