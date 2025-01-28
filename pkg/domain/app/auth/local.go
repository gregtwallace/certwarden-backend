package auth

// localExtraFuncs implements session manager's extraFuncs interface
type localExtraFuncs struct {
	dbUsername     string
	storageService Storage
}

// RefreshCheck for local users just queries the DB to confirm no-error
func (lef *localExtraFuncs) RefreshCheck() error {
	// get user must work
	_, err := lef.storageService.GetOneUserByName(lef.dbUsername)
	if err != nil {
		return err
	}

	return nil
}
