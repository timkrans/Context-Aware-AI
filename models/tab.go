package models

type Tab struct {
	ID     uint   `gorm:"primaryKey"`
	UserID uint   `gorm:"index"`
	Name   string `gorm:"size:255"`
}
