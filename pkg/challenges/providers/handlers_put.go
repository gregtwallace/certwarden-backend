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
	"strconv"

	"github.com/julienschmidt/httprouter"
)

// modifyPayload is the needed payload to update an existing provider
type modifyPayload struct {
	// mandatory
	ID  int    `json:"-"`
	Tag string `json:"tag"`

	// optional
	Domains []string `json:"domains,omitempty"`

	// plus only one of these
	Http01InternalConfig  *http01internal.Config  `json:"http_01_internal,omitempty"`
	Dns01ManualConfig     *dns01manual.Config     `json:"dns_01_manual,omitempty"`
	Dns01AcmeDnsConfig    *dns01acmedns.Config    `json:"dns_01_acme_dns,omitempty"`
	Dns01AcmeShConfig     *dns01acmesh.Config     `json:"dns_01_acme_sh,omitempty"`
	Dns01CloudflareConfig *dns01cloudflare.Config `json:"dns_01_cloudflare,omitempty"`
}

// ModifyProvider modifies the provider specified by the ID in manager with the specified
// configuration.
func (mgr *Manager) ModifyProvider(w http.ResponseWriter, r *http.Request) *output.Error {
	mgr.mu.Lock()
	defer mgr.mu.Unlock()

	// decode body into payload
	var payload modifyPayload
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		mgr.logger.Debug(err)
		return output.ErrValidationFailed
	}

	// params
	idParam := httprouter.ParamsFromContext(r.Context()).ByName("id")
	payload.ID, err = strconv.Atoi(idParam)
	if err != nil {
		mgr.logger.Debug(err)
		return output.ErrValidationFailed
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
				mgr.logger.Debug(errWrongTag)
				return output.ErrValidationFailed
			}
		}
	}

	// didn't find id
	if p == nil {
		mgr.logger.Debug(errBadID(payload.ID))
		return output.ErrValidationFailed
	}

	// if domains included, validate domains
	if payload.Domains != nil {
		err = mgr.unsafeValidateDomains(payload.Domains, p)
		if err != nil {
			mgr.logger.Debugf("failed to validate domains (%s)", err)
			return output.ErrValidationFailed
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

	// check config count, also error on wrong config type
	if configCount > 1 {
		mgr.logger.Debugf("update provider expects max 1 config, received %d", configCount)
		return output.ErrValidationFailed
	} else if configCount == 1 {
		// update provider service first (if cfg specified) so if fails, domains are unchanged
		switch pServ := p.Service.(type) {
		case *http01internal.Service:
			if payload.Http01InternalConfig == nil {
				mgr.logger.Debug("update provider wrong config received")
				return output.ErrValidationFailed
			}
			err = pServ.UpdateService(mgr.childApp, payload.Http01InternalConfig)

		case *dns01manual.Service:
			if payload.Dns01ManualConfig == nil {
				mgr.logger.Debug("update provider wrong config received")
				return output.ErrValidationFailed
			}
			err = pServ.UpdateService(mgr.childApp, payload.Dns01ManualConfig)

		case *dns01acmedns.Service:
			if payload.Dns01AcmeDnsConfig == nil {
				mgr.logger.Debug("update provider wrong config received")
				return output.ErrValidationFailed
			}
			err = pServ.UpdateService(mgr.childApp, payload.Dns01AcmeDnsConfig)

		case *dns01acmesh.Service:
			if payload.Dns01AcmeShConfig == nil {
				mgr.logger.Debug("update provider wrong config received")
				return output.ErrValidationFailed
			}
			err = pServ.UpdateService(mgr.childApp, payload.Dns01AcmeShConfig)

		case *dns01cloudflare.Service:
			if payload.Dns01CloudflareConfig == nil {
				mgr.logger.Debug("update provider wrong config received")
				return output.ErrValidationFailed
			}
			err = pServ.UpdateService(mgr.childApp, payload.Dns01CloudflareConfig)

		default:
			// default fail
			mgr.logger.Error("provider service is unsupported, please report this as a lego bug")
			return output.ErrInternal
		}

		// common error check
		if err != nil {
			mgr.logger.Debugf("failed to update service (%s)", err)
			return output.ErrValidationFailed
		}

		// success, update config
		p.Config = pCfg
	}

	// actually do domains update
	mgr.unsafeUpdateProviderDomains(p, payload.Domains)

	// update config file
	err = mgr.unsafeWriteProvidersConfig()
	if err != nil {
		mgr.logger.Errorf("failed to save config file after providers update (%s)", err)
		return output.ErrInternal
	}

	// write response
	response := &providerResponse{}
	response.StatusCode = http.StatusCreated
	response.Message = "updated provider"
	response.Provider = p

	err = mgr.output.WriteJSON(w, response)
	if err != nil {
		mgr.logger.Errorf("failed to write json (%s)", err)
		return output.ErrWriteJsonError
	}

	return nil
}
