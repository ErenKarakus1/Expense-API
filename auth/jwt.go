package auth

import (
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
)

func GenerateToken(userid uuid.UUID, jwt_secret string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userid.String(),
		"exp":     time.Now().Add(time.Hour * 8).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ss, err := token.SignedString([]byte(jwt_secret))
	if err != nil {
		return "", err
	}
	return ss, nil
}
