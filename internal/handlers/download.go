package handlers

import (
	"log"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/arhcodeclub/arh3d/internal/auth"
	"github.com/arhcodeclub/arh3d/internal/db"
	"github.com/arhcodeclub/arh3d/internal/models"
	"gorm.io/gorm"
)

func AdminDownloadFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	sess, err := auth.GetSession(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	var user models.User
	if err := db.DB.First(&user, sess.UserID).Error; err != nil || user.Role != "admin" {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, "Missing id", http.StatusBadRequest)
		return
	}
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil || id == 0 {
		http.Error(w, "Invalid id", http.StatusBadRequest)
		return
	}

	var req models.PrintRequest
	if err := db.DB.First(&req, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "Request not found", http.StatusNotFound)
			return
		}
		log.Printf("[ADMIN] Error loading request for download: %v", err)
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	if req.FilePath == "" {
		http.Error(w, "No file for this request", http.StatusNotFound)
		return
	}

	filename := filepath.Base(req.FilePath)
	w.Header().Set("Content-Disposition", "attachment; filename=\""+filename+"\"")
	http.ServeFile(w, r, req.FilePath)
}
