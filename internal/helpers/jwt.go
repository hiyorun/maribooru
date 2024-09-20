package helpers

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type JWTUser struct {
	Name            string      `json:"name"`
	ID              uuid.UUID   `json:"id"`
	UpdatedSecurity interface{} `json:"updated_security"`
	jwt.RegisteredClaims
}

func GenerateJWT(uid uuid.UUID, name string, secret []byte) (string, error) {
	claims := JWTUser{
		Name: name,
		ID:   uid,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}
