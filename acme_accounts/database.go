package acme_accounts

import (
	"context"
	"log"
)

func (acmeAccountsApp *AcmeAccountsApp) dbGetAllAcmeAccounts() ([]*acmeAccount, error) {
	ctx, cancel := context.WithTimeout(context.Background(), acmeAccountsApp.Timeout)
	defer cancel()

	query := `SELECT aa.id, pk.id, pk.name, aa.name, aa.description, aa.email, aa.accepted_tos, aa.is_staging 
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
		var oneAccount acmeAccount
		err = rows.Scan(
			&oneAccount.ID,
			&oneAccount.PrivateKeyID,
			&oneAccount.PrivateKeyName,
			&oneAccount.Name,
			&oneAccount.Description,
			&oneAccount.Email,
			&oneAccount.AcceptedTos,
			&oneAccount.IsStaging,
		)
		if err != nil {
			log.Print(err)
			return nil, err
		}
		// TO DO join the key info
		allAccounts = append(allAccounts, &oneAccount)
	}

	return allAccounts, nil
}