package acme_accounts

import (
	"context"
	"log"
)

func (acmeAccountsApp *AcmeAccountsApp) dbGetAllAcmeAccounts() ([]*acmeAccount, error) {
	ctx, cancel := context.WithTimeout(context.Background(), acmeAccountsApp.Timeout)
	defer cancel()

	query := `SELECT aa.id, aa.name, pk.id, pk.name, aa.description, aa.status, aa.email, aa.is_staging 
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
			&oneAccount.privateKeyId,
			&oneAccount.privateKeyName.String,
			&oneAccount.description.String,
			&oneAccount.status.String,
			&oneAccount.email.String,
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
