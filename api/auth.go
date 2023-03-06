package api

import (
	"encoding/json"
	"net/http"

	"github.com/song940/mycenter-go/models"
)

// http://localhost:8088/auth?client_id=2
func (s *Server) Auth(w http.ResponseWriter, r *http.Request) {
	clientId := r.URL.Query().Get("client_id")
	app, err := models.GetApp(s.db, clientId)
	if err != nil {
		s.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	user, ok := r.Context().Value("user").(*models.User)
	if !ok {
		s.Error(w, "Not login", http.StatusInternalServerError)
		return
	}
	if r.Method == http.MethodGet {
		s.Render(w, "auth", H{
			"user": user,
			"app":  app,
		})
		return
	}
	session, err := models.CreateSession(s.db, app.Id, user.Id)
	if err != nil {
		s.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, app.Callback+"?code="+session.Id, http.StatusFound)
}

// http://localhost:8088/token?client_id=2&client_secret=4da44c4f-c949-4c08-8761-b7aca42bf2d0&code=12
func (s *Server) Token(w http.ResponseWriter, r *http.Request) {
	qs := r.URL.Query()
	clientId := qs.Get("client_id")
	clientSecret := qs.Get("client_secret")
	app, err := models.GetApp(s.db, clientId)
	if err != nil {
		s.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if app.Secret != clientSecret {
		s.Error(w, "Invalid secret", http.StatusBadRequest)
		return
	}
	sessionId := qs.Get("code")
	session, err := models.GetSessionByAppId(s.db, clientId, sessionId)
	if err != nil {
		s.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(session)
}
