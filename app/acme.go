package app

import (
	"errors"
	"legocerthub-backend/utils/acme_utils"
	"reflect"
	"time"
)

type AppAcme struct {
	ProdDir    acme_utils.AcmeDirectory
	StagingDir acme_utils.AcmeDirectory
}

// UpdateDirectory updates the directory for the specified environment. It
// returns an error if it can't update the specified directory.
func (app *Application) updateDirectory(env string) error {
	var destinationDirAddr *acme_utils.AcmeDirectory

	switch env {
	case "prod":
		destinationDirAddr = &app.Acme.ProdDir
	case "staging":
		destinationDirAddr = &app.Acme.StagingDir
	default:
		return errors.New("invalid environment")
	}

	app.Logger.Printf("Fetching latest directory from ACME for %s environment.", env)

	dir, err := acme_utils.GetAcmeDirectory(env)
	if err != nil {
		app.Logger.Printf("Error updating %s's directory.", env)
		return err
	} else if reflect.DeepEqual(dir, *destinationDirAddr) {
		app.Logger.Printf("%s environment directory already up to date.", env)
	} else {
		*destinationDirAddr = dir
		app.Logger.Printf("%s environment directory updated succesfully.", env)
	}

	return nil
}

// UpdateAllDirectories will attempt to update both the prod and staging
// directories.  It returns an error if one or both updates are not
// successful.
func (app *Application) UpdateAllDirectories() error {
	app.Logger.Println("Updating all directories from ACME upstream.")

	// production
	prodErr := app.updateDirectory("prod")

	// staging
	stagingErr := app.updateDirectory("staging")

	// return an error if any directory failed to update
	if prodErr != nil || stagingErr != nil {
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
			app.Logger.Printf("error: %v, will retry shortly.", err)
			// if something failed, decrease the wait to try again
			waitTime = failWaitTime
		} else {
			waitTime = defaultWaitTime
		}
		app.Logger.Println("Checking directories again.")
	}
}
