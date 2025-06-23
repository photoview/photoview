package test_utils

import (
	"fmt"

	"github.com/photoview/photoview/api/database"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type TestDBManager struct {
	DB *gorm.DB
}

func (dbm *TestDBManager) SetupOrReset() error {
	if dbm.DB == nil {
		if err := dbm.setup(); err != nil {
			return fmt.Errorf("setup db error: %w", err)
		}
	}

	if err := dbm.reset(); err != nil {
		return fmt.Errorf("reset db error: %w", err)
	}

	return nil
}

func (dbm *TestDBManager) Close() error {
	if dbm.DB == nil {
		return nil
	}

	sqlDB, err := dbm.DB.DB()
	if err != nil {
		return errors.Wrap(err, "get db instance when closing test database")
	}

	sqlDB.Close()
	dbm.DB = nil

	return nil
}

func (dbm *TestDBManager) setup() error {
	config := gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	}
	db, err := database.ConfigureDatabase(&config)
	if err != nil {
		return errors.Wrap(err, "configure test database")
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
