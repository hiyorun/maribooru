package helpers

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type JWTUser struct {
	Name            string      `json:"name"`
	ID              uuid.UUID   `json:"id"`
	UpdatedSecurity interface{} `json:"updated_security"`
	jwt.RegisteredClaims
}

func GenerateJWT(uid uuid.UUID, name string, secret []byte, expiry time.Duration) (string, error) {
	claims := JWTUser{
		Name: name,
		ID:   uid,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiry)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

func GetUserID(c echo.Context, secret []byte) (uuid.UUID, error) {
	token := c.Request().Header.Get("Authorization")
	if token == "" {
		return uuid.Nil, errors.New("Token is empty")
	}

	claims := JWTUser{}
	_, err := jwt.ParseWithClaims(token[7:], &claims, func(token *jwt.Token) (interface{}, error) {
		return secret, nil
	})
	if err != nil {
		return uuid.Nil, err
	}

	return claims.ID, nil
}
