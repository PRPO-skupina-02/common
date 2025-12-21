package request

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/PRPO-skupina-02/common/middleware"
	"github.com/PRPO-skupina-02/common/validation"
	"github.com/PRPO-skupina-02/common/xtesting"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func UUIDParamHandler(c *gin.Context) {
	id, err := GetUUIDParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, map[string]any{"message": "Param is not a valid UUID", "id": id})
		return
	}
	c.JSON(http.StatusOK, map[string]any{"message": "Param is a valid UUID", "id": id})
}

func TestUUIDParam(t *testing.T) {
	r := gin.Default()
	trans, err := validation.RegisterValidation()
	r.Use(middleware.TranslationMiddleware(trans))

	r.GET("/api/v1/:id", UUIDParamHandler)

	require.NoError(t, err)

	tests := []struct {
		name   string
		status int
		id     string
	}{
		{
			name:   "ok",
			status: http.StatusOK,
			id:     "0fa9dd70-ddd8-11f0-ae2d-ef3f0bc8b056",
		},
		{
			name:   "invalid",
			status: http.StatusBadRequest,
			id:     "123",
		},
		{
			name:   "nil",
			status: http.StatusBadRequest,
			id:     "00000000-0000-0000-0000-000000000000",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			targetURL := fmt.Sprintf("/api/v1/%s", testCase.id)

			req := xtesting.NewTestingRequest(t, targetURL, http.MethodGet, nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			assert.Equal(t, testCase.status, w.Code)
			xtesting.AssertGoldenJSON(t, w)
		})
	}
}
