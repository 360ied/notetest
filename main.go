package main

import (
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"notetest/notes"
)

func main() {
	var n *notes.Notes

	m := http.NewServeMux()

	m.HandleFunc("/api/get_note", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		q := r.URL.Query()

		name := q.Get("name")

		if n == nil {
			// n isn't even initialized yet wtf
			log.Printf("[WTF] webapp called /api/get_note before initializing db")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		content, found := n.ViewNote(name)
		if !found {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, content)
	})

	m.HandleFunc("/api/update_note", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		if n == nil {
			// n isn't even initialized yet wtf
			log.Printf("[WTF] webapp called /api/update_note before initializing db")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		jd := json.NewDecoder(r.Body)

		nu := notes.NotesUpdate{}

		if err := jd.Decode(&nu); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		n.UpdateNote(nu)

		w.WriteHeader(http.StatusOK)
	})

	m.HandleFunc("/api/open_db", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		keyHex := r.URL.Query().Get("key")
		key, err := hex.DecodeString(keyHex)
		if err != nil && len(key) != 32 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		ndb, err := notes.UnlockDB(r.Body, key)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		n = ndb

		w.WriteHeader(http.StatusOK)
	})
}
