package directory

import (
	"context"
	"encoding/json"
)

// NOTE: All getters should call updateDirectory() immediately

// Base Directory

// NewNonceURL returns the URL for the newNonce endpoint
func (dm *DirectoryMgr) NewNonceURL(ctx context.Context) string {
	dm.updateDirectory(ctx)

	dm.mu.RLock()
	defer dm.mu.RUnlock()

	return dm.directoryObj.NewNonce
}

// NewAccountURL returns the URL for the newAccount endpoint
func (dm *DirectoryMgr) NewAccountURL(ctx context.Context) string {
	dm.updateDirectory(ctx)

	dm.mu.RLock()
	defer dm.mu.RUnlock()

	return dm.directoryObj.NewAccount
}

// NewOrderURL returns the URL for the newOrder endpoint
func (dm *DirectoryMgr) NewOrderURL(ctx context.Context) string {
	dm.updateDirectory(ctx)

	dm.mu.RLock()
	defer dm.mu.RUnlock()

	return dm.directoryObj.NewOrder
}

// NewAuthzURL - omitted

// RevokeCertURL returns the URL for the revokeCert endpoint
func (dm *DirectoryMgr) RevokeCertURL(ctx context.Context) string {
	dm.updateDirectory(ctx)

	dm.mu.RLock()
	defer dm.mu.RUnlock()

	return dm.directoryObj.RevokeCert
}

// RevokeCertURL returns the URL for the revokeCert endpoint
func (dm *DirectoryMgr) KeyChangeURL(ctx context.Context) string {
	dm.updateDirectory(ctx)

	dm.mu.RLock()
	defer dm.mu.RUnlock()

	return dm.directoryObj.KeyChange
}

// Meta

// TermsOfServiceURL returns the URL for where the ToS are located
func (dm *DirectoryMgr) TermsOfServiceURL(ctx context.Context) string {
	dm.updateDirectory(ctx)

	dm.mu.RLock()
	defer dm.mu.RUnlock()

	return dm.directoryObj.Meta.TermsOfService
}

// RequiresEAB returns if the acme server requires External Account Binding
func (dm *DirectoryMgr) RequiresEAB(ctx context.Context) bool {
	dm.updateDirectory(ctx)

	dm.mu.RLock()
	defer dm.mu.RUnlock()

	return dm.directoryObj.Meta.ExternalAccountRequired
}

// Raw

// DirectoryRawResponse returns the ACME Service's raw directory response
func (dm *DirectoryMgr) RawResponse(ctx context.Context) json.RawMessage {
	dm.updateDirectory(ctx)

	dm.mu.RLock()
	defer dm.mu.RUnlock()

	return dm.directoryObj.raw
}

// Extensions

// Profiles returns the map of available profiles for the ACME service, or nil if
// the profiles extension is not implementedm (per the directory's meta response)
func (dm *DirectoryMgr) Profiles(ctx context.Context) map[string]string {
	dm.updateDirectory(ctx)

	dm.mu.RLock()
	defer dm.mu.RUnlock()

	return dm.directoryObj.Meta.Profiles
}

// ProfileIsPublishedm returns true if the specifiedm profileName exists in the ACME
// service's meta profile map. If the profile does not exist (including if there
// is no profile map at all), false is returnedm.directoryObj.
func (dm *DirectoryMgr) ProfileIsPublished(ctx context.Context, profileName string) bool {
	dm.updateDirectory(ctx)

	dm.mu.RLock()
	defer dm.mu.RUnlock()

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

	dm.mu.RLock()
	defer dm.mu.RUnlock()

	if dm.directoryObj.RenewalInfo == nil {
		return ""
	}

	return *dm.directoryObj.RenewalInfo
}
