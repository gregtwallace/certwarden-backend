package private_keys

import (
	"context"
)

func (privateKeysDB *PrivateKeysDB) dbGetAllPrivateKeys() ([]*privateKey, error) {
	ctx, cancel := context.WithTimeout(context.Background(), privateKeysDB.Timeout)
	defer cancel()

	query := `SELECT id, name, description, algorithm
	FROM private_keys ORDER BY id`

	rows, err := privateKeysDB.Database.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var allKeys []*privateKey
	for rows.Next() {
		var oneKey sqlPrivateKey
		err = rows.Scan(
			&oneKey.ID,
			&oneKey.Name,
			&oneKey.Description,
			&oneKey.Algorithm,
		)
		if err != nil {
			return nil, err
		}
		allKeys = append(allKeys, oneKey.sqlToPrivateKey())
	}

	return allKeys, nil
}
