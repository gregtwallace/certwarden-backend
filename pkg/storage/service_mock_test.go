package storage_test

import (
	"certwarden-backend/pkg/storage"
	"context"
	"errors"
	"io"
	"os"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

const (
	testDataDbFile  = "../../test_data/testdata_v11.db"
	tempFileStorage = "../../test_data/tmp/"
)

// fake app for this package
type fakeApp struct {
	appDataPath string
	logger      *zap.SugaredLogger
}

func (fa *fakeApp) GetDataStorageAppDataPath() string {
	return fa.appDataPath
}

func (fa *fakeApp) GetLogger() *zap.SugaredLogger {
	return fa.logger
}

func (fa *fakeApp) GetShutdownContext() context.Context {
	return context.Background()
}

func (fa *fakeApp) CreateBackupOnDisk() error {
	return errors.New("unimplemented")
}

func makeFakeApp(t *testing.T, appDataPath string) *fakeApp {
	logger := zaptest.NewLogger(t, zaptest.Level(zap.FatalLevel)).Sugar() // use fatal to avoid log output

	return &fakeApp{
		appDataPath: appDataPath,
		logger:      logger,
	}
}

// openStorageWithTestData makes a copy of the testing data db and then returns a storage service
// that uses the copy; Cleanup() should be called at the end of the test
func openStorageWithTestData(t *testing.T, testName string) (_ *storage.Storage, _ error) {
	thisTestFolder := tempFileStorage + testName

	// copy test data to temp appDataPath for tests to run
	_, err := os.Stat(thisTestFolder)
	if err == nil {
		os.RemoveAll(thisTestFolder)
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}

	err = os.MkdirAll(thisTestFolder, os.FileMode(0o777))
	if err != nil {
		return nil, err
	}

	testDataF, err := os.Open(testDataDbFile)
	if err != nil {
		t.Error(err)
		return nil, err
	}
	defer testDataF.Close()

	testDataCopyF, err := os.Create(thisTestFolder + "/appdata.db")
	if err != nil {
		return nil, err
	}

	_, err = io.Copy(testDataCopyF, testDataF)
	if err != nil {
		return nil, err
	}

	err = testDataCopyF.Sync()
	if err != nil {
		return nil, err
	}

	fakeApp := makeFakeApp(t, thisTestFolder)

	storage, err := storage.OpenStorage(fakeApp)
	if err != nil {
		return nil, err
	}

	return storage, nil
}
