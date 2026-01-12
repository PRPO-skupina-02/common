package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/PRPO-skupina-02/common/clients/auth/client"
	"github.com/PRPO-skupina-02/common/clients/auth/client/auth"
	"github.com/PRPO-skupina-02/common/clients/auth/models"
	"github.com/gin-gonic/gin"
	"github.com/go-openapi/strfmt"
	"github.com/google/uuid"
)

const (
	contextUserKey = "user"
)

func UserMiddleware(authHost string) gin.HandlerFunc {
	transportConfig := client.DefaultTransportConfig().WithHost(authHost)
	authClient := client.NewHTTPClientWithConfig(strfmt.Default, transportConfig)

	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, NewUnauthorizedError("Authorization header required"))
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, NewUnauthorizedError("Invalid authorization header format"))
			return
		}

		token := parts[1]

		params := auth.NewVerifyTokenParams()
		params.Token = auth.VerifyTokenBody{
			Token: token,
		}

		result, err := authClient.Auth.VerifyToken(params)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, NewUnauthorizedError("Invalid or expired token"))
			return
		}

		SetContextUser(c, result.Payload)

		c.Next()
	}
}

func SetContextUser(c *gin.Context, user *models.APIUserResponse) {
	c.Set(contextUserKey, user)
}

func GetContextUser(c *gin.Context) *models.APIUserResponse {
	user, ok := c.Get(contextUserKey)
	if !ok {
		_ = c.AbortWithError(http.StatusInternalServerError, errors.New("Could not get user from context"))
		return nil
	}

	return user.(*models.APIUserResponse)
}

func GetContextUserID(c *gin.Context) uuid.UUID {
	user := GetContextUser(c)
	if user == nil {
		return uuid.Nil
	}

	userID, err := uuid.Parse(user.ID)
	if err != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, errors.New("Invalid user ID"))
		return uuid.Nil
	}

	return userID
}

func GetContextUserRole(c *gin.Context) models.ModelsUserRole {
	user := GetContextUser(c)
	if user == nil {
		return ""
	}

	return user.Role
}

func RequireRole(roles ...models.ModelsUserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole := GetContextUserRole(c)
		if userRole == "" {
			c.AbortWithStatusJSON(http.StatusForbidden, NewForbiddenError("Role information not found"))
			return
		}

		// Check if user has one of the required roles
		for _, requiredRole := range roles {
			if userRole == requiredRole {
				c.Next()
				return
			}
		}

		c.AbortWithStatusJSON(http.StatusForbidden, NewForbiddenError("Insufficient permissions"))
	}
}

func RequireAdmin() gin.HandlerFunc {
	return RequireRole(models.ModelsUserRoleAdmin)
}

func RequireEmployee() gin.HandlerFunc {
	return RequireRole(models.ModelsUserRoleEmployee)
}
