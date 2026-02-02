package models

import (
	"time"
)

type Memory struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Type      string    `json:"type"`
	Label     string    `json:"label"`
	Value     string    `json:"value"`
	Timestamp time.Time `json:"timestamp" gorm:"default:current_timestamp"`
}
