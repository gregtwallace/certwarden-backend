package directory

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"go.uber.org/zap"
	"golang.org/x/sync/semaphore"
)

// directory is the actual json directory object
type directory struct {
	NewNonce   string `json:"newNonce"`
	NewAccount string `json:"newAccount"`
	NewOrder   string `json:"newOrder"`
	NewAuthz   string `json:"newAuthz"`
	RevokeCert string `json:"revokeCert"`
	KeyChange  string `json:"keyChange"`
	Meta       struct {
		TermsOfService          string   `json:"termsOfService"`
		Website                 string   `json:"website"`
		CAAIdentities           []string `json:"caaIdentities"`
		ExternalAccountRequired bool     `json:"externalAccountRequired"`

		// ACME Profiles Extension - https://datatracker.ietf.org/doc/draft-aaron-acme-profiles/
		Profiles map[string]string `json:"profiles,omitempty"`
	} `json:"meta"`

	// ACME ARI (ACME Renewal Information) Extension - https://datatracker.ietf.org/doc/draft-ietf-acme-ari/
	RenewalInfo *string `json:"renewalInfo,omitempty"`

	// raw stores the directory URLs raw response
	raw json.RawMessage `json:"-"`
}

// DirectoryMgr struct manages an ACME directory, including retrieving the directory
// from its URI on a periodic basis
type DirectoryMgr struct {
	// management components
	logger      *zap.SugaredLogger
	httpClient  *http.Client
	dirURI      string
	nextFetchAt time.Time
	semHttp     *semaphore.Weighted // to prevent http fetch spam

	// the actual directory
	directoryObj directory
	muData       sync.RWMutex // to guard the directory data
}

// New returns a new ACME Directory
func New(logger *zap.SugaredLogger, httpClient *http.Client, dirURI string) *DirectoryMgr {
	dm := &DirectoryMgr{
		logger:     logger.Named(fmt.Sprintf("acme dir mgr %q", dirURI)),
		httpClient: httpClient,
		dirURI:     dirURI,
		semHttp:    semaphore.NewWeighted(1),
	}

	// prime the initial cache (hopefully)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	go func() {
		defer cancel()
		dm.updateDirectory(ctx)
	}()

	return dm
}
