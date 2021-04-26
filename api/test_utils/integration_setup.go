package test_utils

import (
	"flag"
	"log"
	"path"
	"runtime"
	"testing"

	"github.com/joho/godotenv"
	"github.com/photoview/photoview/api/utils"
	"gorm.io/gorm"
)

type integration_options struct {
	Database   *bool
	Filesystem *bool
}

var integration_flags integration_options = integration_options{
	Database:   flag.Bool("database", false, "run database integration tests"),
	Filesystem: flag.Bool("filesystem", false, "run filesystem integration tests"),
}

var test_dbm TestDBManager = TestDBManager{}

func UnitTestRun(m *testing.M) int {
	flag.Parse()
	return m.Run()
}

func IntegrationTestRun(m *testing.M) int {
	flag.Parse()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		log.Fatal("could not get runtime file path")
	}

	if *integration_flags.Database {

		envPath := path.Join(path.Dir(file), "..", "testing.env")

		if err := godotenv.Load(envPath); err != nil {
			log.Println("No testing.env file found")
		}
	}

	faceModelsPath := path.Join(path.Dir(file), "..", "data", "models")
	utils.ConfigureTestFaceRecognitionModelsPath(faceModelsPath)

	result := m.Run()

	test_dbm.Close()

	return result
}

func FilesystemTest(t *testing.T) {
	if !*integration_flags.Filesystem {
		t.Skip("Filesystem integration tests disabled")
	}
	utils.ConfigureTestCache(t.TempDir())
}

func DatabaseTest(t *testing.T) *gorm.DB {
	if !*integration_flags.Database {
		t.Skip("Database integration tests disabled")
	}

	if err := test_dbm.SetupOrReset(); err != nil {
		t.Fatalf("failed to setup or reset test database: %v", err)
	}

	return test_dbm.DB
}
