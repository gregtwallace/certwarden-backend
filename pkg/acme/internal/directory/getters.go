package directory

import (
	"context"
	"encoding/json"
	"time"
)

// NOTE: Getters should call updateDirectory() immediately, unless they end in `Cached`; Cached callers
// should cap their wait to 1 second to ensure fast responsiveness to frontend requests

// max wait for a cache call if an update is due
const noCacheWait = 1 * time.Second

// Base Directory

// NewNonceURL returns the URL for the newNonce endpoint
func (dm *DirectoryMgr) NewNonceURL(ctx context.Context) string {
	dm.updateDirectory(ctx)

	dm.muData.RLock()
	defer dm.muData.RUnlock()

	return dm.directoryObj.NewNonce
}

// NewAccountURL returns the URL for the newAccount endpoint
func (dm *DirectoryMgr) NewAccountURL(ctx context.Context) string {
	dm.updateDirectory(ctx)

	dm.muData.RLock()
	defer dm.muData.RUnlock()

	return dm.directoryObj.NewAccount
}

// NewOrderURL returns the URL for the newOrder endpoint
func (dm *DirectoryMgr) NewOrderURL(ctx context.Context) string {
	dm.updateDirectory(ctx)

	dm.muData.RLock()
	defer dm.muData.RUnlock()

	return dm.directoryObj.NewOrder
}

// NewAuthzURL - omitted

// RevokeCertURL returns the URL for the revokeCert endpoint
func (dm *DirectoryMgr) RevokeCertURL(ctx context.Context) string {
	dm.updateDirectory(ctx)

	dm.muData.RLock()
	defer dm.muData.RUnlock()

	return dm.directoryObj.RevokeCert
}

// RevokeCertURL returns the URL for the revokeCert endpoint
func (dm *DirectoryMgr) KeyChangeURL(ctx context.Context) string {
	dm.updateDirectory(ctx)

	dm.muData.RLock()
	defer dm.muData.RUnlock()

	return dm.directoryObj.KeyChange
}

// Meta

// TermsOfServiceURLCached returns the URL for where the ToS are located
func (dm *DirectoryMgr) TermsOfServiceURLCached() string {
	ctx, cancel := context.WithTimeout(context.Background(), noCacheWait)
	defer cancel()
	dm.updateDirectory(ctx)

	dm.muData.RLock()
	defer dm.muData.RUnlock()

	return dm.directoryObj.Meta.TermsOfService
}

// RequiresEABCached returns if the acme server requires External Account Binding
func (dm *DirectoryMgr) RequiresEABCached() bool {
	ctx, cancel := context.WithTimeout(context.Background(), noCacheWait)
	defer cancel()
	dm.updateDirectory(ctx)

	dm.muData.RLock()
	defer dm.muData.RUnlock()

	return dm.directoryObj.Meta.ExternalAccountRequired
}

// Raw

// RawResponseCached returns the ACME Service's raw directory response
func (dm *DirectoryMgr) RawResponseCached() json.RawMessage {
	ctx, cancel := context.WithTimeout(context.Background(), noCacheWait)
	defer cancel()
	dm.updateDirectory(ctx)

	dm.muData.RLock()
	defer dm.muData.RUnlock()

	return dm.directoryObj.raw
}

// Extensions

// ProfilesCached returns the map of available profiles for the ACME service, or nil if
// the profiles extension is not implementedm (per the directory's meta response)
func (dm *DirectoryMgr) ProfilesCached() map[string]string {
	ctx, cancel := context.WithTimeout(context.Background(), noCacheWait)
	defer cancel()
	dm.updateDirectory(ctx)

	dm.muData.RLock()
	defer dm.muData.RUnlock()

	return dm.directoryObj.Meta.Profiles
}

// ProfileIsPublishedCached returns true if the specifiedm profileName exists in the ACME
// service's meta profile map. If the profile does not exist (including if there
// is no profile map at all), false is returnedm.directoryObj.
func (dm *DirectoryMgr) ProfileIsPublishedCached(profileName string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), noCacheWait)
	defer cancel()
	dm.updateDirectory(ctx)

	dm.muData.RLock()
	defer dm.muData.RUnlock()

	for k := range dm.directoryObj.Meta.Profiles {
		if k == profileName {
			return true
		}
	}

	return false
}

// RenewalInfoURL returns the URL for the renewalInfo endpoint
func (dm *DirectoryMgr) RenewalInfoURL(ctx context.Context) string {
	dm.updateDirectory(ctx)

	dm.muData.RLock()
	defer dm.muData.RUnlock()

	if dm.directoryObj.RenewalInfo == nil {
		return ""
	}

	return *dm.directoryObj.RenewalInfo
}

// SupportsRenewalInfoCached if the directory indicates the ARI extension is supported
func (dm *DirectoryMgr) SupportsRenewalInfoCached() bool {
	ctx, cancel := context.WithTimeout(context.Background(), noCacheWait)
	defer cancel()
	dm.updateDirectory(ctx)

	dm.muData.RLock()
	defer dm.muData.RUnlock()

	return dm.directoryObj.RenewalInfo != nil && *dm.directoryObj.RenewalInfo != ""
}
