package backup

import (
	"regexp"
	"strings"
	"time"
)

const backupFilePrefix = "lego_certhub_backup."
const backupFileSuffix = ".zip"

// backupFileDetails contains information about an on disk backup file
type backupFileDetails struct {
	Name      string `json:"name"`
	Size      int    `json:"size"`
	ModTime   int    `json:"modtime"`
	CreatedAt *int   `json:"created_at,omitempty"`
}

// makeBackupZipFileName creates the filename for a new backup created now
func makeBackupZipFileName() (filename string, createdAt int) {
	createdTime := time.Now()

	name := backupFilePrefix + createdTime.Local().Format(time.RFC3339) + backupFileSuffix
	return strings.ReplaceAll(name, ":", "--"), int(createdTime.Unix())
}

// getBackupZipFileTime attempts to return the time from the backup zip filename
func backupZipTime(name string) (time.Time, error) {
	name = strings.ReplaceAll(name, "--", ":")
	timeString := strings.TrimSuffix(strings.TrimPrefix(name, backupFilePrefix), backupFileSuffix)

	fileTime, err := time.Parse(time.RFC3339, timeString)
	if err != nil {
		return time.Time{}, err
	}

	return fileTime, nil
}

// isBackupFile returns true if the fileName string provided starts with the
// backup file prefix and ends in the proper file extension; it also only permits
// certain characters in the filename to avoid things like path traversal
func isBackupFile(fileName string) bool {
	// regex for start prefix, contains on alpha numeric, - _ .   and ends in suffix
	regex := regexp.MustCompile(`^` + regexp.QuoteMeta(backupFilePrefix) + `[A-Za-z0-9-_.]+` + regexp.QuoteMeta(backupFileSuffix) + `$`)

	return regex.MatchString(fileName)
}
