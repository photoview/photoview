package test_utils

import (
	"github.com/photoview/photoview/api/database"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type TestDBManager struct {
	db *gorm.DB
	TX *gorm.DB
}

func (dbm *TestDBManager) SetupOrReset(withTransaction bool) error {
	if dbm.db == nil {
		if err := dbm.setup(); err != nil {
			return err
		}
	}

	if dbm.TX != nil {
		dbm.TX.Rollback()
		dbm.TX = nil
	}
	if withTransaction {
		dbm.TX = dbm.db.Begin()
	}

	return nil
}

func (dbm *TestDBManager) Close() error {
	if dbm.db == nil {
		return nil
	}

	sqlDB, err := dbm.db.DB()
	if err != nil {
		return errors.Wrap(err, "get db instance when closing test database")
	}
	if dbm.TX != nil {
		dbm.TX.Rollback()
	}

	sqlDB.Close()

	dbm.db = nil
	dbm.TX = nil

	return nil
}

func (dbm *TestDBManager) setup() error {
	config := &gorm.Config{}
	db, err := database.ConfigureDatabase(config)
	if err != nil {
		return errors.Wrap(err, "configure test database")
	}

	if err := database.MigrateDatabase(db); err != nil {
		return errors.Wrap(err, "migrate test database")
	}

	dbm.db = db

	return nil
}
