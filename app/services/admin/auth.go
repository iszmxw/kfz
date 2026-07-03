package admin

import (
	"strings"

	"golang.org/x/crypto/bcrypt"
)

type PermissionSet map[string]struct{}

func NewPermissionSet(codes []string) PermissionSet {
	set := PermissionSet{}
	for _, code := range codes {
		code = strings.TrimSpace(code)
		if code == "" {
			continue
		}
		set[code] = struct{}{}
	}
	return set
}

func (p PermissionSet) Allows(code string) bool {
	code = strings.TrimSpace(code)
	if code == "" {
		return true
	}
	if _, ok := p["*"]; ok {
		return true
	}
	if _, ok := p[code]; ok {
		return true
	}
	if idx := strings.Index(code, ":"); idx > 0 {
		_, ok := p[code[:idx]+":*"]
		return ok
	}
	return false
}

func HashPassword(raw string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(raw), bcrypt.DefaultCost)
	return string(hash), err
}

func CheckPassword(hash, raw string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(raw)) == nil
}
