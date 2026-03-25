package database

import (
	"ecoguardian-go/internal/model"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func Init(databasePath string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(databasePath), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	err = db.AutoMigrate(
		&model.Sensor{},
		&model.Reading{},
		&model.Alert{},
	)
	if err != nil {
		return nil, err
	}

	return db, nil
}
