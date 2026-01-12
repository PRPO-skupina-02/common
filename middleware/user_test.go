package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/PRPO-skupina-02/common/clients/auth/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestSetAndGetContextUser(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	userID := uuid.New()
	user := &models.APIUserResponse{
		ID:        userID.String(),
		Email:     "test@example.com",
		FirstName: "Test",
		LastName:  "User",
		Role:      models.ModelsUserRoleCustomer,
		Active:    true,
	}

	SetContextUser(c, user)
	retrievedUser := GetContextUser(c)

	assert.NotNil(t, retrievedUser)
	assert.Equal(t, user.Email, retrievedUser.Email)
	assert.Equal(t, user.ID, retrievedUser.ID)
	assert.Equal(t, user.Role, retrievedUser.Role)
}

func TestGetContextUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	userID := uuid.New()
	user := &models.APIUserResponse{
		ID:        userID.String(),
		Email:     "test@example.com",
		FirstName: "Test",
		LastName:  "User",
		Role:      models.ModelsUserRoleCustomer,
		Active:    true,
	}

	SetContextUser(c, user)
	retrievedUserID := GetContextUserID(c)

	assert.Equal(t, userID, retrievedUserID)
}

func TestGetContextUserRole(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	user := &models.APIUserResponse{
		ID:        uuid.New().String(),
		Email:     "admin@example.com",
		FirstName: "Admin",
		LastName:  "User",
		Role:      models.ModelsUserRoleAdmin,
		Active:    true,
	}

	SetContextUser(c, user)
	retrievedRole := GetContextUserRole(c)

	assert.Equal(t, models.ModelsUserRoleAdmin, retrievedRole)
}

func TestRequireRole(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		userRole       models.ModelsUserRole
		requiredRoles  []models.ModelsUserRole
		expectedStatus int
	}{
		{
			name:           "admin accessing admin endpoint",
			userRole:       models.ModelsUserRoleAdmin,
			requiredRoles:  []models.ModelsUserRole{models.ModelsUserRoleAdmin},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "employee accessing admin endpoint",
			userRole:       models.ModelsUserRoleEmployee,
			requiredRoles:  []models.ModelsUserRole{models.ModelsUserRoleAdmin},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "admin accessing employee or admin endpoint",
			userRole:       models.ModelsUserRoleAdmin,
			requiredRoles:  []models.ModelsUserRole{models.ModelsUserRoleAdmin, models.ModelsUserRoleEmployee},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "customer accessing employee or admin endpoint",
			userRole:       models.ModelsUserRoleCustomer,
			requiredRoles:  []models.ModelsUserRole{models.ModelsUserRoleAdmin, models.ModelsUserRoleEmployee},
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, router := gin.CreateTestContext(w)

			user := &models.APIUserResponse{
				ID:        uuid.New().String(),
				Email:     "test@example.com",
				FirstName: "Test",
				LastName:  "User",
				Role:      tt.userRole,
				Active:    true,
			}

			router.GET("/test",
				func(c *gin.Context) {
					SetContextUser(c, user)
					c.Next()
				},
				RequireRole(tt.requiredRoles...),
				func(c *gin.Context) {
					c.JSON(http.StatusOK, gin.H{"message": "success"})
				},
			)

			c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
			router.ServeHTTP(w, c.Request)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestRequireAdmin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		userRole       models.ModelsUserRole
		expectedStatus int
	}{
		{
			name:           "admin accessing admin endpoint",
			userRole:       models.ModelsUserRoleAdmin,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "employee accessing admin endpoint",
			userRole:       models.ModelsUserRoleEmployee,
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "customer accessing admin endpoint",
			userRole:       models.ModelsUserRoleCustomer,
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, router := gin.CreateTestContext(w)

			user := &models.APIUserResponse{
				ID:        uuid.New().String(),
				Email:     "test@example.com",
				FirstName: "Test",
				LastName:  "User",
				Role:      tt.userRole,
				Active:    true,
			}

			router.GET("/admin",
				func(c *gin.Context) {
					SetContextUser(c, user)
					c.Next()
				},
				RequireAdmin(),
				func(c *gin.Context) {
					c.JSON(http.StatusOK, gin.H{"message": "success"})
				},
			)

			c.Request = httptest.NewRequest(http.MethodGet, "/admin", nil)
			router.ServeHTTP(w, c.Request)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}
