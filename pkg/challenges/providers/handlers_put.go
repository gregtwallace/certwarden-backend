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
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

// modifyPayload is the needed payload to update an existing provider
type modifyPayload struct {
	// mandatory
	ID  int    `json:"-"`
	Tag string `json:"tag"`

	// optional
	Domains                  []string `json:"domains,omitempty"`
	PostProvisionWaitSeconds *int     `json:"post_resource_provision_wait"`

	// plus only one of these
	Http01InternalConfig  *http01internal.Config  `json:"http_01_internal,omitempty"`
	Dns01ManualConfig     *dns01manual.Config     `json:"dns_01_manual,omitempty"`
	Dns01AcmeDnsConfig    *dns01acmedns.Config    `json:"dns_01_acme_dns,omitempty"`
	Dns01AcmeShConfig     *dns01acmesh.Config     `json:"dns_01_acme_sh,omitempty"`
	Dns01CloudflareConfig *dns01cloudflare.Config `json:"dns_01_cloudflare,omitempty"`
	Dns01GoAcmeConfig     *dns01goacme.Config     `json:"dns_01_go_acme,omitempty"`
}

// ModifyProvider modifies the provider specified by the ID in manager with the specified
// configuration.
func (mgr *Manager) ModifyProvider(w http.ResponseWriter, r *http.Request) *output.JsonError {
	mgr.mu.Lock()
	defer mgr.mu.Unlock()

	// decode body into payload
	var payload modifyPayload
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		mgr.logger.Debug(err)
		return output.JsonErrValidationFailed(err)
	}

	// params
	idParam := httprouter.ParamsFromContext(r.Context()).ByName("id")
	payload.ID, err = strconv.Atoi(idParam)
	if err != nil {
		mgr.logger.Debug(err)
		return output.JsonErrValidationFailed(err)
	}

	// find provider
	p := (*provider)(nil)
	for _, oneP := range mgr.providers {
		if oneP.ID == payload.ID {

			// once found, verify tag is correct
			if oneP.Tag == payload.Tag {
				p = oneP
				break
			} else {
				err = errWrongTag
				mgr.logger.Debug(err)
				return output.JsonErrValidationFailed(err)
			}
		}
	}

	// didn't find id
	if p == nil {
		err = errBadID(payload.ID)
		mgr.logger.Debug(err)
		return output.JsonErrValidationFailed(err)
	}

	// if domains included, validate domains
	if len(payload.Domains) > 0 {
		err = mgr.unsafeValidateDomains(payload.Domains, p)
		if err != nil {
			err = fmt.Errorf("failed to validate domains (%s)", err)
			mgr.logger.Debug(err)
			return output.JsonErrValidationFailed(err)
		}
	}

	// error if wrong config count received
	configCount := 0
	var pCfg providerConfig
	if payload.Http01InternalConfig != nil {
		configCount++
		pCfg = payload.Http01InternalConfig
	}
	if payload.Dns01ManualConfig != nil {
		configCount++
		pCfg = payload.Dns01ManualConfig
	}
	if payload.Dns01AcmeDnsConfig != nil {
		configCount++
		pCfg = payload.Dns01AcmeDnsConfig
	}
	if payload.Dns01AcmeShConfig != nil {
		configCount++
		pCfg = payload.Dns01AcmeShConfig
	}
	if payload.Dns01CloudflareConfig != nil {
		configCount++
		pCfg = payload.Dns01CloudflareConfig
	}
	if payload.Dns01GoAcmeConfig != nil {
		configCount++
		pCfg = payload.Dns01GoAcmeConfig
	}

	// check config count, also error on wrong config type
	if configCount > 1 {
		err = fmt.Errorf("update provider expects max 1 config, received %d", configCount)
		mgr.logger.Debug(err)
		return output.JsonErrValidationFailed(err)
	} else if configCount == 1 {
		// update provider service first (if cfg specified) so if fails, domains are unchanged
		switch pServ := p.Service.(type) {
		case *http01internal.Service:
			if payload.Http01InternalConfig == nil {
				err = errors.New("update provider wrong config received")
				mgr.logger.Debug(err)
				return output.JsonErrValidationFailed(err)
			}
			err = pServ.UpdateService(mgr.childApp, payload.Http01InternalConfig)

		case *dns01manual.Service:
			if payload.Dns01ManualConfig == nil {
				err = errors.New("update provider wrong config received")
				mgr.logger.Debug(err)
				return output.JsonErrValidationFailed(err)
			}
			err = pServ.UpdateService(mgr.childApp, payload.Dns01ManualConfig)

		case *dns01acmedns.Service:
			if payload.Dns01AcmeDnsConfig == nil {
				err = errors.New("update provider wrong config received")
				mgr.logger.Debug(err)
				return output.JsonErrValidationFailed(err)
			}
			err = pServ.UpdateService(mgr.childApp, payload.Dns01AcmeDnsConfig)

		case *dns01acmesh.Service:
			if payload.Dns01AcmeShConfig == nil {
				err = errors.New("update provider wrong config received")
				mgr.logger.Debug(err)
				return output.JsonErrValidationFailed(err)
			}
			err = pServ.UpdateService(mgr.childApp, payload.Dns01AcmeShConfig)

		case *dns01cloudflare.Service:
			if payload.Dns01CloudflareConfig == nil {
				err = errors.New("update provider wrong config received")
				mgr.logger.Debug(err)
				return output.JsonErrValidationFailed(err)
			}
			err = pServ.UpdateService(mgr.childApp, payload.Dns01CloudflareConfig)

		case *dns01goacme.Service:
			if payload.Dns01GoAcmeConfig == nil {
				err = errors.New("update provider wrong config received")
				mgr.logger.Debug(err)
				return output.JsonErrValidationFailed(err)
			}
			err = pServ.UpdateService(mgr.childApp, payload.Dns01GoAcmeConfig)

		default:
			// default fail
			err = errors.New("provider service is unsupported, please report this as a bug to developer")
			mgr.logger.Error(err)
			return output.JsonErrInternal(err)
		}

		// common error check
		if err != nil {
			err = fmt.Errorf("failed to update service (%s)", err)
			mgr.logger.Debug(err)
			return output.JsonErrValidationFailed(err)
		}

		// success, update config
		p.Config = pCfg
	}

	// update any internal config changes
	mgr.unsafeUpdateProviderDomains(p, payload.Domains)

	if payload.PostProvisionWaitSeconds != nil {
		p.PostProvisionWaitSeconds = *payload.PostProvisionWaitSeconds
	}

	if payload.PostProvisionWaitSeconds != nil {
		p.PostProvisionWaitSeconds = *payload.PostProvisionWaitSeconds
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
	response.Message = "updated provider"
	response.Provider = p

	err = mgr.output.WriteJSON(w, response)
	if err != nil {
		mgr.logger.Errorf("failed to write json (%s)", err)
		return output.JsonErrWriteJsonError(err)
	}

	return nil
}
