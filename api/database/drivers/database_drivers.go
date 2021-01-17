package drivers

import (
	"os"
	"strings"
)

type DatabaseDriverType string

const (
	DatabaseDriverMysql  DatabaseDriverType = "mysql"
	DatabaseDriverSqlite DatabaseDriverType = "sqlite"
)

func DatabaseDriver() DatabaseDriverType {

	var driver DatabaseDriverType
	driverString := strings.ToLower(os.Getenv("PHOTOVIEW_DATABASE_DRIVER"))

	switch driverString {
	case "mysql":
		driver = DatabaseDriverMysql
	case "sqlite":
		driver = DatabaseDriverSqlite
	default:
		driver = DatabaseDriverMysql
	}

	return driver
}
