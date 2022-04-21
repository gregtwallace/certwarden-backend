package acme_accounts

import (
	"context"
	"log"
	"time"
)

func (acmeAccounts *AcmeAccounts) DBGetAllAcmeAccounts() ([]*AcmeAccount, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `SELECT aa.id, pk.id, pk.name, aa.name, aa.description, aa.email, aa.accepted_tos, aa.is_staging 
	FROM
		acme_accounts aa
		LEFT JOIN private_keys pk on (aa.private_key_id = pk.id)
	ORDER BY aa.id`

	rows, err := acmeAccounts.DB.QueryContext(ctx, query)
	if err != nil {
		log.Print(err)
		return nil, err
	}
	defer rows.Close()

	var allAccounts []*AcmeAccount
	for rows.Next() {
		var oneAccount AcmeAccount
		err := rows.Scan(
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
