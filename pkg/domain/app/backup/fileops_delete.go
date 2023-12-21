package backup

import (
	"errors"
	"fmt"
	"os"
	"sort"
	"time"
)

// deleteOlderThan deletes backup files that are older than the specified duration
func (service *Service) deleteOlderThan(maxAge time.Duration) (oldestRemaining time.Time, err error) {
	filesInfo, err := service.listBackupFiles()
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to delete aged backup files (%s)", err)
	}

	anyErr := false
	oldestRemainingUnixTime := time.Now().Unix()

	for i := range filesInfo {
		thisFileUnixTime := int64(filesInfo[i].unixTime())

		// delete if file's unix time + maxAge is before (<) Now()
		if time.Unix(thisFileUnixTime, 0).Add(maxAge).Before(time.Now()) {
			err = os.Remove(service.cleanDataStorageBackupPath + "/" + filesInfo[i].Name)
			if err != nil {
				anyErr = true
				service.logger.Errorf("failed to delete aged backup file %s (%s)", filesInfo[i].Name, err)
			}
		} else if thisFileUnixTime < oldestRemainingUnixTime {
			// if NOT deleted AND file time is oldest currently known, update oldest val
			oldestRemainingUnixTime = thisFileUnixTime
		}
	}

	// return err if issue with any file
	if anyErr {
		return time.Unix(oldestRemainingUnixTime, 0), errors.New("failed to delete some aged backup file(s), review other log entries for more details")
	}

	return time.Unix(oldestRemainingUnixTime, 0), nil
}

// deleteCountGreaterThan deletes the oldest backup files until the count of backup
// files is equal to the specified count
func (service *Service) deleteCountGreaterThan(count int) error {
	// if count <= 0, do nothing
	if count <= 0 {
		return nil
	}

	filesInfo, err := service.listBackupFiles()
	if err != nil {
		return fmt.Errorf("failed to delete backup files over max count (%s)", err)
	}

	// compare backup len to the specified count
	if len(filesInfo) <= count {
		// if not over max, done
		return nil
	}

	// sort backup files by age
	sort.Slice(filesInfo, func(i, j int) bool {
		// sort smallest to the end of the slice
		return filesInfo[i].unixTime() > filesInfo[j].unixTime()
	})

	anyErr := false

	// range through the file indexes greater than the max count and delete them
	for i := count; i < len(filesInfo); i++ {
		err = os.Remove(service.cleanDataStorageBackupPath + "/" + filesInfo[i].Name)
		if err != nil {
			anyErr = true
			service.logger.Errorf("failed to delete old backup file %s that was over max count (%s)", filesInfo[i].Name, err)
		}
	}

	// return err if issue with any file
	if anyErr {
		return errors.New("failed to delete some old backup file(s) over max count, review other log entries for more details")
	}

	return nil
}
