package validation_test

import (
	"certwarden-backend/pkg/validation"
	"errors"
	"fmt"
	"testing"
	"time"
)

// test structure
type retryAfterTest struct {
	retryAfterValue string
	expectedValue   time.Time
	expectedError   error
}

var retryAfterTests = []retryAfterTest{
	// valid +30 seconds
	{
		retryAfterValue: "30",
		expectedValue:   time.Unix(1577876758, 0),
		expectedError:   nil,
	},
	// valid 0 seconds
	{
		retryAfterValue: "0",
		expectedValue:   time.Unix(1577876728, 0),
		expectedError:   nil,
	},
	// invalid negative seconds
	{
		retryAfterValue: "-2",
		expectedValue:   time.Time{},
		expectedError:   validation.ErrHTTPRetryAfterNegativeSeconds,
	},
	// valid ahead of now
	{
		retryAfterValue: "Tue, 07 Jan 2020 23:20:47 GMT",
		expectedValue:   time.Unix(1578439247, 0),
		expectedError:   nil,
	},
	// valid now
	{
		retryAfterValue: "Wed, 01 Jan 2020 11:05:28 GMT",
		expectedValue:   time.Unix(1577876728, 0),
		expectedError:   nil,
	},
	// invalid date before 'now'
	{
		retryAfterValue: "Wed, 01 Jan 2020 11:05:27 GMT",
		expectedValue:   time.Time{},
		expectedError:   validation.ErrHTTPRetryAfterNegativeSeconds,
	},
	// valid ANSIC
	{
		retryAfterValue: "Wed Apr  1 22:05:28 2020",
		expectedValue:   time.Unix(1585778728, 0),
		expectedError:   nil,
	},
	// valid RFC850
	{
		retryAfterValue: "Wednesday, 09-Jan-20 08:05:27 UTC",
		expectedValue:   time.Unix(1578557127, 0),
		expectedError:   nil,
	},
}

// test invalid formats
var retryAfterInvalidFormatTests = []string{
	time.Layout,
	"Mon Jan  2 15:04:05 MST 2006", // time.UnixDate,
	time.RubyDate,
	time.RFC822,
	time.RFC822Z,
	time.RFC1123,
	time.RFC1123Z,
	time.RFC3339,
	time.RFC3339Nano,
	time.Kitchen,
	"Jan  2 15:04:05",           // time.Stamp,
	"Jan  2 15:04:05.000",       // time.StampMilli,
	"Jan  2 15:04:05.000000",    // time.StampMicro,
	"Jan  2 15:04:05.000000000", // time.StampNano,
	time.DateTime,
	time.DateOnly,
	time.TimeOnly,

	"Wed Jan 01 2020 11:05:28 GMT+0000",
	"2020-01-01 06:05:58 -0500 EST",
	"someval",
	"",
}

// run all Retry After validation tests
func TestValidation_RetryAfterValid(t *testing.T) {
	// set time "now"
	// Testing "now" time is:
	// Wed Jan 01 2020 11:05:28 GMT+0000
	// http date format: `Wed, 01 Jan 2020 11:05:28 GMT`
	revertToDefaultTimeNow := validation.SetTimeNow(t, time.Unix(1577876728, 0))
	t.Cleanup(revertToDefaultTimeNow)

	// run tests
	for i, raTest := range retryAfterTests {
		t.Run(fmt.Sprintf("%d: header value %q", i, raTest.retryAfterValue), func(t *testing.T) {
			parsedVal, err := validation.ParseRetryAfter(raTest.retryAfterValue)
			if !errors.Is(err, raTest.expectedError) {
				t.Errorf("retry after value '%s' expected error '%v' but got '%v'", raTest.retryAfterValue, raTest.expectedError, err)
			}

			if !parsedVal.Equal(raTest.expectedValue) {
				t.Errorf("retry after value '%s' expected parse to '%s' but got '%s'", raTest.retryAfterValue, raTest.expectedValue.UTC(), parsedVal.UTC())
			}
		})
	}

	// invalid format
	for i, invalidFormatString := range retryAfterInvalidFormatTests {
		t.Run(fmt.Sprintf("%d: invalid time format %q", i, invalidFormatString), func(t *testing.T) {
			parsedVal, err := validation.ParseRetryAfter(invalidFormatString)
			if !errors.Is(err, validation.ErrHTTPRetryAfterInvalidFormat) {
				t.Errorf("retry after value '%s' expected error '%v' but got '%v'", invalidFormatString, validation.ErrHTTPRetryAfterInvalidFormat, err)
			}

			if !parsedVal.Equal(time.Time{}) {
				t.Errorf("retry after value '%s' expected parse to '%s' but got '%s'", invalidFormatString, time.Time{}, parsedVal)
			}
		})
	}
}
