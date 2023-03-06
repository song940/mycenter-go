package api

import (
	"log"
	"net/http"

	"github.com/song940/mycenter-go/models"
)

func (s *Server) Apps(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		apps, err := models.GetApps(s.db)
		if err != nil {
			s.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		s.Render(w, "app", H{
			"apps": apps,
		})
		return
	}
	if err := r.ParseForm(); err != nil {
		s.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	name := r.FormValue("name")
	description := r.FormValue("description")
	homepage := r.FormValue("homepage")
	callback := r.FormValue("callback")
	app := models.App{
		Name:        name,
		Description: description,
		Homepage:    homepage,
		Callback:    callback,
	}
	app, err := models.CreateApp(s.db, app)
	if err != nil {
		s.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Println(app)
	http.Redirect(w, r, "/apps", http.StatusFound)
}
