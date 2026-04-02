package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port               string
	DatabasePath       string
	SimulationInterval time.Duration
}

func Load() Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	databasePath := os.Getenv("DATABASE_PATH")
	if databasePath == "" {
		databasePath = "ecoguardian.db"
	}

	simulationInterval := 5 * time.Second
	if raw := os.Getenv("SIMULATION_INTERVAL_SECONDS"); raw != "" {
		if seconds, err := strconv.Atoi(raw); err == nil && seconds > 0 {
			simulationInterval = time.Duration(seconds) * time.Second
		}
	}

	return Config{
		Port:               port,
		DatabasePath:       databasePath,
		SimulationInterval: simulationInterval,
	}
}
