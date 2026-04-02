package model

import "time"

type Alert struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	SensorID  uint      `gorm:"not null;index" json:"sensor_id"`
	Sensor    Sensor    `gorm:"foreignKey:SensorID;constraint:OnDelete:CASCADE;" json:"sensor,omitempty"`
	Level     string    `gorm:"size:50;not null" json:"level"`
	Message   string    `gorm:"size:500;not null" json:"message"`
	Status    string    `gorm:"size:50;not null;default:active" json:"status"`
	CreatedAt time.Time `json:"created_at"`
}
