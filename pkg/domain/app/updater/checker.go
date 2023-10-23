package updater

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"reflect"
	"sync"
	"time"
)

// version.json URL
const versionJsonURL = "https://www.legocerthub.com/version.json"

// errors
var (
	errBadChannel = errors.New("channel not found in remote version.json")
)

// Channels
type Channel string

const (
	ChannelRelease Channel = "release"
	ChannelBeta    Channel = "beta"
)

// fetchNewVersion fetches the remote version.json and updates the service
// newVersionAvailable and newVersionInfo accordingly.
func (service *Service) fetchNewVersion() error {
	service.logger.Debugf("fetching most recent version information from %s", versionJsonURL)

	response, err := service.httpClient.Get(versionJsonURL)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}

	var fetchedVersions []versionInfo
	err = json.Unmarshal(body, &fetchedVersions)
	// check for Unmarshal error
	if err != nil {
		return err
	}

	// find the version for the desired channel
	var newestVersion versionInfo
	foundChannel := false
	for _, version := range fetchedVersions {
		if version.Channel == service.checkChannel {
			foundChannel = true
			newestVersion = version
			break
		}
	}

	// if channel doesn't exist in version.json, error
	if !foundChannel {
		return errBadChannel
	}

	// lock since timestamp updates no matter what
	service.newVersion.mu.Lock()
	defer service.newVersion.mu.Unlock()

	changed := !reflect.DeepEqual(newestVersion, service.newVersion.info)
	service.newVersion.lastCheck = time.Now()

	// Only update if the data has changed
	if changed {
		// is the new version actually new?
		newer, err := newestVersion.isNewerThan(service.currentVersion)
		if err != nil {
			service.logger.Errorf("failed to compare new version and old version (%v)", err)
			return err
		}
		if newer {
			service.logger.Infof("new version (%s) of app is available", newestVersion.Version)
			service.newVersion.available = true
		} else {
			service.newVersion.available = false
		}

		// update newest version info
		service.newVersion.info = &newestVersion
		service.logger.Debug("new version info updated succesfully")

	} else {
		// no change
		service.logger.Debug("remote version.json has not changed")
	}

	return nil
}

// backgroundChecker starts a goroutine that is an indefinite for loop
// that checks for an updated version of the application.
// The interval is shorter if the previous update encountered an error.
func (service *Service) backgroundChecker(ctx context.Context, wg *sync.WaitGroup) {
	// log start and update wg
	service.logger.Info("starting updater service")
	wg.Add(1)

	// service routine
	go func(service *Service, ctx context.Context, wg *sync.WaitGroup) {
		// can adjust wait times if desired
		defaultWaitTime := 24 * time.Hour

		// loop logic for periodic updates
		var waitTime time.Duration

		for {
			waitTime = defaultWaitTime
			err := service.fetchNewVersion()
			if err != nil {
				service.logger.Errorf("update check failed (%v)", err)
			}

			// sleep or wait for shutdown context to be done
			select {
			case <-ctx.Done():
				// close routine
				service.logger.Info("updater service shutdown complete")
				wg.Done()
				return

			case <-time.After(waitTime):
				// sleep and retry
			}
		}
	}(service, ctx, wg)
}
