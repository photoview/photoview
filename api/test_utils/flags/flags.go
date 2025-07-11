package flags

import "flag"

var (
	Database   bool
	Filesystem bool
)

func init() {
	flag.BoolVar(&Database, "database", false, "run database integration tests")
	flag.BoolVar(&Filesystem, "filesystem", false, "run filesystem integration tests")
}
