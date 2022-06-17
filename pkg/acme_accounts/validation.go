package acme_accounts

import "legocerthub-backend/pkg/utils"

// IsIdExistingValid returns an error if not valid, nil if valid
// we'll generally assume the id is valid if >= 0
func (service *Service) isAccountIdExisting(idParam int, idPayload *int) error {
	// basic check
	err := utils.IsIdExistingMatch(idParam, idPayload)
	if err != nil {
		return err
	}

	// check id exists in storage
	_, err = service.storage.GetOneAccount(*idPayload)
	if err != nil {
		return err
	}

	return nil
}
