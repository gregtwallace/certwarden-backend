package private_keys

import (
	"context"
)

func (privateKeys *PrivateKeys) dbGetAllPrivateKeys() ([]*privateKey, error) {
	ctx, cancel := context.WithTimeout(context.Background(), privateKeys.DBTimeout)
	defer cancel()

	query := `SELECT id, name, description, algorithm, pem, api_key, created_at, updated_at
	FROM private_keys ORDER BY id`

	rows, err := privateKeys.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var allKeys []*privateKey
	for rows.Next() {
		var oneKey privateKey
		err = rows.Scan(
			&oneKey.ID,
			&oneKey.Name,
			&oneKey.Description,
			&oneKey.Algorithm,
			&oneKey.Pem,
			&oneKey.ApiKey,
			&oneKey.CreatedAt,
			&oneKey.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		allKeys = append(allKeys, &oneKey)
	}

	return allKeys, nil
}
