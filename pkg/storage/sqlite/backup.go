package sqlite

import "context"

// LockForBackup starts a sql transaction that aquires a SHARED (read only)
// lock on the db and returns a function to end that lock. File copy or other
// filesystem backup action should be performed between the two functions
func (store *Storage) LockDBForBackup() (unlockFunc func(), err error) {
	// start sql transaction
	// Do not use timeout, use background to ensure backup actions have all the
	// time they want to do the backup
	tx, err := store.db.BeginTx(context.Background(), nil)
	if err != nil {
		return nil, err
	}
	// no defer rollback, only call via unlockFunc

	// arbitrary select which will cause the SHARED lock to begin
	query := `
		SELECT '' from acme_accounts
	`

	_, err = tx.Exec(query)
	if err != nil {
		return nil, err
	}

	// make function to rollback the tx (remove the lock)
	unlockFunc = func() {
		_ = tx.Rollback()
	}

	return unlockFunc, nil
}
