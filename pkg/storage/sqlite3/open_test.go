package sqlite3

import (
	"bytes"
	"context"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

type fakeApp struct {
	logger   *zap.SugaredLogger
	dataPath string
}

func (fa *fakeApp) GetLogger() *zap.SugaredLogger {
	return fa.logger
}

func (fa *fakeApp) GetDataStorageAppDataPath() string {
	return fa.dataPath
}

func newFakeApp(t *testing.T, dataPath string) *fakeApp {
	logger := zaptest.NewLogger(t, zaptest.Level(zap.FatalLevel)).Sugar() // use fatal to avoid log output

	return &fakeApp{
		logger:   logger,
		dataPath: dataPath,
	}
}

// TESTS

// invalid characters in path/file
func TestOpenDB1_Error1(t *testing.T) {
	fakeApp := newFakeApp(t, "\000x")

	// invalid from name, should cause stat error
	_, _, _, err := OpenSqlite3Database(fakeApp)
	if !errors.Is(err, errStatToFailed) {
		t.Fatalf("open expected stat error of toFile but got '%s'", err)
	}
}

// non-existent path
func TestOpenDB2_Error2(t *testing.T) {
	fakeApp := newFakeApp(t, "./test-opendb2/non-exist")

	// invalid from name, should cause stat error
	_, _, _, err := OpenSqlite3Database(fakeApp)
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("open expected error due to non-existent toFile path but got '%s'", err)
	}
}

// shouldnt overwrite if both files exist, but should still work with 'new' file
func TestOpenDB3_BothExist(t *testing.T) {
	fakeApp := newFakeApp(t, "./test-opendb3")

	// make old data folder and file
	os.RemoveAll("./data")
	err := os.Mkdir("./data", os.FileMode(0755))
	if err != nil && !errors.Is(err, os.ErrExist) {
		t.Fatalf("failed to make dummy data folder (%s)", err)
	}
	defer os.RemoveAll("./data")

	err = os.WriteFile("./data/appdata.db", []byte{'a', 'b', 'd'}, os.FileMode(0755))
	if err != nil && !errors.Is(err, os.ErrExist) {
		t.Fatalf("failed to make dummy test data file (%s)", err)
	}

	// make current data folder and file
	os.RemoveAll("./test-opendb3")
	err = os.Mkdir("./test-opendb3", os.FileMode(0755))
	if err != nil && !errors.Is(err, os.ErrExist) {
		t.Fatalf("failed to make dummy data folder (%s)", err)
	}
	defer os.RemoveAll("./test-opendb3")

	err = os.WriteFile("./test-opendb3/appdata.db", []byte{'1', '2', '4'}, os.FileMode(0755))
	if err != nil && !errors.Is(err, os.ErrExist) {
		t.Fatalf("failed to make dummy test data file (%s)", err)
	}

	// do migration
	db, isNew, onErrCleanup, err := OpenSqlite3Database(fakeApp)
	if err != nil {
		t.Fatalf("migration error '%s' but expected none", err)
	}
	defer onErrCleanup()

	// shouldnt be new
	if isNew {
		t.Fatal("migration reported new db, but shouldnt be new")
	}

	// both exist, ensure new file wasn't overwritten
	toFileContent, err := os.ReadFile("./test-opendb3/appdata.db")
	if err != nil {
		t.Fatalf("failed to read toFile (%s)", err)
	}
	if !bytes.Equal(toFileContent, []byte{'1', '2', '4'}) {
		t.Fatal("toFile content was unexpectedly modified")
	}

	// rewrite db file so it is a valid db
	err = os.WriteFile("./test-opendb3/appdata.db", []byte{}, os.FileMode(0755))
	if err != nil {
		t.Fatalf("failed to re-write db (%s)", err)
	}

	// ensure db connection is usable
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		t.Fatalf("failed to ping db after opening (%s)", err)
	}

	// ensure onErrCleanup doesnt delete the db
	onErrCleanup()
	toFileContent, err = os.ReadFile("./test-opendb3/appdata.db")
	if err != nil {
		t.Fatalf("failed to read toFile (%s)", err)
	}
	if !bytes.Equal(toFileContent, []byte{}) {
		t.Fatal("toFile content was unexpectedly modified")
	}

	// ensure onErrCleanup closed the db
	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err == nil || !strings.Contains(err.Error(), "database is closed") {
		t.Fatalf("expected database closed error but got '%s'", err)
	}
}

// open existing in new location (old doesnt exist)
func TestOpenDB4_OldPathDoesntExist(t *testing.T) {
	fakeApp := newFakeApp(t, "./test-opendb4")

	// remove old path if exists
	os.RemoveAll("./data")

	// make current data folder & file
	os.RemoveAll("./test-opendb4")
	err := os.Mkdir("./test-opendb4", os.FileMode(0755))
	if err != nil && !errors.Is(err, os.ErrExist) {
		t.Fatalf("failed to make dummy data folder (%s)", err)
	}
	defer os.RemoveAll("./test-opendb4")

	err = os.WriteFile("./test-opendb4/appdata.db", []byte{'1', '2', '4'}, os.FileMode(0755))
	if err != nil && !errors.Is(err, os.ErrExist) {
		t.Fatalf("failed to make dummy test data file (%s)", err)
	}

	// do migration
	db, isNew, onErrCleanup, err := OpenSqlite3Database(fakeApp)
	if err != nil {
		t.Fatalf("migration error '%s' but expected none", err)
	}
	defer onErrCleanup()

	// shouldnt be new
	if isNew {
		t.Fatal("migration reported new db, but shouldnt be new")
	}

	// ensure file wasn't modified
	toFileContent, err := os.ReadFile("./test-opendb4/appdata.db")
	if err != nil {
		t.Fatalf("failed to read toFile (%s)", err)
	}
	if !bytes.Equal(toFileContent, []byte{'1', '2', '4'}) {
		t.Fatal("toFile content was unexpectedly modified")
	}

	// rewrite db file so it is a valid db
	err = os.WriteFile("./test-opendb4/appdata.db", []byte{}, os.FileMode(0755))
	if err != nil {
		t.Fatalf("failed to re-write db (%s)", err)
	}

	// ensure db connection is usable
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		t.Fatalf("failed to ping db after opening (%s)", err)
	}

	// ensure onErrCleanup doesnt delete the db
	onErrCleanup()
	toFileContent, err = os.ReadFile("./test-opendb4/appdata.db")
	if err != nil {
		t.Fatalf("failed to read toFile (%s)", err)
	}
	if !bytes.Equal(toFileContent, []byte{}) {
		t.Fatal("toFile content was unexpectedly modified")
	}

	// ensure onErrCleanup closed the db
	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err == nil || !strings.Contains(err.Error(), "database is closed") {
		t.Fatalf("expected database closed error but got '%s'", err)
	}
}

// migrate from old location (only old exists)
func TestOpenDB5_OldPathExistsNewDoesnt(t *testing.T) {
	fakeApp := newFakeApp(t, "./test-opendb5")

	// make old
	os.RemoveAll("./data")
	err := os.Mkdir("./data", os.FileMode(0755))
	if err != nil && !errors.Is(err, os.ErrExist) {
		t.Fatalf("failed to make dummy data folder (%s)", err)
	}
	defer os.RemoveAll("./data")

	err = os.WriteFile("./data/appdata.db", []byte{'z', 'a'}, os.FileMode(0755))
	if err != nil && !errors.Is(err, os.ErrExist) {
		t.Fatalf("failed to make dummy test data file (%s)", err)
	}

	// make new path
	os.RemoveAll("./test-opendb5")
	err = os.Mkdir("./test-opendb5", os.FileMode(0755))
	if err != nil && !errors.Is(err, os.ErrExist) {
		t.Fatalf("failed to make dummy data folder (%s)", err)
	}
	defer os.RemoveAll("./test-opendb5")

	// do migration
	db, isNew, onErrCleanup, err := OpenSqlite3Database(fakeApp)
	if err != nil {
		t.Fatalf("migration error '%s' but expected none", err)
	}
	defer onErrCleanup()

	// shouldnt be new
	if isNew {
		t.Fatal("migration reported new db, but shouldnt be new")
	}

	// ensure file was moved but wasn't modified
	toFileContent, err := os.ReadFile("./test-opendb5/appdata.db")
	if err != nil {
		t.Fatalf("failed to read toFile (%s)", err)
	}
	if !bytes.Equal(toFileContent, []byte{'z', 'a'}) {
		t.Fatal("toFile content was unexpectedly modified")
	}

	// ensure file from 'old' location is gone
	_, err = os.Stat("./data/appdata.db")
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("fromFile expected does not exist error but got '%s'", err)
	}

	// rewrite db file so it is a valid db
	err = os.WriteFile("./test-opendb5/appdata.db", []byte{}, os.FileMode(0755))
	if err != nil {
		t.Fatalf("failed to re-write db (%s)", err)
	}

	// ensure db connection is usable
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		t.Fatalf("failed to ping db after opening (%s)", err)
	}

	// ensure onErrCleanup doesnt delete the db
	onErrCleanup()
	toFileContent, err = os.ReadFile("./test-opendb5/appdata.db")
	if err != nil {
		t.Fatalf("failed to read toFile (%s)", err)
	}
	if !bytes.Equal(toFileContent, []byte{}) {
		t.Fatal("toFile content was unexpectedly modified")
	}

	// ensure onErrCleanup closed the db
	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err == nil || !strings.Contains(err.Error(), "database is closed") {
		t.Fatalf("expected database closed error but got '%s'", err)
	}
}

// neither exist -- setup new
func TestOpenDB6_NoFilesExist(t *testing.T) {
	fakeApp := newFakeApp(t, "./test-opendb6")

	// make old path
	os.RemoveAll("./data")
	err := os.Mkdir("./data", os.FileMode(0755))
	if err != nil && !errors.Is(err, os.ErrExist) {
		t.Fatalf("failed to make dummy data folder (%s)", err)
	}
	defer os.RemoveAll("./data")

	// make new path
	os.RemoveAll("./test-opendb6")
	err = os.Mkdir("./test-opendb6", os.FileMode(0755))
	if err != nil && !errors.Is(err, os.ErrExist) {
		t.Fatalf("failed to make dummy data folder (%s)", err)
	}
	defer os.RemoveAll("./test-opendb6")

	// do migration
	db, isNew, onErrCleanup, err := OpenSqlite3Database(fakeApp)
	if err != nil {
		t.Fatalf("migration error '%s' but expected none", err)
	}
	defer onErrCleanup()

	// should be new
	if !isNew {
		t.Fatal("migration reported not-new db, but should be new")
	}

	// ensure file was created
	_, err = os.Stat("./test-opendb6/appdata.db")
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			t.Fatal("new db file doesn't exist (but was expected)")
		}
		t.Fatalf("stat error trying to check for new db file (%s)", err)
	}

	// ensure db connection is usable
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		t.Fatalf("failed to ping db after opening (%s)", err)
	}

	// ensure onErrCleanup deleted the db
	onErrCleanup()
	_, err = os.Stat("./test-opendb6/appdata.db")
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("stat expected new db doesnt exist error but got '%s'", err)
	}

	// ensure onErrCleanup closed the db
	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err == nil || !strings.Contains(err.Error(), "database is closed") {
		t.Fatalf("expected database closed error but got '%s'", err)
	}
}
