module legocerthub-backend

go 1.18

require github.com/julienschmidt/httprouter v1.3.0
require github.com/mattn/go-sqlite3 v1.14.12

replace legocerthub-backend/pkg/acme_accounts => /pkg/acme_accounts
replace legocerthub-backend/pkg/utils/acme_utils => /pkg/utils/acme_utils
replace legocerthub-backend/pkg/app => /pkg/app
replace legocerthub-backend/pkg/private_keys => /pkg/private_keys
replace legocerthub-backend/pkg/utils => /pkg/utils
