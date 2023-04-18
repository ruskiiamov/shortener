package server

import (
	"errors"
	"net/http"

	"github.com/ruskiiamov/shortener/internal/user"
)

const (
	authCookieName   = "auth"
	userIDCookieName = "user_id"
)

type authMiddleware struct {
	ua user.Authorizer
}

func newAuthMiddleware(ua user.Authorizer) *authMiddleware {
	return &authMiddleware{
		ua: ua,
	}
}

func (a *authMiddleware) handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(authCookieName)
		if err != nil && !errors.Is(err, http.ErrNoCookie) {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var userID string

		if cookie != nil {
			userID, err = a.ua.GetUserID(cookie.Value)
		} else {
			err = errors.New("empty cookie")
		}

		if err != nil {
			var token string
			userID, token, err = a.ua.CreateUser()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			http.SetCookie(w, &http.Cookie{
				Name:  authCookieName,
				Value: token,
				Path:  "/",
			})
		}

		r.AddCookie(&http.Cookie{
			Name:  userIDCookieName,
			Value: userID,
		})

		next.ServeHTTP(w, r)
	})

}
