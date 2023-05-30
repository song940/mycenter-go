package api

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/song940/mycenter-go/models"
	"golang.org/x/crypto/bcrypt"
)

func (s *Server) Join(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		s.Render(w, "join", nil)
		return
	}
	if err := r.ParseForm(); err != nil {
		s.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	source := r.FormValue("source")
	code, err := models.CreateInviteCode(s.db, source)
	if err != nil {
		s.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/signup?code="+code, http.StatusFound)
}

func (s *Server) Signup(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		code := r.URL.Query().Get("code")
		s.Render(w, "signup", H{
			"code": code,
		})
		return
	}
	if err := r.ParseForm(); err != nil {
		s.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	code := r.FormValue("code")
	username := r.FormValue("username")
	password := r.FormValue("password")
	ok := models.VerifyInviteCode(s.db, code)
	if !ok {
		s.Error(w, "Invalid invitation code", http.StatusBadRequest)
		return
	}
	user, err := models.CreateUser(s.db, username, password)
	if err != nil {
		s.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Println(user)
	http.Redirect(w, r, "/login", http.StatusFound)
}

func (s *Server) Login(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodGet {
		s.Render(w, "login", nil)
		return
	}

	if err := r.ParseForm(); err != nil {
		s.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	username := r.FormValue("username")
	password := r.FormValue("password")
	user, err := models.GetUserByUsername(s.db, username)
	if err != nil {
		s.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		s.Error(w, "Invalid username or password", http.StatusBadRequest)
		return
	}
	session, err := models.CreateSession(s.db, user.Id)
	if err != nil {
		s.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    session.Token,
		HttpOnly: true,
		Path:     "/",
	})
	http.Redirect(w, r, "/", http.StatusFound)
}

func (s *Server) Logout(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value("user").(*models.User)
	if !ok {
		s.Error(w, "user not login", http.StatusInternalServerError)
	}
	sessionId := r.URL.Query().Get("id")
	if sessionId == "" {
		http.SetCookie(w, &http.Cookie{
			Name:     "token",
			Value:    "",
			HttpOnly: true,
			Path:     "/",
			Expires:  time.Now().Add(-time.Hour),
		})
	} else {
		err := models.DeleteSessionById(s.db, user.Id, sessionId)
		if err != nil {
			s.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	http.Redirect(w, r, "/", http.StatusFound)
}

func (s *Server) Users(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value("user").(*models.User)
	if !ok {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("user not login"))
	} else {
		json.NewEncoder(w).Encode(user)
	}
}
