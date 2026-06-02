package test_helpers

import (
	"reflect"
	"runtime"
	"strings"
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

// ErrorToVal returns the error string, or the string '<nil>' if err is nil
func ErrorToVal(err error) string {
	if err != nil {
		return err.Error()
	}

	return "<nil>"
}
