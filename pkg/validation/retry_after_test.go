package validation

import (
	"errors"
	"testing"
	"time"
)

// nowTestingFunc returns: Wed Jan 01 2020 11:05:28 GMT+0000
// http date format: `Wed, 01 Jan 2020 11:05:28 GMT`
var nowTestingFunc = func() time.Time { return time.Unix(1577876728, 0) }

// getANSICoffset returns a value to adjust the ANSIC format by (based on local time condition)
func getANSICoffset() time.Duration {
	_, tzOffset := time.Now().Zone()
	return time.Duration(tzOffset) * time.Second
}

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
		expectedError:   ErrHTTPRetryAfterNegativeSeconds,
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
		expectedError:   ErrHTTPRetryAfterNegativeSeconds,
	},
	// valid ANSIC
	{
		retryAfterValue: "Wed Apr  1 18:05:28 2020",
		expectedValue:   time.Unix(1585778728, 0).Add(getANSICoffset()),
		expectedError:   nil,
	},
	// valid RFC850
	{
		retryAfterValue: "Wednesday, 09-Jan-20 03:05:27 EST",
		expectedValue:   time.Unix(1578557127, 0),
		expectedError:   nil,
	},
}

// test invalid formats
var retryAfterInvalidFormatTests = []string{
	time.Layout,
	time.UnixDate,
	time.RubyDate,
	time.RFC822,
	time.RFC822Z,
	time.RFC1123,
	time.RFC1123Z,
	time.RFC3339,
	time.RFC3339Nano,
	time.Kitchen,
	time.Stamp,
	time.StampMilli,
	time.StampMicro,
	time.StampNano,
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
	// run tests
	for _, aTest := range retryAfterTests {
		parsedVal, err := parseRetryAfter(aTest.retryAfterValue, nowTestingFunc)
		if !errors.Is(err, aTest.expectedError) {
			t.Errorf("retry after value '%s' expected error '%v' but got '%v'", aTest.retryAfterValue, aTest.expectedError, err)
		}

		if !parsedVal.Equal(aTest.expectedValue) {
			t.Errorf("retry after value '%s' expected parse to '%s' but got '%s'", aTest.retryAfterValue, aTest.expectedValue.UTC(), parsedVal.UTC())
		}
	}

	// invalid format
	for _, invalidFormatString := range retryAfterInvalidFormatTests {
		parsedVal, err := parseRetryAfter(invalidFormatString, nowTestingFunc)
		if !errors.Is(err, ErrHTTPRetryAfterInvalidFormat) {
			t.Errorf("retry after value '%s' expected error '%v' but got '%v'", invalidFormatString, ErrHTTPRetryAfterInvalidFormat, err)
		}

		if !parsedVal.Equal(time.Time{}) {
			t.Errorf("retry after value '%s' expected parse to '%s' but got '%s'", invalidFormatString, time.Time{}, parsedVal)
		}
	}
}
