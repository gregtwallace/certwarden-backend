package backup

import (
	"errors"
	"fmt"
)

var (
	ErrFileError = errors.New("data backup file error")
)

func errorFileError(path, detail string) error {
	return fmt.Errorf("%w: %s (%s)", ErrFileError, detail, path)
}
