package acme

import (
	"encoding/json"
)

// Values related to the properties of the ACME Server / Service

// Meta

// TermsOfServiceURLCached returns the URL for where the ToS are located
func (s *Service) TermsOfServiceURLCached() string {
	return s.dir.TermsOfServiceURLCached()
}

// RequiresEAB returns if the acme server requires External Account Binding
func (s *Service) RequiresEABCached() bool {
	return s.dir.RequiresEABCached()
}

// Raw

// RawDirectoryCached returns the ACME Service's raw directory
func (s *Service) RawDirectoryCached() json.RawMessage {
	return s.dir.RawResponseCached()
}

// Extensions

// PublishedProfiles returns the map of available profiles for the ACME service, or nil if
// the profiles extension is not implemented (per the directory's meta response)
func (s *Service) PublishedProfilesCached() map[string]string {
	return s.dir.ProfilesCached()
}

// ProfileIsPublished retursn true if the specified profileName exists in the ACME
// service's directory's meta profile map. If the profile does not exist (including if there
// is no profile map at all), false is returned.
func (s *Service) ProfileIsPublishedCached(profileName string) bool {
	return s.dir.ProfileIsPublishedCached(profileName)
}

// SupportsARIExtensionCached returns true if the ACME Service supports ARI (ACME Renewal Info) extension
func (s *Service) SupportsARIExtensionCached() bool {
	return s.dir.SupportsRenewalInfoCached()
}
