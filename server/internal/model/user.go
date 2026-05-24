package model

type User struct {
	ID                 uint64 `gorm:"primaryKey" json:"id"`
	Username           string `gorm:"uniqueIndex;size:50;not null" json:"username"`
	PasswordHash       string `gorm:"size:255;not null" json:"-"`
	Nickname           string `gorm:"size:50;not null" json:"nickname"`
	Avatar             string `gorm:"size:255;not null" json:"avatar"`
	Gender             int    `gorm:"not null;default:0" json:"gender"`
	Signature          string `gorm:"size:255;not null;default:''" json:"signature"`
	PublicKey          string `gorm:"type:text;not null" json:"publicKey"`
	PublicKeyAlgorithm string `gorm:"size:50;not null;default:''" json:"publicKeyAlgorithm"`
}

func (User) TableName() string {
	return "users"
}
