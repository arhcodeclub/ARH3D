package main

import (
    "log"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"
    "context"

    "github.com/arhcodeclub/arh3d/internal/db"
    "github.com/arhcodeclub/arh3d/internal/handlers"
)

func main() {
    db.Connect()

    mux := http.NewServeMux()

    // Serve static files (optional: CSS/JS)
    mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("internal/http/static"))))

    // Routes
    mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        err := handlers.RenderTemplate(w,
            "internal/http/templates/layout.html",
            "internal/http/templates/index.html",
        )
        if err != nil {
            http.Error(w, "Template error", 500)
        }
    })
    mux.HandleFunc("/new", handlers.NewRequestHandler)

    addr := ":8080"
    srv := &http.Server{Addr: addr, Handler: mux}

    // Graceful shutdown
    go func() {
        c := make(chan os.Signal, 1)
        signal.Notify(c, os.Interrupt, syscall.SIGTERM)
        <-c
        log.Println("Shutting down gracefully...")
        ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()
        if err := srv.Shutdown(ctx); err != nil {
            log.Printf("Shutdown error: %v", err)
        }
    }()

    log.Printf("Server running on %s", addr)
    if err := srv.ListenAndServe(); err != http.ErrServerClosed {
        log.Fatalf("ListenAndServe(): %v", err)
    }
}

