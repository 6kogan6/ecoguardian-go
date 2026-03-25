package main

import (
	"log"

	"ecoguardian-go/internal/config"
	"ecoguardian-go/internal/database"
	appRouter "ecoguardian-go/internal/router"
)

func main() {
	cfg := config.Load()

	db, err := database.Init(cfg.DatabasePath)
	if err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}

	router := appRouter.SetupRouter(db)

	log.Printf("EcoGuardian server started on http://localhost:%s", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
