package models

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type App struct {
	Id          int
	Name        string
	Description string
	Secret      string
	Homepage    string
	Callback    string
	CreatedAt   time.Time
}

func GetApps(db *sql.DB) (apps []App, err error) {
	sql := `SELECT id, name, description, secret, homepage, callback, created_at FROM apps`
	rows, err := db.Query(sql)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var app App
		err = rows.Scan(&app.Id, &app.Name, &app.Description, &app.Secret, &app.Homepage, &app.Callback, &app.CreatedAt)
		if err != nil {
			return
		}
		apps = append(apps, app)
	}
	return
}

func CreateApp(db *sql.DB, app App) (App, error) {
	app.Secret = uuid.New().String()
	sql := `INSERT INTO apps (name, description, homepage, callback, secret) VALUES (?, ?, ?, ?, ?)`
	_, err := db.Exec(sql, app.Name, app.Description, app.Homepage, app.Callback, app.Secret)
	return app, err
}

func GetApp(db *sql.DB, clientId string) (app App, err error) {
	sql := `SELECT id, name, description, secret, homepage, callback, created_at FROM apps WHERE id = ?`
	err = db.QueryRow(sql, clientId).Scan(&app.Id, &app.Name, &app.Description, &app.Secret, &app.Homepage, &app.Callback, &app.CreatedAt)
	return
}
