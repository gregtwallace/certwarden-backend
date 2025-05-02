package providers

import (
	"certwarden-backend/pkg/challenges/providers/dns01acmedns"
	"certwarden-backend/pkg/challenges/providers/dns01acmesh"
	"certwarden-backend/pkg/challenges/providers/dns01cloudflare"
	"certwarden-backend/pkg/challenges/providers/dns01goacme"
	"certwarden-backend/pkg/challenges/providers/dns01manual"
	"certwarden-backend/pkg/challenges/providers/http01internal"
	"certwarden-backend/pkg/output"
	"encoding/json"
	"fmt"
	"net/http"
)

// newPayload is used to add a provider
type newPayload struct {
	// mandatory
	Domains []string `json:"domains"`

	// optional
	PreCheckWaitSeconds  *int `json:"precheck_wait"`
	PostCheckWaitSeconds *int `json:"postcheck_wait"`

	// + mandatory, only one of these
	Http01InternalConfig  *http01internal.Config  `json:"http_01_internal,omitempty"`
	Dns01ManualConfig     *dns01manual.Config     `json:"dns_01_manual,omitempty"`
	Dns01AcmeDnsConfig    *dns01acmedns.Config    `json:"dns_01_acme_dns,omitempty"`
	Dns01AcmeShConfig     *dns01acmesh.Config     `json:"dns_01_acme_sh,omitempty"`
	Dns01CloudflareConfig *dns01cloudflare.Config `json:"dns_01_cloudflare,omitempty"`
	Dns01GoAcmeConfig     *dns01goacme.Config     `json:"dns_01_go_acme,omitempty"`
}

// CreateProvider creates a new provider using the specified configuration.
func (mgr *Manager) CreateProvider(w http.ResponseWriter, r *http.Request) *output.JsonError {
	mgr.mu.Lock()
	defer mgr.mu.Unlock()

	// decode body into payload
	var payload newPayload
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		mgr.logger.Debug(err)
		return output.JsonErrValidationFailed(err)
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
	if payload.Dns01GoAcmeConfig != nil {
		configCount++
	}
	if configCount != 1 {
		err = fmt.Errorf("new provider expects 1 config, received %d", configCount)
		mgr.logger.Debug(err)
		return output.JsonErrValidationFailed(err)
	}

	// make internal config
	internalCfg := InternalConfig{
		Domains: payload.Domains,
	}
	if payload.PreCheckWaitSeconds != nil {
		internalCfg.PreCheckWaitSeconds = *payload.PreCheckWaitSeconds
	} else {
		internalCfg.PreCheckWaitSeconds = 3 * 60
	}
	if payload.PostCheckWaitSeconds != nil {
		internalCfg.PostCheckWaitSeconds = *payload.PostCheckWaitSeconds
	} else {
		//internalCfg.PostCheckWaitSeconds = 0
	}

	// try to add the specified provider (actual action)
	var p *provider
	if payload.Http01InternalConfig != nil {
		p, err = mgr.unsafeAddProvider(internalCfg, payload.Http01InternalConfig)

	} else if payload.Dns01ManualConfig != nil {
		p, err = mgr.unsafeAddProvider(internalCfg, payload.Dns01ManualConfig)

	} else if payload.Dns01AcmeDnsConfig != nil {
		p, err = mgr.unsafeAddProvider(internalCfg, payload.Dns01AcmeDnsConfig)

	} else if payload.Dns01AcmeShConfig != nil {
		p, err = mgr.unsafeAddProvider(internalCfg, payload.Dns01AcmeShConfig)

	} else if payload.Dns01CloudflareConfig != nil {
		p, err = mgr.unsafeAddProvider(internalCfg, payload.Dns01CloudflareConfig)

	} else if payload.Dns01GoAcmeConfig != nil {
		p, err = mgr.unsafeAddProvider(internalCfg, payload.Dns01GoAcmeConfig)

	} else {
		mgr.logger.Error("new provider cfg missing, this error should never trigger though, report bug to developer")
	}

	// common err check
	if err != nil {
		err = fmt.Errorf("failed to add new provider (%s)", err)
		mgr.logger.Debug(err)
		return output.JsonErrValidationFailed(err)
	}

	// update config file
	err = mgr.unsafeWriteProvidersConfig()
	if err != nil {
		err = fmt.Errorf("failed to save config file after providers update (%s)", err)
		mgr.logger.Error(err)
		return output.JsonErrInternal(err)
	}

	// write response
	response := &providerResponse{}
	response.StatusCode = http.StatusCreated
	response.Message = "created provider"
	response.Provider = p

	err = mgr.output.WriteJSON(w, response)
	if err != nil {
		mgr.logger.Errorf("failed to write json (%s)", err)
		return output.JsonErrWriteJsonError(err)
	}

	return nil
}
