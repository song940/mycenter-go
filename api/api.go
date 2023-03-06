package api

import (
	"context"
	"database/sql"
	"embed"
	"html/template"
	"log"
	"net/http"

	"github.com/song940/mycenter-go/models"
)

type Server struct {
	db       *sql.DB
	template *template.Template
}

type H map[string]interface{}

func authMiddleware(db *sql.DB, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		cookie, err := r.Cookie("token")
		if err == nil {
			user, _ := models.GetUserByToken(db, cookie.Value)
			ctx = context.WithValue(ctx, "user", user)
		}
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (server *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	mux := http.NewServeMux()
	mux.HandleFunc("/join", server.Join)
	mux.HandleFunc("/signup", server.Signup)
	mux.HandleFunc("/login", server.Login)
	mux.HandleFunc("/logout", server.Logout)
	mux.HandleFunc("/posts", server.Timeline)
	mux.HandleFunc("/apps", server.Apps)
	mux.HandleFunc("/auth", server.Auth)
	mux.HandleFunc("/settings", server.Settings)
	mux.HandleFunc("/settings/account", server.Account)
	mux.HandleFunc("/settings/profile", server.Profile)
	mux.HandleFunc("/", server.Home)
	// auth
	authMux := authMiddleware(server.db, mux)
	authMux.ServeHTTP(w, r)
}

func NewServer() (server *Server, err error) {
	dsn := "mycenter.db"
	db, err := sql.Open("sqlite3", dsn)
	server = &Server{db: db}
	return
}

func (s *Server) Init() (err error) {
	// dsn := "file::memory:?cache=shared"
	_, err = s.db.Exec(`
		CREATE TABLE IF NOT EXISTS apps (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			description TEXT NOT NULL,
			homepage TEXT NOT NULL,
			callback TEXT NOT NULL,
			secret TEXT NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		log.Fatal(err)
	}
	_, err = s.db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT UNIQUE NOT NULL,
			password TEXT NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`)
	if err != nil {
		log.Fatal(err)
	}
	_, err = s.db.Exec(`
		CREATE TABLE IF NOT EXISTS sessions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			app_id INTEGER NOT NULL,
			user_id INTEGER NOT NULL,
			token TEXT UNIQUE NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES user (id)
			FOREIGN KEY (app_id) REFERENCES apps (id)
		)`)
	if err != nil {
		log.Fatal(err)
	}
	_, err = s.db.Exec(`
		CREATE TABLE IF NOT EXISTS profile (
			user_id INTEGER,
			key     TEXT,
			value   TEXT,
			PRIMARY KEY (user_id, key),
			FOREIGN KEY (user_id) REFERENCES user (id)
		)`)
	if err != nil {
		log.Fatal(err)
	}
	_, err = s.db.Exec(`
		CREATE TABLE IF NOT EXISTS invitation (
			source TEXT NOT NULL,
			code TEXT NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (source, code)
		)`)
	if err != nil {
		log.Fatal(err)
	}
	_, err = s.db.Exec(`
		CREATE TABLE IF NOT EXISTS posts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER,
			content TEXT,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES user (id)
		)`)
	if err != nil {
		log.Fatal(err)
	}

	return
}

func (s *Server) LoadTemplates(templatefiles embed.FS) {
	s.template = template.Must(template.ParseGlob("templates/*.html"))
}

func (s *Server) Render(w http.ResponseWriter, name string, data H) {
	w.Header().Add("Content-Type", "text/html")
	if err := s.template.ExecuteTemplate(w, name+".html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) Error(w http.ResponseWriter, err string, status int) {
	w.Header().Add("Content-Type", "text/html")
	w.WriteHeader(status)
	s.Render(w, "error", H{
		"error": err,
	})
}

func (s *Server) Home(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value("user").(*models.User)
	if !ok {
		s.Render(w, "home", nil)
		return
	}
	posts, err := models.GetPosts(s.db, user.Id)
	if err != nil {
		s.Error(w, err.Error(), http.StatusInternalServerError)
	}
	sessions, err := models.GetSessionsWithApp(s.db, user.Id)
	if err != nil {
		s.Error(w, err.Error(), http.StatusInternalServerError)
	}
	s.Render(w, "home", H{
		"user":     user,
		"posts":    posts,
		"sessions": sessions,
	})
}

func (s *Server) Settings(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/settings/account", http.StatusFound)
}

func (s *Server) Account(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		s.Render(w, "account", nil)
		return
	}
	user, ok := r.Context().Value("user").(*models.User)
	if !ok {
		s.Error(w, "user not login", http.StatusInternalServerError)
	}
	if err := r.ParseForm(); err != nil {
		s.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// old_password := r.PostFormValue("old_password")
	new_password := r.PostFormValue("new_password")
	confirm_password := r.PostFormValue("confirm_password")
	if new_password != confirm_password {
		s.Error(w, "passwords do not match", http.StatusBadRequest)
		return
	}
	err := models.UpdatePassword(s.db, user.Id, new_password)
	if err != nil {
		s.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/", http.StatusFound)
}

func (s *Server) Profile(w http.ResponseWriter, r *http.Request) {
	s.Render(w, "profile", nil)
}
