package models

import (
	"database/sql"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Id        int       `json:"id"`
	Username  string    `json:"username"`
	Password  string    `json:"password"`
	CreatedAt time.Time `json:"created_at"`
}

func CreateUser(db *sql.DB, username, password string) (user *User, err error) {
	user = &User{}
	hashPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	err = db.QueryRow(`
		INSERT INTO users (username, password) 
		VALUES (?, ?)
		RETURNING id, username, password, created_at
	`, username, hashPassword).Scan(&user.Id, &user.Username, &user.Password, &user.CreatedAt)
	return
}

func GetUserByUserId(db *sql.DB, userId int) (*User, error) {
	row := db.QueryRow("SELECT id, username, password, created_at FROM users WHERE id = ?", userId)
	user := &User{}
	err := row.Scan(&user.Id, &user.Username, &user.Password, &user.CreatedAt)
	return user, err
}

func GetUserByUsername(db *sql.DB, username string) (*User, error) {
	row := db.QueryRow("SELECT id, username, password, created_at FROM users WHERE username = ?", username)
	user := &User{}
	err := row.Scan(&user.Id, &user.Username, &user.Password, &user.CreatedAt)
	return user, err
}

func UpdatePassword(db *sql.DB, userId int, new_password string) error {
	hashPassword, _ := bcrypt.GenerateFromPassword([]byte(new_password), bcrypt.DefaultCost)
	_, err := db.Exec("UPDATE users SET password = ? WHERE id = ?", hashPassword, userId)
	return err
}
