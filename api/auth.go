package api

import (
	"encoding/json"
	"net/http"

	"github.com/song940/mycenter-go/models"
)

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
			"app": app,
		})
		return
	}
	session, err := models.CreateSession(s.db, user.Id)
	if err != nil {
		s.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, app.Callback+"?code="+session.Id, http.StatusFound)
}

func (s *Server) Token(w http.ResponseWriter, r *http.Request) {
	clientId := r.URL.Query().Get("client_id")
	sessionId := r.URL.Query().Get("code")
	session, err := models.GetSessionByAppId(s.db, clientId, sessionId)
	if err != nil {
		s.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(session)
}
