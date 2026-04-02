package service

import (
	"math"
	"math/rand"
	"sync"
	"time"

	"ecoguardian-go/internal/model"

	"gorm.io/gorm"
)

type SimulatorStatus struct {
	Running             bool      `json:"running"`
	IntervalSeconds     int       `json:"interval_seconds"`
	LastGeneratedAt     time.Time `json:"last_generated_at"`
	LastGeneratedCount  int       `json:"last_generated_count"`
	LastGeneratedAlerts int       `json:"last_generated_alerts"`
}

type SimulatorService struct {
	db       *gorm.DB
	interval time.Duration

	mu                  sync.Mutex
	running             bool
	stopChan            chan struct{}
	lastGeneratedAt     time.Time
	lastGeneratedCount  int
	lastGeneratedAlerts int
}

func NewSimulatorService(db *gorm.DB, interval time.Duration) *SimulatorService {
	rand.Seed(time.Now().UnixNano())

	return &SimulatorService{
		db:       db,
		interval: interval,
	}
}

func (s *SimulatorService) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return nil
	}

	s.running = true
	s.stopChan = make(chan struct{})

	go s.loop(s.stopChan)

	return nil
}

func (s *SimulatorService) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return nil
	}

	close(s.stopChan)
	s.running = false
	s.stopChan = nil

	return nil
}

func (s *SimulatorService) Status() SimulatorStatus {
	s.mu.Lock()
	defer s.mu.Unlock()

	return SimulatorStatus{
		Running:             s.running,
		IntervalSeconds:     int(s.interval.Seconds()),
		LastGeneratedAt:     s.lastGeneratedAt,
		LastGeneratedCount:  s.lastGeneratedCount,
		LastGeneratedAlerts: s.lastGeneratedAlerts,
	}
}

func (s *SimulatorService) GenerateOnce() (int, int, error) {
	count, alerts, err := s.generateCycle()
	if err != nil {
		return 0, 0, err
	}

	s.mu.Lock()
	s.lastGeneratedAt = time.Now()
	s.lastGeneratedCount = count
	s.lastGeneratedAlerts = alerts
	s.mu.Unlock()

	return count, alerts, nil
}

func (s *SimulatorService) loop(stopChan chan struct{}) {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			count, alerts, err := s.generateCycle()
			if err != nil {
				continue
			}

			s.mu.Lock()
			s.lastGeneratedAt = time.Now()
			s.lastGeneratedCount = count
			s.lastGeneratedAlerts = alerts
			s.mu.Unlock()

		case <-stopChan:
			return
		}
	}
}

func (s *SimulatorService) generateCycle() (int, int, error) {
	var sensors []model.Sensor
	if err := s.db.Find(&sensors).Error; err != nil {
		return 0, 0, err
	}

	if len(sensors) == 0 {
		return 0, 0, nil
	}

	createdReadings := 0
	createdAlerts := 0

	for _, sensor := range sensors {
		reading := s.generateReading(sensor)

		if err := s.db.Create(&reading).Error; err != nil {
			return createdReadings, createdAlerts, err
		}
		createdReadings++

		alerts := BuildAlerts(reading)
		if len(alerts) > 0 {
			if err := s.db.Create(&alerts).Error; err != nil {
				return createdReadings, createdAlerts, err
			}
			createdAlerts += len(alerts)
		}
	}

	return createdReadings, createdAlerts, nil
}

func (s *SimulatorService) generateReading(sensor model.Sensor) model.Reading {
	now := time.Now()

	switch sensor.SensorType {
	case "air":
		pm25 := randomRange(8, 28)
		co2 := randomRange(450, 900)
		noise := randomRange(35, 65)

		if chance(0.25) {
			pm25 = randomRange(38, 82)
		}
		if chance(0.20) {
			co2 = randomRange(1100, 1900)
		}
		if chance(0.15) {
			noise = randomRange(72, 95)
		}

		return model.Reading{
			SensorID:    sensor.ID,
			PM25:        round(pm25),
			CO2:         round(co2),
			Noise:       round(noise),
			Temperature: round(randomRange(16, 30)),
			Humidity:    round(randomRange(35, 75)),
			WaterPH:     0,
			RecordedAt:  now,
		}

	case "water":
		waterPH := randomRange(6.8, 8.2)
		if chance(0.20) {
			if chance(0.50) {
				waterPH = randomRange(5.7, 6.3)
			} else {
				waterPH = randomRange(8.7, 9.4)
			}
		}

		return model.Reading{
			SensorID:    sensor.ID,
			PM25:        0,
			CO2:         0,
			Noise:       round(randomRange(20, 45)),
			Temperature: round(randomRange(8, 24)),
			Humidity:    round(randomRange(45, 90)),
			WaterPH:     round(waterPH),
			RecordedAt:  now,
		}

	case "noise":
		noise := randomRange(40, 68)
		if chance(0.30) {
			noise = randomRange(74, 98)
		}

		return model.Reading{
			SensorID:    sensor.ID,
			PM25:        0,
			CO2:         0,
			Noise:       round(noise),
			Temperature: round(randomRange(10, 32)),
			Humidity:    round(randomRange(30, 80)),
			WaterPH:     0,
			RecordedAt:  now,
		}

	case "weather":
		return model.Reading{
			SensorID:    sensor.ID,
			PM25:        round(randomRange(5, 22)),
			CO2:         round(randomRange(380, 700)),
			Noise:       round(randomRange(20, 50)),
			Temperature: round(randomRange(-5, 32)),
			Humidity:    round(randomRange(30, 95)),
			WaterPH:     0,
			RecordedAt:  now,
		}

	default:
		pm25 := randomRange(10, 30)
		co2 := randomRange(500, 950)
		noise := randomRange(35, 70)

		if chance(0.15) {
			pm25 = randomRange(40, 70)
		}
		if chance(0.15) {
			co2 = randomRange(1100, 1800)
		}
		if chance(0.15) {
			noise = randomRange(72, 92)
		}

		return model.Reading{
			SensorID:    sensor.ID,
			PM25:        round(pm25),
			CO2:         round(co2),
			Noise:       round(noise),
			Temperature: round(randomRange(10, 30)),
			Humidity:    round(randomRange(35, 80)),
			WaterPH:     0,
			RecordedAt:  now,
		}
	}
}

func randomRange(min, max float64) float64 {
	return min + rand.Float64()*(max-min)
}

func chance(probability float64) bool {
	return rand.Float64() < probability
}

func round(value float64) float64 {
	return math.Round(value*10) / 10
}
