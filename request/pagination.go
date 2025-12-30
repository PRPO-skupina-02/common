package request

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type PaginatedResponse struct {
	Data any `json:"data"`

	Offset int `json:"offset"`
	Limit  int `json:"limit"`
	Total  int `json:"total"`
}

type PaginationOptions struct {
	Offset int
	Limit  int
}

func PaginateScope(pagination *PaginationOptions) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if pagination == nil {
			return db
		}
		return db.Offset(pagination.Offset).Limit(pagination.Limit)
	}
}

func GetNormalizedPaginationArgs(c *gin.Context) *PaginationOptions {
	offset := GetIntQueryParam(c, OffsetQueryKey)
	if offset < 0 {
		offset = 0
	}

	limit := GetIntQueryParam(c, LimitQueryKey)
	switch {
	case limit > 100:
		limit = 100
	case limit <= 0:
		limit = 10
	}

	return &PaginationOptions{
		Offset: offset,
		Limit:  limit,
	}
}

func RenderPaginatedResponse(c *gin.Context, data any, total int) {
	pagination := GetNormalizedPaginationArgs(c)

	resp := PaginatedResponse{
		Data:   data,
		Limit:  pagination.Limit,
		Offset: pagination.Offset,
		Total:  total,
	}

	c.JSON(http.StatusOK, resp)
}

type SortOptions struct {
	Column string
	Desc   bool
}

func SortScope(sort *SortOptions) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if sort == nil {
			return db
		}
		return db.Order(clause.OrderByColumn{
			Column: clause.Column{
				Name: sort.Column,
			},
			Desc: sort.Desc,
		})
	}
}

func GetSortOptions(c *gin.Context) *SortOptions {
	sort := c.Query("sort")

	if sort == "" {
		return nil
	}

	desc := false

	if strings.HasPrefix(sort, "-") {
		desc = true
	}

	return &SortOptions{
		Column: strings.ToLower(strings.TrimPrefix(sort, "-")),
		Desc:   desc,
	}
}
