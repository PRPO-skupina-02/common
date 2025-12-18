package request

import (
	"github.com/PRPO-skupina-02/common/validation"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UUIDParam struct {
	UUID string `binding:"required,uuid" json:"uuid"`
}

func GetUUIDParam(c *gin.Context, key string) (uuid.UUID, error) {
	param := c.Param(key)
	uuidParam := UUIDParam{
		UUID: param,
	}
	validator, err := validation.GetDefaultValidationEngine()
	if err != nil {
		return uuid.Max, err
	}

	err = validator.Struct(uuidParam)
	if err != nil {
		return uuid.Max, err
	}

	return uuid.MustParse(uuidParam.UUID), nil
}
