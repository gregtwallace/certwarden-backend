package app

import (
	"errors"
	"legocerthub-backend/utils/acme_utils"
	"reflect"
	"time"
)

// UpdateDirectory updates the directory for the specified environment. It
// returns an error if it can't update the specified directory.
func (app *Application) updateDirectory(env string, err chan<- error) {
	var destinationDirAddr *acme_utils.AcmeDirectory

	switch env {
	case "prod":
		destinationDirAddr = &app.Acme.ProdDir
	case "staging":
		destinationDirAddr = &app.Acme.StagingDir
	default:
		err <- errors.New("invalid environment")
		return
	}

	app.Logger.Printf("Fetching latest directory from ACME for %s environment.", env)

	dir, anErr := acme_utils.GetAcmeDirectory(env)
	err <- anErr
	if anErr != nil {
		app.Logger.Printf("Error updating %s's directory.", env)
		return
	} else if reflect.DeepEqual(dir, *destinationDirAddr) {
		app.Logger.Printf("%s environment directory already up to date.", env)
	} else {
		*destinationDirAddr = dir
		app.Logger.Printf("%s environment directory updated succesfully.", env)
	}

	return
}

// UpdateAllDirectories will attempt to update both the prod and staging
// directories.  It returns an error if one or both updates are not
// successful.
func (app *Application) UpdateAllDirectories() error {
	errs := make(chan error)

	app.Logger.Println("Updating all directories from ACME upstream.")

	// production
	go app.updateDirectory("prod", errs)

	// staging
	go app.updateDirectory("staging", errs)

	// return an error if any directory failed to update
	if <-errs != nil || <-errs != nil {
		return errors.New("Error(s) updating one or more ACME directories.")
	}

	return nil
}

// BackgroundDirManagement runs an indefinite for loop that checks for
// directory updates at the specified time interval.  The interval is shorter
// if the previous loop encountered an error, based on the assumption of a
// temporary outage.
func (app *Application) BackgroundDirManagement() {
	defaultWaitTime := 24 * time.Hour
	failWaitTime := 15 * time.Minute

	waitTime := defaultWaitTime

	for {
		time.Sleep(waitTime)

		err := app.UpdateAllDirectories()
		if err != nil {
			app.Logger.Printf("error: %v, will retry directory update shortly.", err)
			// if something failed, decrease the wait to try again
			waitTime = failWaitTime
		} else {
			waitTime = defaultWaitTime
		}
	}
}
