package utils

import (
	"os"
	"time"

	"github.com/tieubaoca/chatbot-be/types"

	"github.com/golang-jwt/jwt/v5"
)

type UserClaims struct {
	ID              string `json:"id"`
	Username        string `json:"username"`
	FullName        string `json:"full_name"`
	ManagementLevel int    `json:"management_level"`
	Role            string `json:"role"`
	WorkspaceRole   string `json:"workspace_role"`
	Workspace       string `json:"workspace"`
	jwt.RegisteredClaims
}
type AdminClaims struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

func GenerateUserToken(user *types.User) (string, error) {
	// Get secret from environment variable
	secret := os.Getenv("JWT_SECRET_USER")
	if secret == "" {
		secret = "default_secret" // Fallback secret, should be changed in production
	}

	// Create claims
	claims := UserClaims{
		ID:              user.ID,
		Username:        user.Username,
		FullName:        user.FullName,
		ManagementLevel: user.ManagementLevel,
		WorkspaceRole:   user.WorkspaceRole,
		Workspace:       user.Workspace,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)), // Token expires in 24 hours
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Subject:   user.ID,
		},
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign and get the complete encoded token as a string
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func GenerateAdminToken(admin *types.Admin) (string, error) {
	// Get secret from environment variable
	secret := os.Getenv("JWT_SECRET_ADMIN")
	if secret == "" {
		secret = "default_admin_secret" // Fallback secret, should be changed in production
	}
	claims := AdminClaims{
		ID:   admin.ID,
		Role: admin.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)), // Token expires in 24 hours
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Subject:   admin.ID,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func ParseUserToken(tokenString string) (*UserClaims, error) {
	// Get secret from environment variable
	secret := os.Getenv("JWT_SECRET_USER")
	if secret == "" {
		secret = "default_secret"
	}
	// Parse token
	token, err := jwt.ParseWithClaims(tokenString, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*UserClaims)
	if !ok || !token.Valid {
		return nil, jwt.ErrInvalidKey
	}
	return claims, nil
}

func ParseAdminToken(tokenString string) (*AdminClaims, error) {
	// Get secret from environment variable
	secret := os.Getenv("JWT_SECRET_ADMIN")
	if secret == "" {
		secret = "default_admin_secret"
	}
	token, err := jwt.ParseWithClaims(tokenString, &AdminClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*AdminClaims)
	if !ok || !token.Valid {
		return nil, jwt.ErrInvalidKey
	}
	return claims, nil
}

func GetIdWithoutCheck(tokenString string) (string, error) {
	claims := jwt.RegisteredClaims{}
	_, _, err := new(jwt.Parser).ParseUnverified(tokenString, &claims)
	if err != nil {
		return "", err
	}
	return claims.Subject, nil
}
