package service

import (
	"fmt"
	"time"

	"ecoguardian-go/internal/model"
)

func BuildAlerts(reading model.Reading) []model.Alert {
	alerts := make([]model.Alert, 0)

	addAlert := func(level string, message string) {
		alerts = append(alerts, model.Alert{
			SensorID:  reading.SensorID,
			Level:     level,
			Message:   message,
			Status:    "active",
			CreatedAt: time.Now(),
		})
	}

	if reading.PM25 > 55 {
		addAlert("danger", fmt.Sprintf("Critical PM2.5 level: %.2f", reading.PM25))
	} else if reading.PM25 > 35 {
		addAlert("warning", fmt.Sprintf("High PM2.5 level: %.2f", reading.PM25))
	}

	if reading.CO2 > 1500 {
		addAlert("danger", fmt.Sprintf("Critical CO2 level: %.2f", reading.CO2))
	} else if reading.CO2 > 1000 {
		addAlert("warning", fmt.Sprintf("High CO2 level: %.2f", reading.CO2))
	}

	if reading.Noise > 85 {
		addAlert("danger", fmt.Sprintf("Critical noise level: %.2f", reading.Noise))
	} else if reading.Noise > 70 {
		addAlert("warning", fmt.Sprintf("High noise level: %.2f", reading.Noise))
	}

	if reading.WaterPH > 0 && (reading.WaterPH < 6.5 || reading.WaterPH > 8.5) {
		addAlert("warning", fmt.Sprintf("Water pH out of normal range: %.2f", reading.WaterPH))
	}

	return alerts
}
