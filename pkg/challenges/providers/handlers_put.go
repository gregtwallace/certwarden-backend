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
	ID                    int                     `json:"-"`
	Tag                   string                  `json:"tag"`
	Http01InternalConfig  *http01internal.Config  `json:"http_01_internal,omitempty"`
	Dns01ManualConfig     *dns01manual.Config     `json:"dns_01_manual,omitempty"`
	Dns01AcmeDnsConfig    *dns01acmedns.Config    `json:"dns_01_acme_dns,omitempty"`
	Dns01AcmeShConfig     *dns01acmesh.Config     `json:"dns_01_acme_sh,omitempty"`
	Dns01CloudflareConfig *dns01cloudflare.Config `json:"dns_01_cloudflare,omitempty"`
}

// ModifyProvider modifies the provider specified by the ID in manager with the specified
// configuration.
func (mgr *Manager) ModifyProvider(w http.ResponseWriter, r *http.Request) (err error) {
	mgr.mu.Lock()
	defer mgr.mu.Unlock()

	// decode body into payload
	var payload modifyPayload
	err = json.NewDecoder(r.Body).Decode(&payload)
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

	// verify correct number of configs received
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
	if configCount != 1 {
		mgr.logger.Debugf("new provider expects 1 config, received %d", configCount)
		return output.ErrValidationFailed
	}

	// find provider
	p := (*provider)(nil)
	for oneP := range mgr.pD {
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

	// validate domains
	err = mgr.unsafeValidateDomains(pCfg, p)
	if err != nil {
		mgr.logger.Debugf("failed to validate domains (%s)", err)
		return output.ErrValidationFailed
	}

	// actually do update
	switch pServ := p.Service.(type) {
	case *http01internal.Service:
		err = pServ.UpdateService(mgr.childApp, payload.Http01InternalConfig)

	case *dns01manual.Service:
		err = pServ.UpdateService(mgr.childApp, payload.Dns01ManualConfig)

	case *dns01acmedns.Service:
		err = pServ.UpdateService(mgr.childApp, payload.Dns01AcmeDnsConfig)

	case *dns01acmesh.Service:
		err = pServ.UpdateService(mgr.childApp, payload.Dns01AcmeShConfig)

	case *dns01cloudflare.Service:
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

	// update manager provider info
	mgr.updateProvider(p, pCfg)

	// update config file
	err = mgr.unsafeWriteProvidersConfig()
	if err != nil {
		mgr.logger.Errorf("failed to save config file after providers update (%s)", err)
		return output.ErrInternal
	}

	// return response to client
	response := output.JsonResponse{
		Status:  http.StatusOK,
		Message: "updated provider",
		ID:      payload.ID,
	}

	err = mgr.output.WriteJSON(w, response.Status, response, "response")
	if err != nil {
		return err
	}

	return nil
}
