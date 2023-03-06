package api

import (
	"log"
	"net/http"

	"github.com/song940/mycenter-go/models"
)

func (s *Server) Timeline(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		s.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	content := r.FormValue("content")
	user, ok := r.Context().Value("user").(*models.User)
	if !ok {
		s.Error(w, "user not found", http.StatusInternalServerError)
		return
	}
	post, err := models.CreatePost(s.db, user.Id, content)
	if err != nil {
		s.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Println(post)
	http.Redirect(w, r, "/", http.StatusFound)
}
