package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/tieubaoca/chatbot-be/types"
	"github.com/tieubaoca/chatbot-be/utils"
)

type JsonResponse struct {
	Error string `json:"error"`
}

type contextKey string

const (
	userContextKey  contextKey = "user"
	adminContextKey contextKey = "admin"
)

func AuthMiddleware(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, types.DataResponse{
			Status:  false,
			Message: "Authorization header is required",
		})
		c.Abort()
		return
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		c.JSON(http.StatusUnauthorized, types.DataResponse{
			Status:  false,
			Message: "Authorization header format must be Bearer {token}",
		})
		c.Abort()
		return
	}

	claims, err := utils.ParseUserToken(parts[1])
	if err != nil {
		c.JSON(http.StatusUnauthorized, types.DataResponse{
			Status:  false,
			Message: "Invalid user token",
		})
		c.Abort()
		return
	}
	ctx := context.WithValue(c.Request.Context(), userContextKey, claims)
	c.Request = c.Request.WithContext(ctx)
	c.Next()
}

func AdminAuthMiddleware(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, types.DataResponse{
			Status:  false,
			Message: "Authorization header is required",
		})
		c.Abort()
		return
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		c.JSON(http.StatusUnauthorized, types.DataResponse{
			Status:  false,
			Message: "Authorization header format must be Bearer {token}",
		})
		c.Abort()
		return
	}

	claims, err := utils.ParseAdminToken(parts[1])
	if err != nil {
		c.JSON(http.StatusUnauthorized, types.DataResponse{
			Status:  false,
			Message: "Invalid admin token",
		})
		c.Abort()
		return
	}
	ctx := context.WithValue(c.Request.Context(), adminContextKey, claims)
	c.Request = c.Request.WithContext(ctx)
	c.Next()

}
