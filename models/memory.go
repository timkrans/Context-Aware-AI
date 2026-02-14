package models

type Memory struct {
	ID       uint   `gorm:"primaryKey"`
	Text     string
	Embedding []byte `gorm:"type:blob"`
	UserID   uint   `gorm:"index"`
	TabID    uint   `gorm:"index"`
}
