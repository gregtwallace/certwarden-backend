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

// import (
// 	"legocerthub-backend/pkg/utils/acme_utils"
// )

// // Create account with LE
// func (service *Service) createLeAccount(payload AccountPayload, keyPem string) (acme_utils.AcmeAccountResponse, error) {
// 	// payload to sent to LE
// 	var acmeAccount acme_utils.AcmeAccount

// 	acmeAccount.TermsOfServiceAgreed = true
// 	if *payload.Email != "" {
// 		acmeAccount.Contact = []string{"mailto:" + *payload.Email}
// 	}

// 	// vars for return
// 	var acmeAccountResponse acme_utils.AcmeAccountResponse
// 	var err error

// 	if *payload.IsStaging == true {
// 		acmeAccountResponse, err = service.acmeStagingDir.CreateAccount(acmeAccount, keyPem)
// 		if err != nil {
// 			return acmeAccountResponse, err
// 		}
// 	} else {
// 		acmeAccountResponse, err = service.acmeProdDir.CreateAccount(acmeAccount, keyPem)
// 		if err != nil {
// 			return acmeAccountResponse, err
// 		}
// 	}

// 	return acmeAccountResponse, nil
// }

// // Create account with LE
// func (service *Service) updateLeAccount(payload AccountPayload, keyPem string, kid string) (acme_utils.AcmeAccountResponse, error) {
// 	// payload to sent to LE
// 	var acmeAccount acme_utils.AcmeAccount

// 	acmeAccount.TermsOfServiceAgreed = true
// 	if *payload.Email != "" {
// 		acmeAccount.Contact = []string{"mailto:" + *payload.Email}
// 	}

// 	// vars for return
// 	var acmeAccountResponse acme_utils.AcmeAccountResponse
// 	var err error

// 	if *payload.IsStaging == true {
// 		acmeAccountResponse, err = service.acmeStagingDir.UpdateAccount(acmeAccount, keyPem, kid)
// 		if err != nil {
// 			return acmeAccountResponse, err
// 		}
// 	} else {
// 		acmeAccountResponse, err = service.acmeProdDir.UpdateAccount(acmeAccount, keyPem, kid)
// 		if err != nil {
// 			return acmeAccountResponse, err
// 		}
// 	}

// 	return acmeAccountResponse, nil
// }
