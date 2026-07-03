package admin

import "testing"

func TestPermissionSetAllowsExactCodeAndWildcard(t *testing.T) {
	permissions := NewPermissionSet([]string{"dashboard:read", "book:*"})

	if !permissions.Allows("dashboard:read") {
		t.Fatal("dashboard:read should be allowed")
	}
	if !permissions.Allows("book:write") {
		t.Fatal("book:write should be allowed by book:*")
	}
	if permissions.Allows("rule:write") {
		t.Fatal("rule:write should not be allowed")
	}
}

func TestPasswordHashVerifiesAndRejectsWrongPassword(t *testing.T) {
	hash, err := HashPassword("secret123")
	if err != nil {
		t.Fatalf("HashPassword err=%v", err)
	}
	if hash == "secret123" {
		t.Fatal("hash should not equal raw password")
	}
	if !CheckPassword(hash, "secret123") {
		t.Fatal("expected password to verify")
	}
	if CheckPassword(hash, "wrong") {
		t.Fatal("wrong password should not verify")
	}
}
