package handlers

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/arhcodeclub/arh3d/internal/auth"
	"github.com/arhcodeclub/arh3d/internal/db"
	"github.com/arhcodeclub/arh3d/internal/models"
	"gorm.io/gorm"
)

func LoginPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	log.Printf("[AUTH] GET /login from %s.", r.RemoteAddr)

	if err := RenderTemplate(w,
		"internal/http/templates/layout.html",
		"internal/http/templates/login.html",
	); err != nil {
		log.Printf("[AUTH] Template error on /login: %v", err)
		http.Error(w, "Template error", http.StatusInternalServerError)
	}
}

func LoginRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		log.Printf("[AUTH] Error parsing login form: %v.", err)
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	email := strings.TrimSpace(strings.ToLower(r.FormValue("email")))
	log.Printf("[AUTH] Login request for email=%s.", email)

	if !strings.HasSuffix(email, "@gsf.nl") {
		log.Printf("[AUTH] Rejected login for email=%s.", email)
		http.Error(w, "You must use your @gsf.nl email address.", http.StatusBadRequest)
		return
	}

	rawToken, err := auth.RandomToken(32)
	if err != nil {
		log.Printf("[AUTH] Error generating login token: %v", err)
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	h := sha256.Sum256([]byte(rawToken))
	tokenHash := hex.EncodeToString(h[:])

	token := models.LoginToken{
		Email:     email,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(10 * time.Minute),
	}

	if err := db.DB.Create(&token).Error; err != nil {
		log.Printf("[AUTH] Error saving login token: %v", err)
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	log.Printf("[AUTH] Created login token id=%d for email=%s. (expires_at=%s)", token.ID, token.Email, token.ExpiresAt)

	baseURL := os.Getenv("APP_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	verifyURL := fmt.Sprintf("%s/login/verify?token=%s", strings.TrimRight(baseURL, "/"), url.QueryEscape(rawToken))
	log.Printf("[AUTH] Verify URL for %s: %s.", email, verifyURL)

	if err := auth.SendMagicLinkEmail(email, verifyURL); err != nil {
		log.Printf("[AUTH] Error sending magic-link email to %s: %v", email, err)
		http.Error(w, "Failed to send email", http.StatusInternalServerError)
		return
	}

	if err := RenderTemplateWithData(w,
		map[string]any{"Email": email},
		"internal/http/templates/layout.html",
		"internal/http/templates/login_sent.html",
	); err != nil {
		log.Printf("[AUTH] Template error on login_sent: %v", err)
		http.Error(w, "Template error", http.StatusInternalServerError)
	}
}

func LoginVerify(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	rawToken := r.URL.Query().Get("token")
	if rawToken == "" {
		log.Printf("[AUTH] /login/verify missing token.")
		http.Error(w, "Missing token", http.StatusBadRequest)
		return
	}

	log.Printf("[AUTH] /login/verify with token_len=%d.", len(rawToken))

	h := sha256.Sum256([]byte(rawToken))
	tokenHash := hex.EncodeToString(h[:])

	var t models.LoginToken
	if err := db.DB.Where("token_hash = ?", tokenHash).First(&t).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Printf("[AUTH] Login token not found for hash prefix=%s.", tokenHash[:8])
			http.Error(w, "Invalid or expired link", http.StatusBadRequest)
			return
		}
		log.Printf("[AUTH] DB error fetching login token: %v", err)
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	log.Printf("[AUTH] Found token id=%d for email=%s. (expires_at=%s used_at=%v)",
		t.ID, t.Email, t.ExpiresAt, t.UsedAt)

	if t.UsedAt != nil || time.Now().After(t.ExpiresAt) {
		log.Printf("[AUTH] Token id=%d invalid. (used_at=%v, now=%s, expires_at=%s)",
			t.ID, t.UsedAt, time.Now(), t.ExpiresAt)
		http.Error(w, "Invalid or expired link", http.StatusBadRequest)
		return
	}

	now := time.Now()
	if err := db.DB.Model(&t).Update("used_at", &now).Error; err != nil {
		log.Printf("[AUTH] Error marking token as used: %v", err)
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	email := strings.ToLower(t.Email)
	log.Printf("[AUTH] Verifying user for email=%s.", email)

	var user models.User
	err := db.DB.Where("email = ?", email).First(&user).Error
	isNew := false

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Printf("[AUTH] No user found for email=%s, creating new user.", email)
			user = models.User{
				Email: email,
				Role:  "student",
			}

			if err := db.DB.Create(&user).Error; err != nil {
				log.Printf("[AUTH] Error creating user: %v", err)
				http.Error(w, "Server error", http.StatusInternalServerError)
				return
			}

			isNew = true
			log.Printf("[AUTH] Created new user id=%d with email=%s.", user.ID, user.Email)
		} else {
			log.Printf("[AUTH] DB error fetching user for email=%s: %v", email, err)
			http.Error(w, "Server error", http.StatusInternalServerError)
			return
		}
	} else {
		log.Printf("[AUTH] Found existing user id=%d with email=%s and name=%q.", user.ID, user.Email, user.Name)
	}

	if err := auth.CreateSession(w, user.ID); err != nil {
		log.Printf("[AUTH] Error creating session: %v", err)
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	if isNew || strings.TrimSpace(user.Name) == "" {
		log.Printf("[AUTH] Redirecting user id=%d to /onboarding.", user.ID)
		http.Redirect(w, r, "/onboarding", http.StatusSeeOther)
		return
	}

	log.Printf("[AUTH] Login complete for user id=%d, redirecting to /.", user.ID)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func OnboardingPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	log.Printf("[AUTH] GET /onboarding.")

	sess, err := auth.GetSession(r)
	if err != nil {
		log.Printf("[AUTH] No session in /onboarding.")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	var user models.User
	if err := db.DB.First(&user, sess.UserID).Error; err != nil {
		log.Printf("[AUTH] DB error in /onboarding fetching user_id=%d: %v", sess.UserID, err)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if err := RenderTemplateWithData(w,
		map[string]any{"User": user},
		"internal/http/templates/layout.html",
		"internal/http/templates/onboarding.html",
	); err != nil {
		log.Printf("[AUTH] Template error on /onboarding: %v", err)
		http.Error(w, "Template error", http.StatusInternalServerError)
	}
}

func OnboardingSubmit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	log.Printf("[AUTH] POST /onboarding.")

	sess, err := auth.GetSession(r)
	if err != nil {
		log.Printf("[AUTH] No session in /onboarding POST.")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	var user models.User
	if err := db.DB.First(&user, sess.UserID).Error; err != nil {
		log.Printf("[AUTH] DB error fetching user_id=%d: %v", sess.UserID, err)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if err := r.ParseForm(); err != nil {
		log.Printf("[AUTH] Error parsing onboarding form: %v", err)
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	name := strings.TrimSpace(r.FormValue("name"))
	department := strings.TrimSpace(r.FormValue("department"))

	log.Printf("[AUTH] Onboarding user id=%d set name=%q at department=%q.", user.ID, name, department)

	user.Name = name
	// TODO: store department

	if err := db.DB.Save(&user).Error; err != nil {
		log.Printf("[AUTH] Error saving user profile id=%d: %v", user.ID, err)
		http.Error(w, "Failed to save profile", http.StatusInternalServerError)
		return
	}

	log.Printf("[AUTH] Onboarding complete for user id=%d.", user.ID)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	log.Printf("[AUTH] POST /logout from %s.", r.RemoteAddr)
	auth.DestroySession(w, r)
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
