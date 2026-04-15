package util

import (
	"math/rand"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// 定义JWT密钥
var jwtSecret = []byte("your-secret-key")

// Claims JWT声明结构
type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// GenerateSession 生成32位小写字母的session
func GenerateSession() string {
	const charset = "abcdefghijklmnopqrstuvwxyz"
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
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecret)
	return tokenString, err
}

// ParseJWT 解析JWT token
func ParseJWT(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, jwt.ErrSignatureInvalid
	}

	return claims, nil
}