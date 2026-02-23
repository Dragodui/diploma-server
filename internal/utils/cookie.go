package utils

import (
	"net/http"
)

const (
	// AuthCookieName is the name of the authentication cookie
	AuthCookieName = "auth_token"

	// AuthCookieMaxAge is the maximum age of the auth cookie (24 hours)
	AuthCookieMaxAge = 24 * 60 * 60 // 24 hours in seconds
)

// SetAuthCookie sets a secure HTTP-only cookie with the JWT token
// This prevents token exposure through:
// - Browser history (tokens not in URL)
// - Server logs (tokens not in URL)
// - Referrer headers (tokens not in URL)
// - XSS attacks (HttpOnly flag prevents JavaScript access)
func SetAuthCookie(w http.ResponseWriter, token string, secure bool) {
	cookie := &http.Cookie{
		Name:     AuthCookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   AuthCookieMaxAge,
		HttpOnly: true,  // Prevents JavaScript access (XSS protection)
		Secure:   secure, // Only send over HTTPS (set to true in production)
		SameSite: http.SameSiteLaxMode, // CSRF protection
	}

	http.SetCookie(w, cookie)
}

// ClearAuthCookie removes the authentication cookie
func ClearAuthCookie(w http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:     AuthCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1, // Delete cookie immediately
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}

	http.SetCookie(w, cookie)
}

// GetAuthCookie retrieves the JWT token from the authentication cookie
func GetAuthCookie(r *http.Request) (string, error) {
	cookie, err := r.Cookie(AuthCookieName)
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}
