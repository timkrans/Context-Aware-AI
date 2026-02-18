package models

type Document struct {
    ID        uint   `gorm:"primaryKey"`
    UserID    uint
    TabID     uint
    Source    string
    Content   string
    Embedding []byte
}
