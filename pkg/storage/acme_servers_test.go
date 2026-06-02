package storage_test

import (
	"certwarden-backend/pkg/domain/acme_servers"
	"testing"
)

// CompareAcmeServers compares key to expectedKey and throws appropriate errors for any differences
func CompareAcmeServer(t *testing.T, server, expectedServer acme_servers.Server) {
	if server.ID != expectedServer.ID {
		t.Errorf("acme server: id expected '%d' but got '%d'", expectedServer.ID, server.ID)
	}

	if server.Name != expectedServer.Name {
		t.Errorf("acme server: name expected '%s' but got '%s'", expectedServer.Name, server.Name)
	}

	if server.Description != expectedServer.Description {
		t.Errorf("acme server: description expected '%s' but got '%s'", expectedServer.Description, server.Description)
	}

	if server.DirectoryURL != expectedServer.DirectoryURL {
		t.Errorf("acme server: directory url expected '%s' but got '%s'", expectedServer.DirectoryURL, server.DirectoryURL)
	}

	if server.IsStaging != expectedServer.IsStaging {
		t.Errorf("acme server: is staging expected '%t' but got '%t'", expectedServer.IsStaging, server.IsStaging)
	}

	if !server.CreatedAt.Equal(expectedServer.CreatedAt) {
		t.Errorf("acme server: last access expected '%s' but got '%s'", expectedServer.CreatedAt.UTC(), server.CreatedAt.UTC())
	}

	if !server.UpdatedAt.Equal(expectedServer.UpdatedAt) {
		t.Errorf("acme server: last access expected '%s' but got '%s'", expectedServer.UpdatedAt.UTC(), server.UpdatedAt.UTC())
	}
}
