module legocerthub-backend

go 1.18

require github.com/julienschmidt/httprouter v1.3.0
require github.com/mattn/go-sqlite3 v1.14.12

replace legocerthub-backend/pkg/domain/acme_accounts => /pkg/domain/acme_accounts
replace legocerthub-backend/pkg/domain/app => /pkg/domain/app
replace legocerthub-backend/pkg/domain/private_keys => /pkg/domain/private_keys
replace legocerthub-backend/pkg/domain/private_keys/key_crypto => /pkg/domain/private_keys/key_crypto
replace legocerthub-backend/pkg/storage/sqlite => /pkg/storage/sqlite
replace legocerthub-backend/pkg/utils => /pkg/utils
replace legocerthub-backend/pkg/utils/acme_utils => /pkg/utils/acme_utils
