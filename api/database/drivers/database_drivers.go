package drivers

import (
	"strings"

	"github.com/kkovaletp/photoview/api/utils"
	"gorm.io/gorm"
)

// DatabaseDriverType represents the name of a database driver
type DatabaseDriverType string

const (
	MYSQL    DatabaseDriverType = "mysql"
	SQLITE   DatabaseDriverType = "sqlite"
	POSTGRES DatabaseDriverType = "postgres"
)

func DatabaseDriverFromEnv() DatabaseDriverType {

	var driver DatabaseDriverType
	driverString := strings.ToLower(utils.EnvDatabaseDriver.GetValue())

	switch driverString {
	case "mysql":
		driver = MYSQL
	case "sqlite":
		driver = SQLITE
	case "postgres":
		driver = POSTGRES
	default:
		driver = MYSQL
	}

	return driver
}

func (driver DatabaseDriverType) MatchDatabase(db *gorm.DB) bool {
	return db.Dialector.Name() == string(driver)
}

func GetDatabaseDriverType(db *gorm.DB) (driver DatabaseDriverType) {
	switch db.Dialector.Name() {
	case "mysql":
		driver = MYSQL
	case "sqlite":
		driver = SQLITE
	case "postgres":
		driver = POSTGRES
	default:
		driver = MYSQL
	}

	return
}
