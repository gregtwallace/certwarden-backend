module legocerthub-backend

go 1.19

require github.com/julienschmidt/httprouter v1.3.0

require github.com/mattn/go-sqlite3 v1.14.12

require (
	github.com/cloudflare/cloudflare-go v0.55.0 // indirect
	github.com/golang-jwt/jwt/v4 v4.4.2 // indirect
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/go-retryablehttp v0.7.2 // indirect
	github.com/natefinch/lumberjack v2.0.0+incompatible // indirect
	go.uber.org/atomic v1.9.0 // indirect
	go.uber.org/multierr v1.8.0 // indirect
	go.uber.org/zap v1.21.0 // indirect
	golang.org/x/crypto v0.7.0 // indirect
	golang.org/x/net v0.8.0 // indirect
	golang.org/x/text v0.8.0 // indirect
	golang.org/x/time v0.3.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace legocerthub-backend/pkg/acme => /pkg/acme

replace legocerthub-backend/pkg/acme/nonces => /pkg/acme/nonces

replace legocerthub-backend/pkg/challenges => /pkg/challenges

replace legocerthub-backend/pkg/challenges/dns_checker => /pkg/challenges/dns_checker

replace legocerthub-backend/pkg/challenges/providers/dns01acmedns => /pkg/challenges/providers/dns01acmedns

replace legocerthub-backend/pkg/challenges/providers/dns01acmesh => /pkg/challenges/providers/dns01acmesh

replace legocerthub-backend/pkg/challenges/providers/dns01cloudflare => /pkg/challenges/providers/dns01cloudflare

replace legocerthub-backend/pkg/challenges/providers/dns01manual => /pkg/challenges/providers/dns01manual

replace legocerthub-backend/pkg/challenges/providers/http01internal => /pkg/challenges/providers/http01internal

replace legocerthub-backend/pkg/datatypes => /pkg/datatypes

replace legocerthub-backend/pkg/domain/acme_accounts => /pkg/domain/acme_accounts

replace legocerthub-backend/pkg/domain/acme_servers => /pkg/domain/acme_servers

replace legocerthub-backend/pkg/domain/app => /pkg/domain/app

replace legocerthub-backend/pkg/domain/app/auth => /pkg/domain/app/auth

replace legocerthub-backend/pkg/domain/app/updater => /pkg/domain/app/updater

replace legocerthub-backend/pkg/domain/authorizations => /pkg/domain/authorizations

replace legocerthub-backend/pkg/domain/certificates => /pkg/domain/certificates

replace legocerthub-backend/pkg/domain/orders => /pkg/domain/orders

replace legocerthub-backend/pkg/domain/private_keys => /pkg/domain/private_keys

replace legocerthub-backend/pkg/domain/private_keys/key_crypto => /pkg/domain/private_keys/key_crypto

replace legocerthub-backend/pkg/httpclient => /pkg/httpclient

replace legocerthub-backend/pkg/output => /pkg/output

replace legocerthub-backend/pkg/pagination_sort => /pkg/pagination_sort

replace legocerthub-backend/pkg/storage => /pkg/storage

replace legocerthub-backend/pkg/storage/sqlite => /pkg/storage/sqlite

replace legocerthub-backend/pkg/randomness => /pkg/randomness

replace legocerthub-backend/pkg/validation => /pkg/validation
