package sqlite3

import (
	"bytes"
	"errors"
	"os"
	"testing"
)

// succesful migration
func TestMigrateDBFile1_MoveOldToNew(t *testing.T) {
	// make from data folder
	os.RemoveAll("./test-data-old-1")
	err := os.Mkdir("./test-data-old-1", os.FileMode(0755))
	if err != nil && !errors.Is(err, os.ErrExist) {
		t.Fatalf("failed to make dummy data folder (%s)", err)
	}
	defer os.RemoveAll("./test-data-old-1")

	// make to folder
	os.RemoveAll("./test-data-1")
	err = os.Mkdir("./test-data-1", os.FileMode(0755))
	if err != nil && !errors.Is(err, os.ErrExist) {
		t.Fatalf("failed to make dummy test-data folder (%s)", err)
	}
	defer os.RemoveAll("./test-data-1")

	// make fake "old" db file
	err = os.WriteFile("./test-data-old-1/appdata.db", []byte{'a', 'b', 'd'}, os.FileMode(0755))
	if err != nil && !errors.Is(err, os.ErrExist) {
		t.Fatalf("failed to make dummy test data file (%s)", err)
	}

	didMigrate, err := migrateDbFileLocation("./test-data-old-1/appdata.db", "./test-data-1/appdata.db")
	if err != nil {
		t.Fatalf("migration error (%s)", err)
	}
	if !didMigrate {
		t.Fatal("expected migration, but function didn't perform one")
	}

	migratedFileContent, err := os.ReadFile("./test-data-1/appdata.db")
	if err != nil {
		t.Fatalf("failed to read file, post-migration (%s)", err)
	}
	if !bytes.Equal(migratedFileContent, []byte{'a', 'b', 'd'}) {
		t.Fatal("file post-migration content did not match")
	}
}

// succesful non-migration
func TestMigrateDBFile2_OldDoesntExit(t *testing.T) {
	// ensure source doesnt exist
	os.RemoveAll("./test-data-old-2")

	didMigrate, err := migrateDbFileLocation("./test-data-old-2/appdata.db", "./test-data-2/appdata.db")
	if err != nil {
		t.Fatalf("migration error (%s)", err)
	}
	if didMigrate {
		t.Fatal("expected no migration, but function performed one")
	}
}

// invalid characters in source path/file
func TestMigrateDBFile3_Error1(t *testing.T) {
	// invalid from name, should cause stat error
	didMigrate, err := migrateDbFileLocation("\000x", "./test-data-3/appdata.db")
	if !errors.Is(err, errStatFromFailed) {
		t.Fatalf("migration expected stat error but got '%s'", err)
	}
	if didMigrate {
		t.Fatal("expected no migration, but function performed one")
	}
}

// invalid characters in destination path/file
func TestMigrateDBFile4_Error2(t *testing.T) {
	// make from data folder
	os.RemoveAll("./test-data-old-4")
	err := os.Mkdir("./test-data-old-4", os.FileMode(0755))
	if err != nil && !errors.Is(err, os.ErrExist) {
		t.Fatalf("failed to make dummy data folder (%s)", err)
	}
	defer os.RemoveAll("./test-data-old-4")

	// make fake "old" db file
	err = os.WriteFile("./test-data-old-4/appdata.db", []byte{'a', 'b', 'd'}, os.FileMode(0755))
	if err != nil && !errors.Is(err, os.ErrExist) {
		t.Fatalf("failed to make dummy test data file (%s)", err)
	}

	// invalid to name, should cause stat error
	didMigrate, err := migrateDbFileLocation("./test-data-old-4/appdata.db", "\000x")
	if !errors.Is(err, errStatToFailed) {
		t.Fatalf("migration expected stat error but got '%s'", err)
	}
	if didMigrate {
		t.Fatal("expected no migration, but function performed one")
	}
}

// non-existent path
func TestMigrateDBFile5_Error3(t *testing.T) {
	// make from data folder
	os.RemoveAll("./test-data-old-5")
	err := os.Mkdir("./test-data-old-5", os.FileMode(0755))
	if err != nil && !errors.Is(err, os.ErrExist) {
		t.Fatalf("failed to make dummy data folder (%s)", err)
	}
	defer os.RemoveAll("./test-data-old-5")

	// make fake "old" db file
	err = os.WriteFile("./test-data-old-5/appdata.db", []byte{'a', 'b', 'd'}, os.FileMode(0755))
	if err != nil && !errors.Is(err, os.ErrExist) {
		t.Fatalf("failed to make dummy test-data folder (%s)", err)
	}

	// migrate to non-existent dir
	didMigrate, err := migrateDbFileLocation("./test-data-old-5/appdata.db", "./test-data-5/appdata.db")
	if !errors.Is(err, errRenameFailed) {
		t.Fatalf("migration expected rename error but got '%s'", err)
	}
	if didMigrate {
		t.Fatal("expected no migration, but function performed one")
	}
}

// try to overwrite existing file
func TestMigrateDBFile6_Error4(t *testing.T) {
	// make from data folder
	os.RemoveAll("./test-data-old-6")
	err := os.Mkdir("./test-data-old-6", os.FileMode(0755))
	if err != nil && !errors.Is(err, os.ErrExist) {
		t.Fatalf("failed to make dummy data folder (%s)", err)
	}
	defer os.RemoveAll("./test-data-old-6")

	// make fake "old" db file
	err = os.WriteFile("./test-data-old-6/appdata.db", []byte{'a', 'b', 'd'}, os.FileMode(0755))
	if err != nil && !errors.Is(err, os.ErrExist) {
		t.Fatalf("failed to make dummy test-data folder (%s)", err)
	}

	// make to data folder
	os.RemoveAll("./test-data-6")
	err = os.Mkdir("./test-data-6", os.FileMode(0755))
	if err != nil && !errors.Is(err, os.ErrExist) {
		t.Fatalf("failed to make dummy data folder (%s)", err)
	}
	defer os.RemoveAll("./test-data-6")

	// make fake "new" db file
	err = os.WriteFile("./test-data-6/appdata.db", []byte{'1', '2', '4'}, os.FileMode(0755))
	if err != nil && !errors.Is(err, os.ErrExist) {
		t.Fatalf("failed to make dummy test-data folder (%s)", err)
	}

	// migrate to non-existent dir
	didMigrate, err := migrateDbFileLocation("./test-data-old-6/appdata.db", "./test-data-6/appdata.db")
	if !errors.Is(err, errStatAlreadyExists) {
		t.Fatalf("migration expected rename error but got '%s'", err)
	}
	if didMigrate {
		t.Fatal("expected no migration, but function performed one")
	}

	// validate file actually didn't get modified
	toFileContent, err := os.ReadFile("./test-data-6/appdata.db")
	if err != nil {
		t.Fatalf("failed to read toFile (%s)", err)
	}
	if !bytes.Equal(toFileContent, []byte{'1', '2', '4'}) {
		t.Fatal("toFile content was unexpectedly modified")
	}
}
