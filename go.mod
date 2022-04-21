module legocerthub-backend

go 1.18

require github.com/julienschmidt/httprouter v1.3.0
require github.com/mattn/go-sqlite3 v1.14.12

replace legocerthub-backend/acme_accounts => /acme_accounts
replace legocerthub-backend/app => /app
replace legocerthub-backend/utils => /utils
