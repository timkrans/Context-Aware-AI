package models

type User struct {
	ID           uint   `gorm:"primaryKey"`
	UserName     string `gorm:"size:255;unique"`
	PasswordHash string `gorm:"size:255"`
}
