package flags

import "flag"

var (
	// Database enables database-integration tests when the `-database` flag is passed to `go test`.
	Database bool

	// Filesystem enables filesystem-integration tests when the `-filesystem` flag is passed to `go test`.
	Filesystem bool
)

func init() {
	flag.BoolVar(&Database, "database", false, "run database integration tests")
	flag.BoolVar(&Filesystem, "filesystem", false, "run filesystem integration tests")
}
