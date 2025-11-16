package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/arhcodeclub/arh3d/internal/auth"
	"github.com/arhcodeclub/arh3d/internal/db"
	"github.com/arhcodeclub/arh3d/internal/handlers"
	"github.com/arhcodeclub/arh3d/internal/models"
)

func main() {
	log.Printf("Starting service.")

	db.Connect()

	mux := http.NewServeMux()

	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("internal/http/static"))))

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[HTTP] %s %s", r.Method, r.URL.Path)

		data := map[string]any{}

		if sess, err := auth.GetSession(r); err == nil {
			var user models.User
			if err := db.DB.First(&user, sess.UserID).Error; err == nil {
				var activeCount int64
				if err := db.DB.Model(&models.PrintRequest{}).
					Where("user_id = ? AND status IN ?", user.ID, []string{"in_queue", "printing"}).
					Count(&activeCount).Error; err != nil {
					log.Printf("[HTTP] Error counting active requests for user_id=%d: %v", user.ID, err)
					activeCount = 0
				}

				data["User"] = user
				data["ActiveCount"] = activeCount
			} else {
				log.Printf("[HTTP] Could not load user for session user_id=%d: %v", sess.UserID, err)
			}
		}

		if err := handlers.RenderTemplateWithData(w,
			data,
			"internal/http/templates/layout.html",
			"internal/http/templates/index.html",
		); err != nil {
			log.Printf("[HTTP] Template error on /: %v", err)
			http.Error(w, "Template error", 500)
		}
	})

	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[HTTP] %s /login.", r.Method)
		if r.Method == http.MethodGet {
			handlers.LoginPage(w, r)
		} else if r.Method == http.MethodPost {
			handlers.LoginRequest(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/login/verify", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[HTTP] %s /login/verify.", r.Method)
		handlers.LoginVerify(w, r)
	})

	mux.HandleFunc("/onboarding", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[HTTP] %s /onboarding.", r.Method)
		if r.Method == http.MethodGet {
			handlers.OnboardingPage(w, r)
		} else if r.Method == http.MethodPost {
			handlers.OnboardingSubmit(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[HTTP] %s /logout.", r.Method)
		handlers.Logout(w, r)
	})

	mux.Handle("/new", auth.RequireAuth(http.HandlerFunc(handlers.NewRequestHandler)))
	mux.Handle("/status", auth.RequireAuth(http.HandlerFunc(handlers.StatusHandler)))

	mux.Handle("/admin", http.HandlerFunc(handlers.AdminHandler))
	mux.Handle("/admin/update", http.HandlerFunc(handlers.AdminHandler))

	addr := ":8080"
	srv := &http.Server{Addr: addr, Handler: mux}

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		<-c
		log.Println("[SERVER] Shutting down gracefully...")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			log.Println("[SERVER] Forced shutdown:", err)
		}
	}()

	log.Printf("[SERVER] Listening on %s.", addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal("[SERVER] ListenAndServe:", err)
	}
}
