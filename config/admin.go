package config

import (
	"os"

	"goapi/pkg/config"
)

func init() {
	config.Add("admin", config.StrMap{
		"default_username":  envString("ADMIN_DEFAULT_USERNAME", "admin"),
		"default_password":  envString("ADMIN_DEFAULT_PASSWORD", ""),
		"token_ttl_seconds": config.Env("ADMIN_TOKEN_TTL_SECONDS", 86400),
	})
}

func envString(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
