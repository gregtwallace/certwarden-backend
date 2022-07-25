module legocerthub-backend

go 1.18

require github.com/julienschmidt/httprouter v1.3.0
require github.com/mattn/go-sqlite3 v1.14.12

require (
	go.uber.org/atomic v1.9.0 // indirect
	go.uber.org/multierr v1.8.0 // indirect
	go.uber.org/zap v1.21.0 // indirect
)

replace legocerthub-backend/pkg/acme => /pkg/acme
replace legocerthub-backend/pkg/acme/nonces => /pkg/acme/nonces
replace legocerthub-backend/pkg/acme/challenges/http01 => /pkg/acme/challenges/http01
replace legocerthub-backend/pkg/domain/acme_accounts => /pkg/domain/acme_accounts
replace legocerthub-backend/pkg/domain/app => /pkg/domain/app
replace legocerthub-backend/pkg/domain/certificates => /pkg/domain/certificates
replace legocerthub-backend/pkg/domain/certificates/challenges => /pkg/domain/certificates/challenges
replace legocerthub-backend/pkg/domain/private_keys => /pkg/domain/private_keys
replace legocerthub-backend/pkg/domain/private_keys/key_crypto => /pkg/domain/private_keys/key_crypto
replace legocerthub-backend/pkg/httpclient => /pkg/httpclient
replace legocerthub-backend/pkg/output => /pkg/output
replace legocerthub-backend/pkg/storage => /pkg/storage
replace legocerthub-backend/pkg/storage/sqlite => /pkg/storage/sqlite
replace legocerthub-backend/pkg/utils => /pkg/utils
replace legocerthub-backend/pkg/validation => /pkg/validation
