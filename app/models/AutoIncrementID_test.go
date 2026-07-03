package models

import (
	"reflect"
	"testing"
)

func TestModelsUseAutoIncrementUintIDs(t *testing.T) {
	tests := []struct {
		name  string
		model any
	}{
		{"Book", Book{}},
		{"PriceSnapshot", PriceSnapshot{}},
		{"ScanLog", ScanLog{}},
		{"AdminUser", AdminUser{}},
		{"AdminRole", AdminRole{}},
		{"AdminUserRole", AdminUserRole{}},
		{"AdminMenu", AdminMenu{}},
		{"AdminAPIPermission", AdminAPIPermission{}},
		{"AdminRolePermission", AdminRolePermission{}},
		{"AdminOperationLog", AdminOperationLog{}},
		{"ManualDecision", ManualDecision{}},
		{"RecycleRule", RecycleRule{}},
		{"ImportTask", ImportTask{}},
		{"ImportTaskRow", ImportTaskRow{}},
	}

	uint64Type := reflect.TypeOf(uint64(0))
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field, ok := reflect.TypeOf(tt.model).FieldByName("ID")
			if !ok {
				t.Fatal("ID field not found")
			}
			if field.Type != uint64Type {
				t.Fatalf("ID type = %s, want uint64", field.Type)
			}
		})
	}
}
