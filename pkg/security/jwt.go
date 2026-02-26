package security

import (
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type MyClaims struct {
	UserID int `json:"uid"`
	jwt.RegisteredClaims
}

func GenerateToken(userID int, subject string, secret []byte, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := MyClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
		Subject:   subject,
		ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		IssuedAt:  jwt.NewNumericDate(now),
		NotBefore: jwt.NewNumericDate(now),
	},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

func ParseToken(tokenStr string, secret []byte) (*MyClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &MyClaims{}, func(t *jwt.Token) (interface{}, error) {
		return secret, nil
	})

	if err != nil || !token.Valid {
		return nil, err
	}

	return token.Claims.(*MyClaims), nil
}
