package main

import (
	"embed"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
	"github.com/song940/mycenter-go/api"
)

//go:embed templates
var templatefiles embed.FS

func main() {

	server, err := api.NewServer()
	server.LoadTemplates(templatefiles)
	if err != nil {
		panic(err)
	}
	err = server.Init()
	if err != nil {
		panic(err)
	}
	http.ListenAndServe(":8088", server)
}
