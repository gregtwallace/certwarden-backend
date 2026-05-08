package sqlite3

import (
	"errors"
	"fmt"
	"os"
)

// pre-migration file path
const oldFilePath = "./data"

// migrateDbFileLocation moves the db file from its "old" location to the current one
func migrateDbFileLocation(fromFile, toFile string) (didMigrate bool, _ error) {
	// stat old location
	_, err := os.Stat(fromFile)
	if err != nil {
		// no old file
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}
		// any other error
		return false, fmt.Errorf("sqlite3: could not check for db file at old location (%w) (%w)", err, errStatFromFailed)
	}

	// old file exists
	_, err = os.Stat(toFile)
	if err == nil {
		return false, fmt.Errorf("sqlite3: cannot migrate (%w)", errStatAlreadyExists)
	} else if !errors.Is(err, os.ErrNotExist) {
		return false, fmt.Errorf("sqlite3: could not check for db file at new location (%w) (%w)", err, errStatToFailed)
	}

	err = os.Rename(fromFile, toFile)
	if err != nil {
		return false, fmt.Errorf("sqlite3: failed to move existing db file from %s to %s (%w) (%w)", fromFile, toFile, err, errRenameFailed)
	}

	return true, nil
}
