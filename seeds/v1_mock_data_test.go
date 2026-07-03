package seeds

import (
	"os"
	"regexp"
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestV1MockDataIncludesAdminLoginUser(t *testing.T) {
	raw, err := os.ReadFile("v1_mock_data.sql")
	if err != nil {
		t.Fatalf("read seed file: %v", err)
	}
	sql := string(raw)

	requiredSnippets := []string{
		"INSERT INTO t_admin_user",
		"'admin'",
		"INSERT INTO t_admin_user_role",
		"r.code = 'SUPER_ADMIN'",
	}
	for _, snippet := range requiredSnippets {
		if !strings.Contains(sql, snippet) {
			t.Fatalf("seed file missing %q", snippet)
		}
	}

	hashPattern := regexp.MustCompile(`'\$2[aby]\$[^']+'`)
	hashLiteral := hashPattern.FindString(sql)
	if hashLiteral == "" {
		t.Fatal("seed file missing bcrypt password hash")
	}
	hash := strings.Trim(hashLiteral, "'")
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte("admin123")); err != nil {
		t.Fatalf("seed admin password hash does not verify admin123: %v", err)
	}
	for _, insert := range []string{
		"INSERT INTO t_admin_user (\n  id,",
		"INSERT INTO t_admin_role (\n  id,",
		"INSERT INTO t_admin_menu (\n  id,",
		"INSERT INTO t_admin_role_permission (\n  id,",
	} {
		if strings.Contains(sql, insert) {
			t.Fatalf("seed file should not specify auto increment primary key in %q", insert)
		}
	}
}

func TestV1MockDataDoesNotSpecifyStringPrimaryIDs(t *testing.T) {
	raw, err := os.ReadFile("v1_mock_data.sql")
	if err != nil {
		t.Fatalf("read seed file: %v", err)
	}
	sql := string(raw)
	for _, forbidden := range []string{
		"'admin_seed_admin'",
		"'role_super_admin'",
		"'role_operator'",
		"'menu_",
		"'rp_",
		"'rule_default_v1'",
		"'price_978",
	} {
		if strings.Contains(sql, forbidden) {
			t.Fatalf("seed file still contains string primary key marker %q", forbidden)
		}
	}
}
