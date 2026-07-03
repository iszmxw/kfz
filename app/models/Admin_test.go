package models

import (
	"sync"
	"testing"

	"gorm.io/gorm/schema"
)

func TestImportTaskRowUsesSafeRowNumberColumn(t *testing.T) {
	parsed, err := schema.Parse(&ImportTaskRow{}, &sync.Map{}, schema.NamingStrategy{})
	if err != nil {
		t.Fatalf("parse ImportTaskRow schema: %v", err)
	}

	field := parsed.LookUpField("RowNumber")
	if field == nil {
		t.Fatal("RowNumber field not found")
	}
	if field.DBName != "row_no" {
		t.Fatalf("RowNumber DBName = %q, want row_no", field.DBName)
	}
}
