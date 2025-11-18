package handlers

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/arhcodeclub/arh3d/internal/auth"
	"github.com/arhcodeclub/arh3d/internal/db"
	"github.com/arhcodeclub/arh3d/internal/models"
)

func NewRequestHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		log.Printf("[REQUEST] GET /new.")

		sess, err := auth.GetSession(r)
		if err != nil {
			log.Printf("[REQUEST] No session in GET /new.")
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		var user models.User
		if err := db.DB.First(&user, sess.UserID).Error; err != nil {
			log.Printf("[REQUEST] DB error fetching user_id=%d for /new: %v", sess.UserID, err)
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		var filaments []models.Filament
		if err := db.DB.Order("type, colour").Find(&filaments).Error; err != nil {
			log.Printf("[REQUEST] Error loading filaments: %v", err)
			filaments = nil
		}

		data := map[string]any{
			"User":      user,
			"Filaments": filaments,
		}

		if err := RenderTemplateWithData(w,
			data,
			"internal/http/templates/layout.html",
			"internal/http/templates/new.html",
		); err != nil {
			log.Printf("[REQUEST] Template error on /new: %v", err)
			http.Error(w, "Template error", 500)
			return
		}

	case http.MethodPost:
		log.Printf("[REQUEST] POST /new.")

		if err := r.ParseMultipartForm(50 << 20); err != nil {
			log.Printf("[REQUEST] Error parsing form: %v", err)
			http.Error(w, "Failed to parse form", 400)
			return
		}

		sess, err := auth.GetSession(r)
		if err != nil {
			log.Printf("[REQUEST] No session in POST /new.")
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		var user models.User
		if err := db.DB.First(&user, sess.UserID).Error; err != nil {
			log.Printf("[REQUEST] DB error fetching user_id=%d: %v", sess.UserID, err)
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		name := user.Name
		inputType := r.FormValue("input_type")
		link := r.FormValue("link")
		description := r.FormValue("description")
		colour := r.FormValue("colour")
		comments := r.FormValue("comments")
		filePath := ""

		log.Printf("[REQUEST] User id=%d submitting request: input_type=%s colour=%s.", user.ID, inputType, colour)

		if inputType == "file" {
			file, header, err := r.FormFile("file")
			if err == nil {
				defer file.Close()
				os.MkdirAll("uploads", os.ModePerm)
				filePath = filepath.Join("uploads", header.Filename)
				dst, err := os.Create(filePath)
				if err != nil {
					log.Printf("[REQUEST] Error creating upload file: %v", err)
					http.Error(w, "Failed to save file", 500)
					return
				}
				defer dst.Close()
				if _, err = dst.ReadFrom(file); err != nil {
					log.Printf("[REQUEST] Error writing upload file: %v", err)
					http.Error(w, "Failed to save file", 500)
					return
				}
				log.Printf("[REQUEST] Saved upload to %s.", filePath)
			} else {
				log.Printf("[REQUEST] Error reading uploaded file: %v", err)
			}
		}

		var queueCount int64
		if err := db.DB.Model(&models.PrintRequest{}).
			Where("status = ?", "in_queue").
			Count(&queueCount).Error; err != nil {
			log.Printf("[REQUEST] Error counting in-queue requests: %v", err)
		}

		request := models.PrintRequest{
			UserID:        user.ID,
			Name:          name,
			InputType:     inputType,
			FilePath:      filePath,
			Link:          link,
			Description:   description,
			Colour:        colour,
			Comments:      comments,
			Status:        "pending",
			QueuePosition: int(queueCount) + 1,
		}

		if err := db.DB.Create(&request).Error; err != nil {
			log.Printf("[REQUEST] Error saving PrintRequest: %v", err)
			http.Error(w, "Failed to save request", 500)
			return
		}

		log.Printf("[REQUEST] Saved PrintRequest id=%d for user_id=%d with status=%s queue_pos=%d.",
			request.ID, request.UserID, request.Status, request.QueuePosition)

		http.Redirect(w, r, "/status", http.StatusSeeOther)

	default:
		http.Error(w, "Method not allowed", 405)
	}
}
