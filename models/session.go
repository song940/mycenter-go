package models

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type Session struct {
	AppName string `json:"-"`

	Id        string    `json:"id"`
	UserID    int       `json:"user_id"`
	Token     string    `json:"token"`
	CreatedAt time.Time `json:"created_at"`
}

func CreateSession(db *sql.DB, appId, userId int) (session *Session, err error) {
	token := uuid.New().String()
	session = &Session{}
	err = db.QueryRow(`
		INSERT INTO sessions (app_id, user_id, token)
		VALUES (?, ?, ?)
		RETURNING id, user_id, token, created_at
	`, appId, userId, token).Scan(&session.Id, &session.UserID, &session.Token, &session.CreatedAt)
	return
}

func GetSession(db *sql.DB, sessionId string) (session *Session, err error) {
	session = &Session{}
	err = db.QueryRow(`
		SELECT id, user_id, token, created_at 
		FROM sessions 
		WHERE id = ?
	`, sessionId).Scan(&session.Id, &session.UserID, &session.Token, &session.CreatedAt)
	return
}

func GetSessionByAppId(db *sql.DB, appId string, sessionId string) (session *Session, err error) {
	session = &Session{}
	err = db.QueryRow(`
		SELECT id, user_id, token, created_at
		FROM sessions
		WHERE app_id = ? AND id = ?
`, appId, sessionId).Scan(&session.Id, &session.UserID, &session.Token, &session.CreatedAt)
	return
}

func GetUserIdByToken(db *sql.DB, token string) (userId int, err error) {
	err = db.QueryRow("SELECT user_id FROM sessions WHERE token = ?", token).Scan(&userId)
	return
}

func GetUserByToken(db *sql.DB, token string) (user *User, err error) {
	user = &User{}
	err = db.QueryRow(`
		SELECT u.id, u.username, u.password, u.created_at 
		FROM sessions s, users u 
		WHERE s.token = ? and s.user_id = u.id
		`, token).Scan(&user.Id, &user.Username, &user.Password, &user.CreatedAt)
	return
}

func GetSessions(db *sql.DB, userId int64) (sessions []*Session, err error) {
	rows, err := db.Query("SELECT user_id, token, created_at FROM sessions WHERE user_id = ?", userId)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		session := &Session{}
		err = rows.Scan(&session.UserID, &session.Token, &session.CreatedAt)
		if err != nil {
			return
		}
		sessions = append(sessions, session)
	}
	return
}

func GetSessionsWithApp(db *sql.DB, userId int) (sessions []*Session, err error) {
	rows, err := db.Query(`
		SELECT s.id, a.name, s.token, s.created_at from sessions s, apps a 
		WHERE s.app_id = a.id AND s.user_id = ?
	`, userId)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		session := &Session{}
		err = rows.Scan(&session.Id, &session.AppName, &session.Token, &session.CreatedAt)
		if err != nil {
			return
		}
		sessions = append(sessions, session)
	}
	return
}

func DeleteSessionById(db *sql.DB, userId int, id string) (err error) {
	_, err = db.Exec("DELETE FROM sessions WHERE id = ? AND user_id = ?", id, userId)
	return
}
