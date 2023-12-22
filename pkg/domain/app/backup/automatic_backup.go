package backup

import (
	"time"
)

// create a backup every X amount of time
const DefaultBackupEnabled = true
const DefaultBackupDays = 7

// keep backups for X amount of time OR max count of backups
// negative means indefinite / don't enforce
const DefaultBackupRetentionDays = 180
const DefaultBackupRetentionCount = -1

// Config holds the config for the automatic backup service
type Config struct {
	Enabled      *bool `yaml:"enabled" json:"enabled"`
	IntervalDays *int  `yaml:"interval_days" json:"interval_days"`

	Retention struct {
		MaxDays  *int `yaml:"max_days" json:"max_days"`
		MaxCount *int `yaml:"max_count" json:"max_count"`
	} `yaml:"retention" json:"retention"`
}

// StartAutoBackupService starts the automated backup process using the specified
// configuration params
func (service *Service) StartAutoBackupService(app App, cfg *Config) {
	// set service config
	service.config = cfg

	// read config and convert vals as needed
	enabled := DefaultBackupEnabled
	if cfg != nil && cfg.Enabled != nil {
		enabled = *cfg.Enabled
	}
	backupInterval := DefaultBackupDays * 24 * time.Hour
	if cfg != nil && cfg.IntervalDays != nil {
		backupInterval = time.Duration(*cfg.IntervalDays) * 24 * time.Hour
	}
	retentionDuration := DefaultBackupRetentionDays * 24 * time.Hour
	if cfg != nil && cfg.Retention.MaxDays != nil {
		retentionDuration = time.Duration(*cfg.Retention.MaxDays) * 24 * time.Hour
	}
	retentionCount := DefaultBackupRetentionCount
	if cfg != nil && cfg.Retention.MaxCount != nil {
		retentionCount = *cfg.Retention.MaxCount
	}

	// return no-op if cfg not enabled or interval <= 0 days
	if !enabled || backupInterval <= 0 {
		service.logger.Warnf("not starting automatic backup service (not enabled or invalid backup interval)")
		return
	}

	// find newest and oldest backups' timestamp
	newestBackupUnixTime := -1
	// use now as baseline for finding min
	oldestBackupUnixTime := int(time.Now().Unix())

	// list backups
	filesInfo, err := service.listBackupFiles()
	if err != nil {
		// no op, skip down
	} else {
		// got list of files -  find newest backup time stamp
		for i := range filesInfo {
			thisFileUnixTime := filesInfo[i].unixTime()

			// update max if this time is bigger
			if thisFileUnixTime > newestBackupUnixTime {
				newestBackupUnixTime = thisFileUnixTime
			}

			// update min if this time is smaller
			if thisFileUnixTime < oldestBackupUnixTime {
				oldestBackupUnixTime = thisFileUnixTime
			}
		}
	}

	// calculate times
	lastBackupTime := time.Unix(int64(newestBackupUnixTime), 0)
	oldestBackupTime := time.Unix(int64(oldestBackupUnixTime), 0)

	// do a backup on start if backup is overdue
	if time.Since(lastBackupTime) > backupInterval {
		newBackupFileDetails, err := service.CreateBackupOnDisk()
		if err != nil {
			service.logger.Errorf("failed to create automatic on disk backup (%s)", err)
		}

		// update last backup time to the one that was just made
		lastBackupTime = time.Unix(int64(newBackupFileDetails.unixTime()), 0)
	}

	// shutdown context and wg
	shutdownCtx := app.GetShutdownContext()
	shutdownWg := app.GetShutdownWaitGroup()

	service.logger.Infof("starting automatic data backup service")

	// start go routine for auto backups
	shutdownWg.Add(1)
	go func() {
		nextBackup := lastBackupTime.Add(backupInterval)

		for {
			select {
			case <-shutdownCtx.Done():
				// exit
				service.logger.Info("automatic data backup service shutdown complete")
				shutdownWg.Done()
				return

			case <-time.After(time.Until(nextBackup)):
				// continue and run
			}

			// create a backup and update last backup time
			newBackupFileDetails, err := service.CreateBackupOnDisk()
			if err != nil {
				// fail: try again in a day
				nextBackup = time.Now().Add(24 * time.Hour)

				service.logger.Errorf("failed to create automatic on disk backup (%s), will try again at %s", err, nextBackup)

			} else {
				// success: run again after next interval
				nextBackup = time.Unix(int64(newBackupFileDetails.unixTime()), 0).Add(backupInterval)

				err = service.deleteCountGreaterThan(retentionCount)
				if err != nil {
					service.logger.Errorf("failed to delete backups over retention count (%s)", err)
				}
			}
		}
	}()

	// start go routine for auto delete (if time based retention policy is set)
	if retentionDuration > 0 {
		shutdownWg.Add(1)
		go func() {
			service.logger.Infof("starting data backup time based deletion service")

			nextDelete := oldestBackupTime.Add(retentionDuration)
			for {
				select {
				case <-shutdownCtx.Done():
					// exit
					service.logger.Info("data backup time based deletion service shutdown complete")
					shutdownWg.Done()
					return

				case <-time.After(time.Until(nextDelete)):
					// continue and run
				}

				// do delete
				oldestBackupTime, err = service.deleteOlderThan(retentionDuration)
				if err != nil {
					// fail: try again in a day
					nextDelete = time.Now().Add(24 * time.Hour)

					service.logger.Errorf("failed to delete backups over retention time duration (%s), will try again at %s", err, nextDelete)

				} else {
					// success: run again when oldest backup hits threshhold
					nextDelete = oldestBackupTime.Add(retentionDuration)
				}
			}
		}()
	}
}
