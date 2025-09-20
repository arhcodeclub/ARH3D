package main

import (
	"context"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
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
	srv := &http.Server{Addr: addr, Handler: mux}

	// Graceful shutdown.
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		<-c
		log.Println("Shutting down gracefully...")
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			log.Printf("Shutdown error: %v", err)
		}
	}()

	log.Printf("Server listening on %s", addr)
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("ListenAndServe(): %v", err)
	}
	log.Println("Server stopped")
}
