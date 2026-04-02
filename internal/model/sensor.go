package model

import "time"

type Sensor struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	Name       string    `gorm:"size:100;not null" json:"name"`
	Location   string    `gorm:"size:255;not null" json:"location"`
	SensorType string    `gorm:"size:100;not null" json:"sensor_type"`
	CreatedAt  time.Time `json:"created_at"`
}
