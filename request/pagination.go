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

func PaginateScope(offset, limit int) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Offset(offset).Limit(limit)
	}
}

func GetNormalizedPaginationArgs(c *gin.Context) (int, int) {
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

	return offset, limit
}

func RenderPaginatedResponse(c *gin.Context, data any, total int) {
	limit, offset := GetNormalizedPaginationArgs(c)

	resp := PaginatedResponse{
		Data:   data,
		Limit:  limit,
		Offset: offset,
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
