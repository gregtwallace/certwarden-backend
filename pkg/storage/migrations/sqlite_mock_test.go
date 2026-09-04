package migrations_test

import (
	"certwarden-backend/pkg/helpers_test"
	"certwarden-backend/pkg/storage/migrations"
	"certwarden-backend/pkg/storage/sqlite3"
	"context"
	"database/sql"
	"testing"

	"github.com/pressly/goose/v3"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

const tempFileStorage = "../../../test_data/tmp/"

type gooseLogger struct {
	logger *zap.SugaredLogger
}

// implement goose.Logger
func (gl *gooseLogger) Fatalf(format string, v ...any) {
	gl.logger.Fatalf(format, v...)
}

func (gl *gooseLogger) Printf(format string, v ...any) {
	gl.logger.Infof(format, v...)
}

// make app
type fakeApp struct {
	gooseLogger *gooseLogger
	dataPath    string
}

func (fa *fakeApp) GetLogger() *zap.SugaredLogger {
	return fa.gooseLogger.logger
}

func (fa *fakeApp) GetDataStorageAppDataPath() string {
	return fa.dataPath
}

func newFakeApp(t *testing.T, dataPath string) *fakeApp {
	logger := zaptest.NewLogger(t, zaptest.Level(zap.FatalLevel)).Sugar() // use fatal to avoid log output

	gl := &gooseLogger{
		logger: logger,
	}

	goose.SetLogger(gl)

	return &fakeApp{
		gooseLogger: gl,
		dataPath:    dataPath,
	}
}

// setupDBV0 sets up an empty db with goose configured; testName is used to form a temporary file
// storage path.
func setupDBV0(t *testing.T, ctx context.Context, testName string) *sql.DB {
	thisTestFolder := tempFileStorage + testName

	// make temp data folder
	helpers_test.MakeTempStorage(t, thisTestFolder)

	// file will be new (empty)
	a := newFakeApp(t, thisTestFolder)

	db, cleanup, err := sqlite3.OpenSqlite3Database(a)
	if err != nil {
		t.Fatalf("failed to open sqlite3 db: %s", err)
	}
	t.Cleanup(cleanup)
	t.Cleanup(func() {
		err := db.Close()
		if err != nil {
			t.Errorf("failed to close storage (%s)", err)
		}
	})

	// setup empty db
	err = migrations.SetupGoose(ctx, db)
	if err != nil {
		t.Fatalf("goose setup failed: %s", err)
	}

	return db
}
