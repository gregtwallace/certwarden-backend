package acme_accounts

// acme account payload from PUT/POST
type AccountPayload struct {
	ID           *int    `json:"id"`
	Name         *string `json:"name"`
	Description  *string `json:"description"`
	Email        *string `json:"email"`
	PrivateKeyID *int    `json:"private_key_id"`
	AcceptedTos  *bool   `json:"accepted_tos"`
	IsStaging    *bool   `json:"is_staging"`
}
