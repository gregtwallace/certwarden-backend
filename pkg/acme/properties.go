package acme

import (
	"context"
	"encoding/json"
)

// Values related to the properties of the ACME Server / Service

// Meta

// TermsOfServiceURL returns the URL for where the ToS are located
func (s *Service) TermsOfServiceURL(ctx context.Context) string {
	return s.dir.TermsOfServiceURL(ctx)
}

// RequiresEAB returns if the acme server requires External Account Binding
func (s *Service) RequiresEAB(ctx context.Context) bool {
	return s.dir.RequiresEAB(ctx)
}

// Raw

// DirectoryRawResponse returns the ACME Service's raw directory
func (s *Service) RawDirectory(ctx context.Context) json.RawMessage {
	return s.dir.RawResponse(ctx)
}

// Extensions

// PublishedProfiles returns the map of available profiles for the ACME service, or nil if
// the profiles extension is not implemented (per the directory's meta response)
func (s *Service) PublishedProfiles(ctx context.Context) map[string]string {
	return s.dir.Profiles(ctx)
}

// ProfileIsPublished returns true if the specified profileName exists in the ACME
// service's directory's meta profile map. If the profile does not exist (including if there
// is no profile map at all), false is returned.
func (s *Service) ProfileIsPublished(ctx context.Context, profileName string) bool {
	return s.dir.ProfileIsPublished(ctx, profileName)
}

// SupportsARIExtension returns true if the ACME Service supports ARI (ACME Renewal Info) extension
func (s *Service) SupportsARIExtension(ctx context.Context) bool {
	return s.dir.RenewalInfoURL(ctx) != ""
}
