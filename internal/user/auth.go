// Package users is the shortener service logic for users.
package user

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"

	"github.com/gofrs/uuid"
)

// Authorizer provides the logic for user authentication.
type Authorizer interface {
	CreateUser() (userID, token string, err error)
	GetUserID(token string) (string, error)
}

type authorizer struct {
	key []byte
}

// NewAuthorizer returns Authorizer instance.
func NewAuthorizer(key []byte) Authorizer {
	return &authorizer{
		key: key,
	}
}

// CreateUser returns generated user ID with auth token.
func (a *authorizer) CreateUser() (userID, token string, err error) {
	id, err := uuid.NewV4()
	if err != nil {
		return "", "", err
	}

	userID = id.String()

	hmacHash := hmac.New(sha256.New, a.key)

	_, err = hmacHash.Write([]byte(userID))
	if err != nil {
		return "", "", err
	}

	signature := hmacHash.Sum(nil)
	hmacHash.Reset()

	token = base64.URLEncoding.EncodeToString(append(signature, []byte(userID)...))

	return userID, token, nil
}

// GetUserID returns user ID by auth token.
func (a *authorizer) GetUserID(token string) (string, error) {
	value, err := base64.URLEncoding.DecodeString(token)
	if err != nil {
		return "", err
	}

	if len(value) < sha256.Size {
		return "", errors.New("wrong token")
	}

	userID := value[sha256.Size:]
	expectedSignature := value[:sha256.Size]

	hmacHash := hmac.New(sha256.New, a.key)
	_, err = hmacHash.Write(userID)
	if err != nil {
		return "", err
	}

	signature := hmacHash.Sum(nil)
	hmacHash.Reset()

	if hmac.Equal(signature, expectedSignature) {

		return string(userID), nil
	}

	return "", errors.New("wrong token")
}
