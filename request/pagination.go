package request

import (
	"fmt"
	"net/http"
	"strings"
	"time"

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

type Filter interface {
	Apply(db *gorm.DB) *gorm.DB
}

type FilterOptions struct {
	Filters []Filter
}

func FilterScope(opts *FilterOptions) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if opts == nil || len(opts.Filters) == 0 {
			return db
		}

		query := db
		for _, filter := range opts.Filters {
			if filter == nil {
				continue
			}
			query = filter.Apply(query)
		}
		return query
	}
}

func NewFilterOptions(filters ...Filter) *FilterOptions {
	return &FilterOptions{Filters: filters}
}

func (fo *FilterOptions) AddFilter(f Filter) *FilterOptions {
	fo.Filters = append(fo.Filters, f)
	return fo
}

type DateRangeFilter struct {
	Column string
	Date   time.Time
}

func (f DateRangeFilter) Apply(db *gorm.DB) *gorm.DB {
	startOfDay := f.Date.Truncate(24 * time.Hour)
	endOfDay := startOfDay.Add(24 * time.Hour)
	return db.Where(fmt.Sprintf("%s >= ? AND %s < ?", f.Column, f.Column), startOfDay, endOfDay)
}

func GetDateFilter(c *gin.Context, column string) Filter {
	dateStr := c.Query("date")
	if dateStr == "" {
		return nil
	}

	date, err := time.Parse(time.DateOnly, dateStr)
	if err != nil {
		return nil
	}

	return &DateRangeFilter{Column: column, Date: date}
}
