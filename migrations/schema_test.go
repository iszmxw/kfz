package migrations

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestMigrationsOnlyKeepFinalLocalSchema(t *testing.T) {
	files, err := filepath.Glob("*.sql")
	if err != nil {
		t.Fatalf("list sql migrations: %v", err)
	}
	if len(files) != 1 || files[0] != "001_init.sql" {
		t.Fatalf("local dev should keep only final schema 001_init.sql, got %v", files)
	}
}

func TestInitSchemaUsesAutoIncrementIDs(t *testing.T) {
	raw, err := os.ReadFile("001_init.sql")
	if err != nil {
		t.Fatalf("read init schema: %v", err)
	}
	sql := string(raw)

	for _, table := range []string{
		"t_book",
		"t_price_snapshot",
		"t_scan_log",
		"t_admin_user",
		"t_admin_role",
		"t_admin_user_role",
		"t_admin_menu",
		"t_admin_api_permission",
		"t_admin_role_permission",
		"t_admin_operation_log",
		"t_manual_decision",
		"t_recycle_rule",
		"t_import_task",
		"t_import_task_row",
		"t_kongfz_collect_task",
		"t_kongfz_collect_raw_row",
	} {
		pattern := regexp.MustCompile(`(?is)CREATE TABLE IF NOT EXISTS ` + table + ` \(\s*id bigint unsigned NOT NULL AUTO_INCREMENT`)
		if !pattern.MatchString(sql) {
			t.Fatalf("%s does not use bigint unsigned auto increment id", table)
		}
	}
	if regexp.MustCompile(`(?im)^\s*id varchar\(36\)`).MatchString(sql) {
		t.Fatal("schema still contains string primary id")
	}
	if !strings.Contains(sql, "UNIQUE KEY uk_book_isbn (isbn)") {
		t.Fatal("book schema must keep isbn as a unique business key")
	}
	if !strings.Contains(sql, "cover_url varchar(1000) NULL") {
		t.Fatal("schema must include book cover url fields")
	}
	if !regexp.MustCompile(`(?is)CREATE TABLE IF NOT EXISTS t_kongfz_collect_raw_row .*cover_url varchar\(1000\) NULL`).MatchString(sql) {
		t.Fatal("kongfz raw row schema must keep collected cover url")
	}
}
