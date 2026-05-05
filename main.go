package main

import (
	"crypto/rand"
	"encoding/hex"
	"flag"
	"log"
	"net/http"
)

func main() {
	addr := flag.String("addr", ":8765", "listen address")
	token := flag.String("token", "", "auth token (auto-generated if empty)")
	flag.Parse()

	authToken := *token
	if authToken == "" {
		buf := make([]byte, 16)
		rand.Read(buf)
		authToken = hex.EncodeToString(buf)
	}

	sm := NewSessionManager()

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		if r.Form.Get("token") != authToken {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		handleWS(sm, w, r)
	})

	log.Printf("wand-agent listening on %s", *addr)
	log.Printf("auth token: %s", authToken)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
