package output

import (
	"crypto/sha1"
	"fmt"
)

// etagValue returns the sha1 of the data with quotes around it to use
// as an Etag value
func etagValue(data []byte) string {
	return fmt.Sprintf("\"%x\"", sha1.Sum(data))
}
