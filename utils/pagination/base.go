package pagination

import (
	"encoding/json"
	"errors"
	"math"
	"reflect"
	"strconv"

	sq "github.com/Masterminds/squirrel"
	"github.com/gofiber/fiber/v2"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

const (
	DefaultPerPage    int    = 30
	MaxPerPage        int    = 500
	PageQueryParam    string = "page"
	PerPageQueryParam string = "perPage"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Result defines the return result of a pagination
type PaginationResult struct {
	Page       int               `json:"page"`
	PerPage    int               `json:"perPage"`
	TotalItems int               `json:"totalItems"`
	TotalPages int               `json:"totalPages"`
	Items      []json.RawMessage `json:"items"`
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Pagination represents a pagination.
type Pagination struct {
	page       int
	perPage    int
	totalItems int
	totalPages int
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// New creates and returns a pagination
func NewFromApi(f *fiber.Ctx) *Pagination {
	page := page(f)
	perPage := perPage(f)

	return &Pagination{
		page:    page,
		perPage: perPage,
	}
}

// New creates and returns a pagination
func New(p int, pp int) *Pagination {
	// Normalize the page value
	if p <= 0 {
		p = 1
	}

	// Normalize the perPage value
	if pp <= 0 {
		pp = DefaultPerPage
	} else if pp > MaxPerPage {
		pp = MaxPerPage
	}

	return &Pagination{
		page:    p,
		perPage: pp,
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (p *Pagination) Limit() int {
	return p.perPage
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Offset calculates and return an offset value
func (p *Pagination) Offset() int {
	return p.perPage * (p.page - 1)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// TotalItems returns the total number of items
func (p *Pagination) TotalItems() int {
	return p.totalItems
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// TotalPages returns the total number of pages
func (p *Pagination) TotalPages() int {
	return p.totalPages
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// SetCount sets the total number of items and calculates the total number of pages
func (p *Pagination) SetCount(count int) {
	p.totalItems = count
	p.totalPages = int(math.Ceil(float64(p.totalItems) / float64(p.perPage)))
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Apply applies the pagination to a query builder
func (p *Pagination) Apply(queryBuilder sq.SelectBuilder) sq.SelectBuilder {
	return queryBuilder.Offset(uint64(p.Offset())).Limit(uint64(p.Limit()))
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// BuildResult builds a result object from the pagination values, which is suitable for a HTTP
// response
func (p *Pagination) BuildResult(m any) (*PaginationResult, error) {

	// Slice to hold the marshaled items
	items := []json.RawMessage{}

	// Use reflection to ensure m is a slice and iterate over it
	v := reflect.ValueOf(m)
	if v.Kind() == reflect.Slice {
		for i := 0; i < v.Len(); i++ {
			raw, err := json.Marshal(v.Index(i).Interface())
			if err != nil {
				return nil, err
			}
			items = append(items, raw)
		}
	} else {
		return nil, errors.New("input is not a slice")
	}

	return &PaginationResult{
		Page:       p.page,
		PerPage:    p.perPage,
		TotalItems: p.totalItems,
		TotalPages: p.totalPages,
		Items:      items,
	}, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Page normalizes the `page` field
func page(f *fiber.Ctx) int {
	res, err := strconv.Atoi(f.Query(PageQueryParam))
	if err != nil || res <= 0 {
		return 1
	} else {
		return res
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// PerPage normalizes the `perPage` field
func perPage(f *fiber.Ctx) int {
	res, err := strconv.Atoi(f.Query(PerPageQueryParam))
	if err != nil || res <= 0 {
		return DefaultPerPage
	} else if res > MaxPerPage {
		return MaxPerPage
	}

	return res
}
