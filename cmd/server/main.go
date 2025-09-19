package main

import (
	"html/template"
	"log"
	"net/http"
)

func main() {
	tmpl := template.Must(template.ParseFiles(
		"internal/http/templates/layout.html",
		"internal/http/templates/index.html",
	))

	// ServeMux compares and calls the requested handler.
	// https://pkg.go.dev/net/http#ServeMux
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Render the root template.
		if err := tmpl.ExecuteTemplate(w, "index.html", nil); err != nil {
				http.Error(w, "template error", 500)
				return
		}
	})

	addr := ":8080"
	log.Printf("Server listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}