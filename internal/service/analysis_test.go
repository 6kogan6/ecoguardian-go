package service_test

import (
	"strings"
	"testing"

	"ecoguardian-go/internal/model"
	"ecoguardian-go/internal/service"
)

func TestBuildAlerts_NoAlertsForNormalReading(t *testing.T) {
	reading := model.Reading{
		SensorID: 1,
		PM25:     20,
		CO2:      700,
		Noise:    50,
		WaterPH:  7.2,
	}

	alerts := service.BuildAlerts(reading)

	if len(alerts) != 0 {
		t.Fatalf("expected 0 alerts, got %d", len(alerts))
	}
}

func TestBuildAlerts_CreatesExpectedAlerts(t *testing.T) {
	tests := []struct {
		name         string
		reading      model.Reading
		wantTotal    int
		wantWarnings int
		wantDangers  int
		wantMessages []string
	}{
		{
			name: "pm25 warning",
			reading: model.Reading{
				SensorID: 1,
				PM25:     40,
			},
			wantTotal:    1,
			wantWarnings: 1,
			wantDangers:  0,
			wantMessages: []string{"PM2.5"},
		},
		{
			name: "critical co2 and noise",
			reading: model.Reading{
				SensorID: 2,
				CO2:      1700,
				Noise:    90,
			},
			wantTotal:    2,
			wantWarnings: 0,
			wantDangers:  2,
			wantMessages: []string{"CO2", "noise"},
		},
		{
			name: "mixed alerts including water ph",
			reading: model.Reading{
				SensorID: 3,
				PM25:     60,
				CO2:      1200,
				Noise:    75,
				WaterPH:  9.1,
			},
			wantTotal:    4,
			wantWarnings: 3,
			wantDangers:  1,
			wantMessages: []string{"PM2.5", "CO2", "noise", "Water pH"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			alerts := service.BuildAlerts(tt.reading)

			if len(alerts) != tt.wantTotal {
				t.Fatalf("expected %d alerts, got %d", tt.wantTotal, len(alerts))
			}

			warnings := 0
			dangers := 0
			joinedMessages := ""

			for _, alert := range alerts {
				if alert.SensorID != tt.reading.SensorID {
					t.Fatalf("expected sensor_id %d, got %d", tt.reading.SensorID, alert.SensorID)
				}

				if alert.Status != "active" {
					t.Fatalf("expected alert status active, got %s", alert.Status)
				}

				switch alert.Level {
				case "warning":
					warnings++
				case "danger":
					dangers++
				default:
					t.Fatalf("unexpected alert level: %s", alert.Level)
				}

				joinedMessages += alert.Message + "\n"
			}

			if warnings != tt.wantWarnings {
				t.Fatalf("expected %d warnings, got %d", tt.wantWarnings, warnings)
			}

			if dangers != tt.wantDangers {
				t.Fatalf("expected %d dangers, got %d", tt.wantDangers, dangers)
			}

			for _, fragment := range tt.wantMessages {
				if !strings.Contains(joinedMessages, fragment) {
					t.Fatalf("expected messages to contain %q, got:\n%s", fragment, joinedMessages)
				}
			}
		})
	}
}
