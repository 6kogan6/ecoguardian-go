package main

import (
	"log"

	"ecoguardian-go/internal/config"
	"ecoguardian-go/internal/database"
	appRouter "ecoguardian-go/internal/router"
	"ecoguardian-go/internal/service"
)

func main() {
	cfg := config.Load()

	db, err := database.Init(cfg.DatabasePath)
	if err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}

	simulator := service.NewSimulatorService(db, cfg.SimulationInterval)
	defer func() {
		_ = simulator.Stop()
	}()

	router := appRouter.SetupRouter(db, simulator)

	log.Printf("EcoGuardian server started on http://localhost:%s", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
