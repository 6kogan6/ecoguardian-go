package config

import "os"

type Config struct {
	Port         string
	DatabasePath string
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

	return Config{
		Port:         port,
		DatabasePath: databasePath,
	}
}
