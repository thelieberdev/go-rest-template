package database

import "fmt"

func UniqueConstraint(table string, field string) string {
	return fmt.Sprintf(
		`pq: duplicate key value violates unique constraint "%s_%s_key"`,
		table,
		field,
	)
}
