package test_utils

import (
	"fmt"

	"github.com/photoview/photoview/api/database"
	"gorm.io/gorm"
)

type TestDBManager struct {
	DB *gorm.DB
}

func (dbm *TestDBManager) SetupOrReset() error {
	if dbm.DB == nil {
		return dbm.setup()
	} else {
		return dbm.reset()
	}
}

func (dbm *TestDBManager) Close() error {
	if dbm.DB == nil {
		return nil
	}

	if err := dbm.reset(); err != nil {
		return err
	}

	sqlDB, err := dbm.DB.DB()
	if err != nil {
		return fmt.Errorf("get db instance when closing test database: %w", err)
	}

	sqlDB.Close()
	dbm.DB = nil

	return nil
}

func (dbm *TestDBManager) setup() error {
	config := gorm.Config{}
	db, err := database.ConfigureDatabase(&config)
	if err != nil {
		return fmt.Errorf("configure test database: %w")
	}

	if err := database.MigrateDatabase(db); err != nil {
		return fmt.Errorf("migrate test database: %w")
	}

	dbm.DB = db

	if err := dbm.reset(); err != nil {
		return err
	}

	return nil
}

func (dbm *TestDBManager) reset() error {

	if err := database.ClearDatabase(dbm.DB); err != nil {
		return fmt.Errorf("reset test database: %w", err)
	}

	return nil
}
