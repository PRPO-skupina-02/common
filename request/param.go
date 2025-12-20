package request

import (
	"strconv"

	"github.com/PRPO-skupina-02/common/validation"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	LimitQueryKey  = "limit"
	OffsetQueryKey = "offset"
)

type UUIDParam struct {
	UUID string `binding:"required,uuid,non-nil-uuid" json:"uuid"`
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

func GetIntQueryParam(c *gin.Context, key string) int {
	return GetIntQueryParamWithDefault(c, key, 0)
}

func GetIntQueryParamWithDefault(c *gin.Context, key string, fallback int) int {
	param, err := strconv.Atoi(c.Query(key))
	if err != nil {
		return fallback
	}
	return param
}
