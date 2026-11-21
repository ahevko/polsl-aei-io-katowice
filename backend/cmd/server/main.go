package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
)

func main() {
	addr := flag.String("addr", ":8080", "adres nasłuchiwania HTTP")
	flag.Parse()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/ping", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "Pong"})
	})

	log.Printf("Serwer działa na http://localhost%s", *addr)
	log.Fatal(http.ListenAndServe(*addr, mux))
}
