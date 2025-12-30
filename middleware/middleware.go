package middleware

import (
	"errors"
	"net/http"
	"slices"

	"github.com/gin-gonic/gin"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

func TransactionMiddleware(DB *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		tx := DB.Begin()
		SetContextTransaction(c, tx)

		c.Next()

		if len(c.Errors) == 0 {
			tx.Commit()
		} else {
			tx.Rollback()
		}
	}
}

func TranslationMiddleware(trans ut.Translator) gin.HandlerFunc {
	return func(c *gin.Context) {
		SetContextTranslation(c, trans)

		c.Next()
	}
}

func ErrorMiddleware(c *gin.Context) {
	trans := GetContextTranslation(c)

	c.Next()

	if len(c.Errors) == 0 {
		return
	}

	err := c.Errors.Last()

	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.Errors = slices.Delete(c.Errors, len(c.Errors)-1, len(c.Errors))
		c.JSON(http.StatusNotFound, NewNotFoundError())
		return
	}

	var verr validator.ValidationErrors
	if errors.As(err, &verr) {
		c.Errors = slices.Delete(c.Errors, len(c.Errors)-1, len(c.Errors))
		c.JSON(http.StatusBadRequest, NewValidationError(verr, trans))
		return
	}

	var httpError *HttpError
	if errors.As(err, &httpError) {
		c.JSON(httpError.Code, httpError)
		return
	}

	c.JSON(http.StatusInternalServerError, map[string]any{
		"error": c.Errors.Last().Err,
	})
}
