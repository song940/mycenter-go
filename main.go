package main

import (
	"flag"
	"net/http"

	"github.com/song940/mycenter-go/api"
)

func main() {

	var configFile string
	flag.StringVar(&configFile, "config", "config.yaml", "config file")
	flag.Parse()

	config, err := api.LoadConfig(configFile)
	if err != nil {
		panic(err)
	}
	server := api.NewServer(config)
	server.Init()
	server.LoadTemplates()
	http.ListenAndServe(config.Address, server)
}
