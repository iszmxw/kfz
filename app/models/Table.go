package models

import "goapi/pkg/config"

func tableName(name string) string {
	return config.GetString("database.mysql.prefix") + name
}
