package helpers_test

import (
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// GetFunctionName uses reflection to return a string with the function name; if something
// other than a function is passed, GetFunctionName may panic or perform other unspecified
// behaviors. Additionally, if an inline function is passed, the name is likely
// non-deterministic
func GetFunctionName(f interface{}) string {
	// check for nil
	runtimeFunc := runtime.FuncForPC(reflect.ValueOf(f).Pointer())
	if runtimeFunc == nil {
		return "<nil>"
	}

	strs := strings.Split(runtimeFunc.Name(), ".")
	return strs[len(strs)-1]
}

// StringPointerToVal returns the string sp points to, or the string '<nil>' if sp is nil
func StringPointerToVal(sp *string) string {
	if sp != nil {
		return *sp
	}

	return "<nil>"
}

// IntPointerToVal returns the string representation of the value i points to,
// or the string '<nil>' if i is nil
func IntPointerToVal(i *int) string {
	if i != nil {
		return strconv.Itoa(*i)
	}

	return "<nil>"
}

// ErrorToVal returns the error string, or the string '<nil>' if err is nil
func ErrorToVal(err error) string {
	if err != nil {
		return err.Error()
	}

	return "<nil>"
}

// TimeToVal returns the time string (in UTC), formatted in RFC3339,
// or the string '<nil>' if t is nil
func TimeToVal(t *time.Time) string {
	if t != nil {
		return t.UTC().Format(time.RFC3339)
	}

	return "<nil>"
}
