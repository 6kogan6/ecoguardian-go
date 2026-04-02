package model

import "time"

type Reading struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	SensorID    uint      `gorm:"not null;index" json:"sensor_id"`
	Sensor      Sensor    `gorm:"foreignKey:SensorID;constraint:OnDelete:CASCADE;" json:"sensor,omitempty"`
	PM25        float64   `json:"pm25"`
	CO2         float64   `json:"co2"`
	Noise       float64   `json:"noise"`
	Temperature float64   `json:"temperature"`
	Humidity    float64   `json:"humidity"`
	WaterPH     float64   `json:"water_ph"`
	RecordedAt  time.Time `json:"recorded_at"`
}
