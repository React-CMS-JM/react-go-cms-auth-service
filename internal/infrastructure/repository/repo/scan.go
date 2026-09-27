package repo

import (
	"database/sql"
	"time"
)

// InstantFrom converts a nullable SQL timestamp. An invalid value stays nil so the response formatter emits JSON null.
func InstantFrom(value sql.NullTime) *time.Time {
	if !value.Valid {
		return nil
	}
	var instant time.Time
	instant = value.Time
	return &instant
}

// LocalFrom converts a nullable SQL local datetime. An invalid value stays nil so the response formatter emits JSON null.
func LocalFrom(value sql.NullTime) *time.Time {
	if !value.Valid {
		return nil
	}
	var local time.Time
	local = value.Time
	return &local
}

// StrPtr converts a nullable SQL string. An invalid value stays nil.
func StrPtr(value sql.NullString) *string {
	if !value.Valid {
		return nil
	}
	var text string
	text = value.String
	return &text
}
