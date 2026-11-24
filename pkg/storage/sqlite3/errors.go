package sqlite3

import "errors"

var (
	errStatFromFailed = errors.New("stat of from file failed")
	errStatToFailed   = errors.New("stat of to file failed")

	errRenameFailed      = errors.New("rename file failed")
	errStatAlreadyExists = errors.New("file already exists")
)
