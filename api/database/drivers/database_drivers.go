package drivers

import (
	"strings"

	"github.com/photoview/photoview/api/utils"
)

// DatabaseDriverType represents the name of a database driver
type DatabaseDriverType string

const (
	DatabaseDriverMysql    DatabaseDriverType = "mysql"
	DatabaseDriverSqlite   DatabaseDriverType = "sqlite"
	DatabaseDriverPostgres DatabaseDriverType = "postgres"
)

func DatabaseDriver() DatabaseDriverType {

	var driver DatabaseDriverType
	driverString := strings.ToLower(utils.EnvDatabaseDriver.GetValue())

	switch driverString {
	case "mysql":
		driver = DatabaseDriverMysql
	case "sqlite":
		driver = DatabaseDriverSqlite
	case "postgres":
		driver = DatabaseDriverPostgres
	default:
		driver = DatabaseDriverMysql
	}

	return driver
}
