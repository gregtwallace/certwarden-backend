package environment

import (
	"strings"
)

// Params holds environment params
type Params struct {
	paramSlice     []string
	paramKeyValMap map[string]string
}

// NewParams creates a new Params. If any of the envParams that are passed in are
// invalid, they will be excluded from Params but returned as part of the invalidParams
// slice.
func NewParams(envParams []string) (p *Params, invalidParams []string) {
	// if envParams nil, return empty p
	if envParams == nil {
		return &Params{
			paramSlice:     []string{},
			paramKeyValMap: make(map[string]string),
		}, nil
	}

	// make return slice, map, and invalid slice
	pSlice := []string{}
	pMap := make(map[string]string)
	invalidParams = []string{}

	// for each param, parse to get key and value
	for _, oneParam := range envParams {
		paramName, paramValue, sepFound := strings.Cut(oneParam, "=")
		if !sepFound {
			// invalid param, append to invalid
			invalidParams = append(invalidParams, oneParam)
			continue
		}

		// remove quoting, if present
		if (strings.HasPrefix(paramName, "\"") && strings.HasSuffix(paramName, "\"")) ||
			(strings.HasPrefix(paramName, "'") && strings.HasSuffix(paramName, "'")) {
			paramName = paramName[1 : len(paramName)-1]
		}
		if (strings.HasPrefix(paramValue, "\"") && strings.HasSuffix(paramValue, "\"")) ||
			(strings.HasPrefix(paramValue, "'") && strings.HasSuffix(paramValue, "'")) {
			paramValue = paramValue[1 : len(paramValue)-1]
		}

		// add to valid slice & map
		pSlice = append(pSlice, oneParam)
		pMap[paramName] = paramValue
	}

	p = &Params{
		paramSlice:     pSlice,
		paramKeyValMap: pMap,
	}

	return p, invalidParams
}

// StringSlice returns a slice of strings for the envrionment params. Each string is
// in the format Key_Name=Value
func (p *Params) StringSlice() []string {
	if p == nil {
		return nil
	}

	return p.paramSlice
}

// Map returns a key value map for the envrionment params. The map is of the form
// map[key]value
func (p *Params) KeyValMap() map[string]string {
	if p == nil {
		return nil
	}

	return p.paramKeyValMap
}
