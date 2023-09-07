package providers

import (
	"encoding/json"
	"legocerthub-backend/pkg/challenges/providers/dns01acmedns"
	"legocerthub-backend/pkg/challenges/providers/dns01acmesh"
	"legocerthub-backend/pkg/challenges/providers/dns01cloudflare"
	"legocerthub-backend/pkg/challenges/providers/dns01manual"
	"legocerthub-backend/pkg/challenges/providers/http01internal"
	"legocerthub-backend/pkg/output"
	"net/http"
)

// newPayload is used to add a provider
type newPayload struct {
	Http01InternalConfig  *http01internal.Config  `json:"http_01_internal,omitempty"`
	Dns01ManualConfig     *dns01manual.Config     `json:"dns_01_manual,omitempty"`
	Dns01AcmeDnsConfig    *dns01acmedns.Config    `json:"dns_01_acme_dns,omitempty"`
	Dns01AcmeShConfig     *dns01acmesh.Config     `json:"dns_01_acme_sh,omitempty"`
	Dns01CloudflareConfig *dns01cloudflare.Config `json:"dns_01_cloudflare,omitempty"`
}

// CreateProvider creates a new provider using the specified configuration.
func (mgr *Manager) CreateProvider(w http.ResponseWriter, r *http.Request) (err error) {
	mgr.mu.Lock()
	defer mgr.mu.Unlock()

	// decode body into payload
	var payload newPayload
	err = json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		mgr.logger.Debug(err)
		return output.ErrValidationFailed
	}

	// verify correct number of configs received
	configCount := 0
	if payload.Http01InternalConfig != nil {
		configCount++
	}
	if payload.Dns01ManualConfig != nil {
		configCount++
	}
	if payload.Dns01AcmeDnsConfig != nil {
		configCount++
	}
	if payload.Dns01AcmeShConfig != nil {
		configCount++
	}
	if payload.Dns01CloudflareConfig != nil {
		configCount++
	}
	if configCount != 1 {
		mgr.logger.Debugf("new provider expects 1 config, received %d", configCount)
		return output.ErrValidationFailed
	}

	// try to add the specified provider (actual action)
	var id int
	if payload.Http01InternalConfig != nil {
		id, err = mgr.unsafeAddProvider(payload.Http01InternalConfig)

	} else if payload.Dns01ManualConfig != nil {
		id, err = mgr.unsafeAddProvider(payload.Dns01ManualConfig)

	} else if payload.Dns01AcmeDnsConfig != nil {
		id, err = mgr.unsafeAddProvider(payload.Dns01AcmeDnsConfig)

	} else if payload.Dns01AcmeShConfig != nil {
		id, err = mgr.unsafeAddProvider(payload.Dns01AcmeShConfig)

	} else if payload.Dns01CloudflareConfig != nil {
		id, err = mgr.unsafeAddProvider(payload.Dns01CloudflareConfig)

	} else {
		mgr.logger.Error("new provider cfg missing, this error should never trigger though, report lego bug")
	}

	// common err check
	if err != nil {
		mgr.logger.Debugf("failed to add new provider (%s)", err)
		return output.ErrValidationFailed
	}

	// update config file
	err = mgr.unsafeWriteProvidersConfig()
	if err != nil {
		mgr.logger.Errorf("failed to save config file after providers update (%s)", err)
		return output.ErrInternal
	}

	// return response to client
	response := output.JsonResponse{
		Status:  http.StatusCreated,
		Message: "created",
		ID:      id,
	}

	err = mgr.output.WriteJSON(w, response.Status, response, "response")
	if err != nil {
		return err
	}

	return nil
}
