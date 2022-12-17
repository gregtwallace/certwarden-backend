package pagination_sort

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type Query struct {
	limit  int
	offset int
	sort   sorting
}

type sorting struct {
	field     string
	direction string
}

// funcs to access Query members. These allow access to
// the data but prevent changing it outside of this pkg.
func (q Query) Limit() int {
	return q.limit
}
func (q Query) Offset() int {
	return q.offset
}
func (q Query) SortField() string {
	return q.sort.field
}
func (q Query) SortDirection() string {
	// if value is bad, return 'asc'
	if q.sort.direction != "asc" && q.sort.direction != "desc" {
		return "asc"
	}
	return q.sort.direction
}

// QueryAll is a query that returns all records
var QueryAll = Query{
	limit:  -1,
	offset: 0,
}

// configure paramters for validation
const (
	defaultLimit = 25
	minLimit     = 1
	maxLimit     = 1000

	defaultOffset = 0
	minOffset     = 0
)

var validFieldNames = []string{
	"algorithm",
	"description",
	"id",
	"name",
	"subject",
	"valid_to",
}

// ParseRequestToQuery returns pagination and sorting params
// in a Query struct
func ParseRequestToQuery(r *http.Request) Query {
	// get all query params
	v := r.URL.Query()

	return Query{
		limit:  limit(v),
		offset: offset(v),
		sort:   sort(v),
	}
}

// limit parses, sanitizes, and returns the limit from url.Values
func limit(v url.Values) int {
	limitStr := v.Get("limit")

	// convert to int
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		return defaultLimit
	}

	// set acceptable bounds
	switch {
	case limit > maxLimit:
		limit = maxLimit
	case limit < minLimit:
		limit = defaultLimit
	default:
		// no-op
	}

	return limit
}

// offset parses, sanitizes, and returns the offset from url.Values
func offset(v url.Values) int {
	offsetStr := v.Get("offset")

	// convert to int
	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		return defaultOffset
	}

	// set acceptable bounds
	switch {
	// no max
	case offset < minOffset:
		offset = defaultOffset
	default:
		// no-op
	}

	return offset
}

// sort parses, sanitizes, and returns the sort parameters from url.Values
func sort(v url.Values) sorting {
	sortStr := v.Get("sort")

	// if blank, return blank
	if sortStr == "" {
		return sorting{}
	}

	// split string and confirm proper length
	sortSplit := strings.Split(sortStr, ".")
	if len(sortSplit) != 2 {
		return sorting{}
	}

	// get field and dir from the split slice
	field := strings.ToLower(sortSplit[0])
	direction := strings.ToLower(sortSplit[1])

	// validate field name is in acceptable list
	found := false
	for i := range validFieldNames {
		// if found
		if field == validFieldNames[i] {
			found = true
			break
		}
	}
	if !found {
		return sorting{}
	}

	// validate direction
	if direction != "asc" && direction != "desc" {
		return sorting{}
	}

	return sorting{
		field:     field,
		direction: direction,
	}
}
