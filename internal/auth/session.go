package auth

import (
	"errors"
    "log"
    "net/http"
	"sync"
	"time"
)

const (
	sessionCookieName = "arh3d_session"
	sessionTTL = 7 * 24 * time.Hour
)

type Session struct {
	UserID uint
	ExpiresAt time.Time
}

var (
	sessions = make(map[string]*Session)
	sessionsMu sync.RWMutex
)

func CreateSession(w http.ResponseWriter, userID uint) error {
	token, err := RandomToken(32)
	if err != nil {
		return err
	}

	sessionsMu.Lock()
	sessions[token] = &Session{
		UserID:    userID,
		ExpiresAt: time.Now().Add(sessionTTL),
	}
	sessionsMu.Unlock()

    log.Printf("[SESSION] Created session for user_id=%d. (len=%d)", userID, len(token))

    http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
        Secure:   false, // TODO: set to true in prod
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(sessionTTL),
	})

	return nil
}

func GetSession(r *http.Request) (*Session, error) {
	c, err := r.Cookie(sessionCookieName)
	if err != nil {
        log.Printf("[SESSION] No session cookie on request %s %s.", r.Method, r.URL.Path)
		return nil, err
	}

	sessionsMu.RLock()
	sess, ok := sessions[c.Value]
	sessionsMu.RUnlock()
	if !ok {
        log.Printf("[SESSION] Session not found for token.")
		return nil, errors.New("session not found")
	}

	if time.Now().After(sess.ExpiresAt) {
        log.Printf("[SESSION] Session expired for user_id=%d.", sess.UserID)
		sessionsMu.Lock()
		delete(sessions, c.Value)
		sessionsMu.Unlock()
		return nil, errors.New("session expired")
	}

	log.Printf("[SESSION] Found valid session for user_id=%d.", sess.UserID)
    return sess, nil
}

func DestroySession(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie(sessionCookieName)
	if err == nil {
		sessionsMu.Lock()
		delete(sessions, c.Value)
		sessionsMu.Unlock()
        log.Printf("[SESSION] Destroyed session.")
	}

	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	})
}
