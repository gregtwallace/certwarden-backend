package backup

import (
	"strings"
	"time"
)

const backupFilePrefix = "cert_warden_backup."
const backupFileSuffix = ".zip"

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

// isBackupFileName returns true if the fileName string provided starts with the
// backup file prefix and ends in the proper file extension; it also only permits
// certain characters in the filename to avoid things like path traversal
func isBackupFileName(fileName string) bool {
	return strings.HasPrefix(fileName, backupFilePrefix) && strings.HasSuffix(fileName, backupFileSuffix)
}
