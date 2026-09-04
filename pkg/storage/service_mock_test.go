package storage_test

import (
	"certwarden-backend/pkg/helpers_test"
	"certwarden-backend/pkg/pagination_sort"
	"certwarden-backend/pkg/storage"
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
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
func openStorageWithTestData(t *testing.T, testName string) *storage.Storage {
	thisTestFolder := tempFileStorage + testName

	// copy test data to temp appDataPath for tests to run
	helpers_test.MakeTempStorage(t, thisTestFolder)

	testDataF, err := os.Open(testDataDbFile)
	if err != nil {
		t.Fatalf("failed to open test data file '%s'", err)
	}
	t.Cleanup(func() {
		err := testDataF.Close()
		if err != nil {
			t.Errorf("failed to close testDataF (%s)", err)
		}
	})

	testDataCopyF, err := os.Create(thisTestFolder + "/appdata.db")
	if err != nil {
		t.Fatalf("failed to create test data file copy '%s'", err)
	}
	t.Cleanup(func() {
		err := testDataCopyF.Close()
		if err != nil {
			t.Errorf("failed to close testDataCopyF (%s)", err)
		}
	})

	_, err = io.Copy(testDataCopyF, testDataF)
	if err != nil {
		t.Fatalf("failed to copy test data '%s'", err)
	}

	err = testDataCopyF.Sync()
	if err != nil {
		t.Fatalf("failed to sync test data '%s'", err)
	}

	fakeApp := makeFakeApp(t, thisTestFolder)

	store, err := storage.OpenStorage(fakeApp)
	if err != nil {
		t.Fatalf("failed to open storage '%s'", err)
	}
	t.Cleanup(func() {
		err := store.Close()
		if err != nil {
			t.Errorf("failed to close storage (%s)", err)
		}
	})

	return store
}

// queryBuilderForTest generates a Query for use in tests
func queryBuilderForTest(limit, offset int, sortField string, sortDesc bool) pagination_sort.Query {
	sortDirText := "asc"
	if sortDesc {
		sortDirText = "desc"
	}

	p := url.Values{
		"limit":  {strconv.Itoa(limit)},
		"offset": {strconv.Itoa(offset)},
		"sort":   {sortField + "." + sortDirText},
	}

	// make fake request just for query parsing
	r := &http.Request{}
	u, err := url.Parse("https://example.com/")
	if err != nil {
		panic("url must parse failed")
	}
	r.URL = u
	r.URL.RawQuery = p.Encode()

	return pagination_sort.ParseRequestToQuery(r)
}
