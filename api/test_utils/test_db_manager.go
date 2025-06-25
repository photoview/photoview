package test_utils

import (
	"fmt"

	"github.com/photoview/photoview/api/database"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type TestDBManager struct {
	DB *gorm.DB
}

func (dbm *TestDBManager) SetupAndReset() error {
	if dbm.DB == nil {
		if err := dbm.setup(); err != nil {
			return fmt.Errorf("setup db error: %w", err)
		}
	}

	return dbm.reset()
}

func (dbm *TestDBManager) Close() error {
	if dbm.DB == nil {
		return nil
	}

	sqlDB, err := dbm.DB.DB()
	if err != nil {
		return fmt.Errorf("get db instance when closing test database error: %w", err)
	}

	sqlDB.Close()
	dbm.DB = nil

	return nil
}

func (dbm *TestDBManager) setup() error {
	config := gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	}
	db, err := database.ConfigureDatabase(&config)
	if err != nil {
		return fmt.Errorf("configure test database error: %w", err)
	}

	dbm.DB = db

	return nil
}

func (dbm *TestDBManager) reset() error {
	if err := database.ClearDatabase(dbm.DB); err != nil {
		return fmt.Errorf("clean database error: %w", err)
	}

	if err := database.MigrateDatabase(dbm.DB); err != nil {
		return fmt.Errorf("migrate database error: %w", err)
	}

	return nil
}
