package migrations

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

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
}
