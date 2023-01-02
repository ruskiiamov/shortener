package server

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"hash"
	"net/http"

	"github.com/gofrs/uuid"
)

const (
	authCookieName   = "auth"
	userIDCookieName = "user_id"
)

var h hash.Hash

func initAuth(key string) {
	h = hmac.New(sha256.New, []byte(key))
}

func auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, err := getOrCreateUserID(w, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		r.AddCookie(&http.Cookie{
			Name:  userIDCookieName,
			Value: userID,
		})

		next.ServeHTTP(w, r)
	})
}

func getOrCreateUserID(w http.ResponseWriter, r *http.Request) (string, error) {
	cookie, err := r.Cookie(authCookieName)
	if err != nil && !errors.Is(err, http.ErrNoCookie) {
		return "", err
	}

	userID, ok := getValidUserID(cookie)
	if !ok {
		userID, err = createUserID(w)
		if err != nil {
			return "", nil
		}
	}

	return userID, nil
}

func getValidUserID(c *http.Cookie) (string, bool) {
	if c == nil {
		return "", false
	}

	value, err := base64.URLEncoding.DecodeString(c.Value)
	if err != nil {
		return "", false
	}

	if len(value) < sha256.Size {
		return "", false
	}

	userID := value[sha256.Size:]
	expectedSignature := value[:sha256.Size]

	_, err = h.Write(userID)
	if err != nil {
		return "", false
	}
	signature := h.Sum(nil)
	h.Reset()

	if hmac.Equal(signature, expectedSignature) {
		return string(userID), true
	}

	return "", false
}

func createUserID(w http.ResponseWriter) (string, error) {
	id, err := uuid.NewV4()
	if err != nil {
		return "", err
	}

	userID := id.String()

	_, err = h.Write([]byte(userID))
	if err != nil {
		return "", err
	}
	signature := h.Sum(nil)
	h.Reset()

	value := string(signature) + userID
	valueStr := base64.URLEncoding.EncodeToString([]byte(value))

	http.SetCookie(w, &http.Cookie{
		Name:  authCookieName,
		Value: valueStr,
	})

	return userID, nil
}
