package acme

import "crypto"

// RegisterAccount posts a secure message to the NewAccount URL of the directory
func (service *Service) RegisterAccount(payload any, uri string, privateKey crypto.PrivateKey) {
	var accountKey AccountKey

	// Register account should never use kid, it must always use JWK
	accountKey.Key = privateKey
	// set kid to nothing as this makes the post function use JWK
	// accountKey.Kid = ""
	// TODO: Remove this code, should not be needed

	service.postToUrlSigned(payload, service.dir.NewAccount, accountKey)

	// TODO: return response
}
