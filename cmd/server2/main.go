package main

import (
	"encoding/json"
	"log"
	"net/http"
)

func healthHandler(w http.ResponseWriter, r *http.Request) {
	response := map[string]string{
		"status": "ok",
		"node":   "server2",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func main() {
	http.HandleFunc("/health", healthHandler)

	log.Println("Server2 running on port 8081")
	log.Fatal(http.ListenAndServe(":8081", nil))
}
