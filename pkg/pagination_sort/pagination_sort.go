package pagination_sort

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type sorting struct {
	field      string
	descending bool
}

type Query struct {
	limit  int
	offset int
	sort   sorting
}

// funcs to access Query members. These allow access to data but prevent changing
// it outside of this pkg.
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
	if q.sort.descending {
		return "desc"
	}

	return "asc"
}

// QueryAll is a query that returns all records
var QueryAll = Query{
	limit:  -1,
	offset: 0,
}

var validFieldNames = []string{
	"accountname",
	"algorithm",
	"created_at",
	"description",
	"email",
	"id",
	"is_staging",
	"keyname",
	"name",
	"status",
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

// limit paramters for validation
const (
	limitDefault = 20
	limitMin     = 1
	limitMax     = 1000
)

// limit parses, sanitizes, and returns the limit from url.Values
func limit(v url.Values) int {
	limitStr := v.Get("limit")

	// convert to int
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		return limitDefault
	}

	// set acceptable bounds
	if limit > limitMax {
		return limitMax
	}

	if limit < limitMin {
		return limitMin
	}

	return limit
}

// offset parses, sanitizes, and returns the offset from url.Values. If there is an
// issue with the offset, the default 0 is returned.
func offset(v url.Values) int {
	offsetStr := v.Get("offset")

	// convert to int
	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		return 0
	}

	// if invalid offset, use default
	if offset < 0 {
		return 0
	}

	return offset
}

// sort parses, sanitizes, and returns the sort parameters from url.Values. If either
// the field or direction are invalid, default sorting of unspecified field and ascending
// is returned.
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
		field:      field,
		descending: direction == "desc",
	}
}
