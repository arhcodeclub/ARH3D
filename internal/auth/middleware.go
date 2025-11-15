package auth

import (
    "log"
    "net/http"

	"github.com/arhcodeclub/arh3d/internal/db"
	"github.com/arhcodeclub/arh3d/internal/models"
	"gorm.io/gorm"
)

func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        log.Printf("[AUTH] Authenticating on %s %s.", r.Method, r.URL.Path)

        sess, err := GetSession(r)
		if err != nil {
            log.Printf("[AUTH] No valid session.")
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		var user models.User
		if err := db.DB.First(&user, sess.UserID).Error; err != nil {
            log.Printf("[AUTH] Session has user_id=%d but user not found, destroying session.", sess.UserID)
			if err == gorm.ErrRecordNotFound {
				DestroySession(w, r)
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}

            log.Printf("[AUTH] DB error fetching user_id=%d: %v", sess.UserID, err)
			http.Error(w, "Server error", http.StatusInternalServerError)
			return
		}

        log.Printf("[AUTH] Authenticated user. (id=%d email=%s role=%s)", user.ID, user.Email, user.Role)

		ctx := ContextWithUser(r.Context(), &user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
