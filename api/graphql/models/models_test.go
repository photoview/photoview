package models_test

import (
	"testing"

	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/test_utils"
	"gorm.io/gorm"
)

func TestMain(m *testing.M) {
	test_utils.IntegrationTestRun(m)
}

func TestDateShot(t *testing.T) {
	db := test_utils.DatabaseTest(t)

	s := gorm.Statement{DB: db}
	if err := s.Parse(&models.MediaEXIF{}); err != nil {
		t.Fatal("can't parse:", err)
	}

	for _, name := range []string{"DateShot", "Maker"} {
		field := s.Schema.FieldsByName[name]
		if field == nil {
			t.Fatal("no field")
		}

		t.Error("table:", s.Schema.Table)
		t.Error("field:", field.DBName)
		t.Error("type:", field.DataType, field.GORMDataType)
	}

	cts, err := db.Migrator().ColumnTypes(&models.MediaEXIF{})
	if err != nil {
		t.Fatal("can't get table:", err)
	}

	for _, ct := range cts {
		if ct.Name() == "date_shot" {
			ctt, ok := ct.ColumnType()
			t.Error("column type:", ctt, ok)
		}
		if ct.Name() == "maker" {
			ctt, ok := ct.ColumnType()
			t.Error("column type:", ctt, ok)
		}
	}
}
