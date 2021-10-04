package main

import (
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
}
