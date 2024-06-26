package api

import (
	"context"
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"os"

	_ "github.com/mattn/go-sqlite3"
	"github.com/song940/mycenter-go/models"
	"github.com/song940/mycenter-go/templates"
	"gopkg.in/yaml.v3"
)

type Config struct {
	AppId       int    `yaml:"app_id"`
	Address     string `yaml:"listen"`
	Title       string `yaml:"title"`
	Description string `yaml:"description"`
	Database    struct {
		Driver   string `yaml:"driver"`
		Database string `yaml:"database"`
	} `yaml:"database"`
}

func LoadConfig(filename string) (config *Config, err error) {
	config = &Config{}
	data, err := os.ReadFile(filename)
	if err != nil {
		return
	}
	err = yaml.Unmarshal(data, &config)
	return
}

type H map[string]interface{}

type Server struct {
	config   *Config
	db       *sql.DB
	template *template.Template
}

func NewServer(config *Config) (server *Server) {
	server = &Server{config: config}
	return
}

func authMiddleware(db *sql.DB, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		token := r.Header.Get("Authorization")
		if cookie, err := r.Cookie("token"); token == "" && err == nil {
			token = cookie.Value
		}
		user, err := models.GetUserByToken(db, token)
		if err == nil {
			ctx = context.WithValue(ctx, "user", user)
		}
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (server *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", server.Home)
	mux.HandleFunc("/join", server.Join)
	mux.HandleFunc("/signup", server.Signup)
	mux.HandleFunc("/login", server.Login)
	mux.HandleFunc("/logout", server.Logout)
	mux.HandleFunc("/posts", server.Timeline)
	mux.HandleFunc("/users", server.Users)
	mux.HandleFunc("/apps", server.Apps)
	mux.HandleFunc("/auth", server.Auth)
	mux.HandleFunc("/token", server.Token)
	mux.HandleFunc("/settings", server.Settings)
	mux.HandleFunc("/settings/account", server.Account)
	mux.HandleFunc("/settings/profile", server.Profile)
	// auth
	authMux := authMiddleware(server.db, mux)
	authMux.ServeHTTP(w, r)
}

func (s *Server) Init() (err error) {
	db, err := sql.Open(s.config.Database.Driver, s.config.Database.Database)
	if err != nil {
		return
	}
	_, err = db.Exec(`
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
		return
	}
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT UNIQUE NOT NULL,
			password TEXT NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`)
	if err != nil {
		return
	}
	_, err = db.Exec(`
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
		return
	}
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS profile (
			user_id INTEGER,
			key     TEXT,
			value   TEXT,
			PRIMARY KEY (user_id, key),
			FOREIGN KEY (user_id) REFERENCES user (id)
		)`)
	if err != nil {
		return
	}
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS invitation (
			source TEXT NOT NULL,
			code TEXT NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (source, code)
		)`)
	if err != nil {
		return
	}
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS posts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER,
			content TEXT,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES user (id)
		)`)
	if err != nil {
		return
	}
	s.db = db
	return
}

// Render renders an HTML template with the provided data.
func (s *Server) Render(w http.ResponseWriter, templateName string, data H) {
	if data == nil {
		data = H{}
	}
	data["Title"] = s.config.Title
	data["Description"] = s.config.Description
	// tmpl, err := template.ParseFiles("templates/layout.html", "templates/"+templateName+".html")
	// Parse templates from embedded file system
	tmpl, err := template.New("").ParseFS(templates.Files, "layout.html", templateName+".html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Execute "index.html" within the layout and write to response
	err = tmpl.ExecuteTemplate(w, "layout", data)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
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
