package test_utils

import (
	"flag"
	"log"
	"os"
	"testing"

	"github.com/joho/godotenv"
	"github.com/photoview/photoview/api/scanner/media_encoding/executable_worker"
	"github.com/photoview/photoview/api/test_utils/flags"
	"github.com/photoview/photoview/api/utils"
	"gorm.io/gorm"
)

var test_dbm TestDBManager = TestDBManager{}

func UnitTestRun(m *testing.M) {
	exitCode := 1
	defer func() {
		os.Exit(exitCode)
	}()

	flag.Parse()

	exitCode = m.Run()
}

func IntegrationTestRun(m *testing.M) {
	exitCode := 1
	defer func() {
		os.Exit(exitCode)
	}()

	flag.Parse()

	if flags.Database {
		envPath := PathFromAPIRoot("testing.env")

		if err := godotenv.Load(envPath); err != nil {
			log.Println("No testing.env file found")
		}
	}
	defer test_dbm.Close()

	faceModelsPath := PathFromAPIRoot("data", "models")
	utils.ConfigureTestFaceRecognitionModelsPath(faceModelsPath)

	terminateWorkers := executable_worker.Initialize()
	defer terminateWorkers()

	exitCode = m.Run()
}

func FilesystemTest(t *testing.T) {
	if !flags.Filesystem {
		t.Skip("Filesystem integration tests disabled")
	}
	utils.ConfigureTestCache(t.TempDir())
}

func DatabaseTest(t *testing.T) *gorm.DB {
	if !flags.Database {
		t.Skip("Database integration tests disabled")
	}

	if err := test_dbm.SetupAndReset(); err != nil {
		t.Fatalf("failed to setup or reset test database: %v", err)
	}

	return test_dbm.DB
}
