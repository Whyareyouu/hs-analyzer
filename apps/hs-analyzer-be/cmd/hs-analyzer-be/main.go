package main

import (
	"log"
	"net/http"

	"hsanalyzer/apps/hs-analyzer-be/internal/api"
)

func main() {
	store := api.NewMemoryStore()
	server := api.NewServer(store)

	log.Println("hs-analyzer-be listening on 127.0.0.1:8080")
	if err := http.ListenAndServe("127.0.0.1:8080", server.Routes()); err != nil {
		log.Fatal(err)
	}
}
