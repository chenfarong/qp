package util

import (
	"math/rand"
	"time"

	"zgame/config"

	"github.com/golang-jwt/jwt/v5"
)

// Claims JWT声明结构
type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// GenerateSession 生成32位小写字母和数字的session
func GenerateSession() string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	sb := make([]byte, 32)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := range sb {
		sb[i] = charset[r.Intn(len(charset))]
	}
	return string(sb)
}

// GenerateJWT 生成JWT token
func GenerateJWT(username string) (string, error) {
	claims := Claims{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(config.AppConfig.Auth.TokenExpiry) * time.Second)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(config.AppConfig.Auth.SecretKey))
	return tokenString, err
}

// ParseJWT 解析JWT token
func ParseJWT(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.AppConfig.Auth.SecretKey), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, jwt.ErrSignatureInvalid
	}

	return claims, nil
}