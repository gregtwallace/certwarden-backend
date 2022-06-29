package acme_accounts

import (
	"legocerthub-backend/pkg/acme"
)

// newAccountPayload() generates the payload for ACME to post to the
// new-account endpoint
func (account *Account) newAccountPayload() acme.NewAccountPayload {
	var contact []string

	if account.Email != "" {
		contact = append(contact, "mailto:"+account.Email)
	}

	return acme.NewAccountPayload{
		TosAgreed: account.AcceptedTos,
		Contact:   contact,
	}
}

// // updateLEAccount updates account settings with LE
// func (service *Service) updateLEAccount(payload AccountPayload) error {
// 	var acmeAccountResponse acme_utils.AcmeAccountResponse

// 	// fetch appropriate key
// 	keyPem, err := service.storage.GetAccountPem(*payload.ID)
// 	if err != nil {
// 		return err
// 	}

// 	// get kid
// 	kid, err := service.storage.GetAccountKid(*payload.ID)
// 	if err != nil {
// 		return err
// 	}

// 	acmeAccountResponse, err = service.updateLeAccount(payload, keyPem, kid)
// 	if err != nil {
// 		return err
// 	}

// 	// Write the returned account info from LE to the db
// 	err = service.storage.PutLEAccountInfo(*payload.ID, acmeAccountResponse)
// 	if err != nil {
// 		return err
// 	}

// 	return nil
// }
