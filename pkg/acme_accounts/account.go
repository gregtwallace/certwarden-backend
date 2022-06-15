package acme_accounts

import (
	"legocerthub-backend/pkg/private_keys"
)

// a single ACME Account
type Account struct {
	ID             int    `json:"id"`
	Name           string `json:"name"`
	Description    string `json:"description"`
	PrivateKeyID   int    `json:"private_key_id"`
	PrivateKeyName string `json:"private_key_name"` // comes from a join with key table
	Status         string `json:"status"`
	Email          string `json:"email"`
	AcceptedTos    bool   `json:"accepted_tos,omitempty"`
	IsStaging      bool   `json:"is_staging"`
	CreatedAt      int    `json:"created_at,omitempty"`
	UpdatedAt      int    `json:"updated_at,omitempty"`
	Kid            string `json:"kid,omitempty"`
}

// acme account payload from PUT/POST
type AccountPayload struct {
	ID           *int   `json:"id"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	Email        string `json:"email"`
	PrivateKeyID *int   `json:"private_key_id"`
	AcceptedTos  bool   `json:"accepted_tos"`
	IsStaging    bool   `json:"is_staging"`
}

// acme account payload from an html form (which only supports strings)
type accountHtmlPayload struct {
	ID           *string `json:"id"`
	Name         string  `json:"name"`
	Description  string  `json:"description"`
	Email        string  `json:"email"`
	PrivateKeyID *string `json:"private_key_id"`
	AcceptedTos  string  `json:"accepted_tos"`
	IsStaging    string  `json:"is_staging"`
}

// new account info
// used to return info about valid options when making a new account
type newAccountOptions struct {
	TosUrl        string             `json:"tos_url"`
	StagingTosUrl string             `json:"staging_tos_url"`
	AvailableKeys []private_keys.Key `json:"available_keys"`
}
