package handlers

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/arhcodeclub/arh3d/internal/auth"
	"github.com/arhcodeclub/arh3d/internal/db"
	"github.com/arhcodeclub/arh3d/internal/models"
	"gorm.io/gorm"
)

func AdminHandler(w http.ResponseWriter, r *http.Request) {
	sess, err := auth.GetSession(r)
	if err != nil {
		log.Printf("[ADMIN] No session, redirecting to /login.")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	var user models.User
	if err := db.DB.First(&user, sess.UserID).Error; err != nil {
		log.Printf("[ADMIN] DB error fetching user_id=%d: %v", sess.UserID, err)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if user.Role != "admin" {
		log.Printf("[ADMIN] Access denied for user_id=%d, role=%s.", user.ID, user.Role)
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	switch {
        case r.Method == http.MethodGet && r.URL.Path == "/admin":
            adminListPage(w, r, &user)
        case r.Method == http.MethodPost && r.URL.Path == "/admin/update":
            adminUpdateRequest(w, r, &user)
        default:
            http.Error(w, "Not found", http.StatusNotFound)
	}
}

func adminListPage(w http.ResponseWriter, r *http.Request, admin *models.User) {
	log.Printf("[ADMIN] GET /admin by user_id=%d.", admin.ID)

	statusFilter := strings.TrimSpace(r.URL.Query().Get("status"))

	var requests []models.PrintRequest
	query := db.DB.Order("created_at DESC")
	if statusFilter != "" {
		query = query.Where("status = ?", statusFilter)
	}

	if err := query.Find(&requests).Error; err != nil {
		log.Printf("[ADMIN] DB error fetching all requests: %v", err)
		http.Error(w, "Failed to load admin dashboard", http.StatusInternalServerError)
		return
	}

	data := map[string]any{
		"Admin":    admin,
		"Requests": requests,
		"StatusFilter": statusFilter,
	}

	if err := RenderTemplateWithData(w,
		data,
		"internal/http/templates/layout.html",
		"internal/http/templates/admin.html",
	); err != nil {
		log.Printf("[ADMIN] Template error on /admin: %v", err)
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}
}

func adminUpdateRequest(w http.ResponseWriter, r *http.Request, admin *models.User) {
	log.Printf("[ADMIN] POST /admin/update by user_id=%d.", admin.ID)

	if err := r.ParseForm(); err != nil {
		log.Printf("[ADMIN] Error parsing admin update form: %v", err)
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	idStr := r.FormValue("id")
	status := strings.TrimSpace(r.FormValue("status"))
	queuePosStr := strings.TrimSpace(r.FormValue("queue_position"))

	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil || id == 0 {
		log.Printf("[ADMIN] Invalid request ID: %q", idStr)
		http.Error(w, "Invalid request ID", http.StatusBadRequest)
		return
	}

	allowedStatuses := map[string]bool{
		"pending": true,
		"in_queue": true,
		"printing": true,
		"finished": true,
		"rejected": true,
		"canceled": true,
		"on_hold": true,
	}
	if status != "" && !allowedStatuses[status] {
		log.Printf("[ADMIN] Invalid status %q for request id=%d", status, id)
		http.Error(w, "Invalid status", http.StatusBadRequest)
		return
	}

	var req models.PrintRequest
	if err := db.DB.First(&req, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Printf("[ADMIN] PrintRequest id=%d not found.", id)
			http.Error(w, "Request not found", http.StatusNotFound)
			return
		}
		log.Printf("[ADMIN] DB error fetching PrintRequest id=%d: %v", id, err)
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	if status != "" {
		req.Status = status
	}

	if queuePosStr != "" {
		if qp, err := strconv.Atoi(queuePosStr); err == nil && qp >= 0 {
			req.QueuePosition = qp
		} else {
			log.Printf("[ADMIN] Invalid queue position %q for request id=%d", queuePosStr, id)
		}
	}

	if err := db.DB.Save(&req).Error; err != nil {
		log.Printf("[ADMIN] Error saving updated PrintRequest id=%d: %v", req.ID, err)
		http.Error(w, "Failed to update request", http.StatusInternalServerError)
		return
	}

	log.Printf("[ADMIN] Updated PrintRequest id=%d: status=%s queue_pos=%d", req.ID, req.Status, req.QueuePosition)

	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}
