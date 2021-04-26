package test_utils

import (
	"github.com/photoview/photoview/api/database"
	"github.com/pkg/errors"
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
		return errors.Wrap(err, "get db instance when closing test database")
	}

	sqlDB.Close()
	dbm.DB = nil

	return nil
}

func (dbm *TestDBManager) setup() error {
	config := gorm.Config{}
	db, err := database.ConfigureDatabase(&config)
	if err != nil {
		return errors.Wrap(err, "configure test database")
	}

	if err := database.MigrateDatabase(db); err != nil {
		return errors.Wrap(err, "migrate test database")
	}

	dbm.DB = db

	if err := dbm.reset(); err != nil {
		return err
	}

	return nil
}

func (dbm *TestDBManager) reset() error {

	if err := database.ClearDatabase(dbm.DB); err != nil {
		return errors.Wrap(err, "reset test database")
	}

	return nil
}
