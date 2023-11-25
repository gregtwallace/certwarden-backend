package providers

import (
	"errors"
	"fmt"
)

var (
	errWrongTag = errors.New("manager provider action failed due to tag mismatch")

	errBadID = func(id int) error { return fmt.Errorf("no provider exists with id %d", id) }
)
