package models

import (
	"database/sql"

	"github.com/google/uuid"
)

func CreateInviteCode(db *sql.DB, source string) (code string, err error) {
	code = uuid.New().String()
	sql := `INSERT INTO invitation (source, code) VALUES (?, ?)`
	_, err = db.Exec(sql, source, code)
	return
}

func VerifyInviteCode(db *sql.DB, code string) (pass bool) {
	sql := `SELECT code, source FROM invitation WHERE code = ?`
	_, err := db.Exec(sql, code)
	if err != nil {
		return false
	}
	return true
}
