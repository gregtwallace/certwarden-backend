package acme_accounts

import (
	"context"
	"log"
)

func (acmeAccountsApp *AcmeAccountsApp) dbGetAllAcmeAccounts() ([]*acmeAccount, error) {
	ctx, cancel := context.WithTimeout(context.Background(), acmeAccountsApp.Timeout)
	defer cancel()

	query := `SELECT aa.id, aa.name, aa.description, pk.id, pk.name, aa.status, aa.email, aa.is_staging 
	FROM
		acme_accounts aa
		LEFT JOIN private_keys pk on (aa.private_key_id = pk.id)
	ORDER BY aa.id`

	rows, err := acmeAccountsApp.Database.QueryContext(ctx, query)
	if err != nil {
		log.Print(err)
		return nil, err
	}
	defer rows.Close()

	var allAccounts []*acmeAccount
	for rows.Next() {
		var oneAccount acmeAccountDb
		err = rows.Scan(
			&oneAccount.id,
			&oneAccount.name,
			&oneAccount.description,
			&oneAccount.privateKeyId,
			&oneAccount.privateKeyName,
			&oneAccount.status,
			&oneAccount.email,
			&oneAccount.isStaging,
		)
		if err != nil {
			log.Print(err)
			return nil, err
		}

		convertedAccount, err := oneAccount.acmeAccountDbToAcc()
		if err != nil {
			log.Print(err)
			return nil, err
		}

		allAccounts = append(allAccounts, convertedAccount)
	}

	return allAccounts, nil
}

func (acmeAccountsApp *AcmeAccountsApp) dbGetOneAcmeAccount(id int) (*acmeAccount, error) {
	ctx, cancel := context.WithTimeout(context.Background(), acmeAccountsApp.Timeout)
	defer cancel()

	query := `SELECT aa.id, aa.name, aa.description, pk.id, pk.name, aa.status, aa.email, aa.is_staging,
	aa.accepted_tos, aa.kid, aa.created_at, aa.updated_at
	FROM
		acme_accounts aa
		LEFT JOIN private_keys pk on (aa.private_key_id = pk.id)
	WHERE aa.id = $1
	ORDER BY aa.id`

	row := acmeAccountsApp.Database.QueryRowContext(ctx, query, id)

	var oneAccount acmeAccountDb
	err := row.Scan(
		&oneAccount.id,
		&oneAccount.name,
		&oneAccount.description,
		&oneAccount.privateKeyId,
		&oneAccount.privateKeyName,
		&oneAccount.status,
		&oneAccount.email,
		&oneAccount.isStaging,
		&oneAccount.acceptedTos,
		&oneAccount.kid,
		&oneAccount.createdAt,
		&oneAccount.updatedAt,
	)
	if err != nil {
		log.Print(err)
		return nil, err
	}

	convertedAccount, err := oneAccount.acmeAccountDbToAcc()
	if err != nil {
		return nil, err
	}

	return convertedAccount, nil
}

// type acmeAccount struct {
// 	ID             int    `json:"id"`
// 	Name           string `json:"name"`
// 	PrivateKeyID   int    `json:"private_key_id"`
// 	PrivateKeyName string `json:"private_key_name"` // comes from a join with key table
// 	Description    string `json:"description"`
// 	Status         string `json:"status"`
// 	Email          string `json:"email"`
// 	AcceptedTos    bool   `json:"accepted_tos,omitempty"`
// 	IsStaging      bool   `json:"is_staging"`
// 	CreatedAt      int    `json:"created_at,omitempty"`
// 	UpdatedAt      int    `json:"updated_at,omitempty"`
// 	Kid            string `json:"kid,omitempty"`
// }
