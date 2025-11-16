package handlers

import (
	"log"
	"net/http"

	"github.com/arhcodeclub/arh3d/internal/auth"
	"github.com/arhcodeclub/arh3d/internal/db"
	"github.com/arhcodeclub/arh3d/internal/models"
)

func StatusHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	log.Printf("[STATUS] GET /status.")

	sess, err := auth.GetSession(r)
	if err != nil {
		log.Printf("[STATUS] No session in /status, redirecting to /login.")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	var user models.User
	if err := db.DB.First(&user, sess.UserID).Error; err != nil {
		log.Printf("[STATUS] DB error fetching user_id=%d: %v", sess.UserID, err)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

    // TODO: user id instead of name
	var requests []models.PrintRequest
	if err := db.DB.
		Where("name = ?", user.Name).
		Order("created_at DESC").
		Find(&requests).Error; err != nil {
		log.Printf("[STATUS] DB error fetching requests for user name=%q: %v", user.Name, err)
		http.Error(w, "Failed to load status", http.StatusInternalServerError)
		return
	}

	data := map[string]any{
		"User":     user,
		"Requests": requests,
	}

	if err := RenderTemplateWithData(w,
		data,
		"internal/http/templates/layout.html",
		"internal/http/templates/status.html",
	); err != nil {
		log.Printf("[STATUS] Template error on /status: %v", err)
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}
}
