package test_utils

import (
	"flag"
	"log"
	"path"
	"runtime"
	"testing"

	"github.com/joho/godotenv"
	"github.com/photoview/photoview/api/scanner/media_encoding/executable_worker"
	"github.com/photoview/photoview/api/test_utils/flags"
	"github.com/photoview/photoview/api/utils"
	"gorm.io/gorm"
)

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

	if flags.Database {

		envPath := path.Join(path.Dir(file), "..", "testing.env")

		if err := godotenv.Load(envPath); err != nil {
			log.Println("No testing.env file found")
		}
	}

	faceModelsPath := path.Join(path.Dir(file), "..", "data", "models")
	utils.ConfigureTestFaceRecognitionModelsPath(faceModelsPath)

	terminateWorkers := executable_worker.Initialize()
	defer terminateWorkers()

	result := m.Run()

	test_dbm.Close()

	return result
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
