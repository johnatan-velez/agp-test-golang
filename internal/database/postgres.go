package database

import (
	"github.com/acmecorp/platform-api/pkg/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type DB struct {
	conn *gorm.DB
}

func Connect(dsn string) (*DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	if err := db.AutoMigrate(&models.User{}, &models.Tenant{}); err != nil {
		return nil, err
	}

	return &DB{conn: db}, nil
}

func (db *DB) Health() string {
	sqlDB, err := db.conn.DB()
	if err != nil {
		return "unhealthy"
	}
	if err := sqlDB.Ping(); err != nil {
		return "unhealthy"
	}
	return "healthy"
}
