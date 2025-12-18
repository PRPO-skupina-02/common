package middleware

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	ut "github.com/go-playground/universal-translator"
	"gorm.io/gorm"
)

const (
	contextTransactionKey = "transaction"
	contextTranslationKey = "translation"
)

func SetContextTransaction(c *gin.Context, tx *gorm.DB) {
	c.Set(contextTransactionKey, tx)
}

func GetContextTransaction(c *gin.Context) *gorm.DB {
	tx, ok := c.Get(contextTransactionKey)
	if !ok {
		slog.Error("Could not get transaction from context")
		return nil
	}

	return tx.(*gorm.DB)
}

func SetContextTranslation(c *gin.Context, trans ut.Translator) {
	c.Set(contextTranslationKey, trans)
}

func GetContextTranslation(c *gin.Context) ut.Translator {
	trans, ok := c.Get(contextTranslationKey)
	if !ok {
		slog.Error("Could not get translation from context")
		return nil
	}

	return trans.(ut.Translator)
}
