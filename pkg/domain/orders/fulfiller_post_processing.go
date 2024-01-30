package orders

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

const postProcessClientPostRoute = "/legocerthubclient/api/v1/install"
const postProcessClientPort = 5055

// postProcessInnerClientPayload is the data that will be marshalled and
// encrypted, then encoded, then embedded in the outer struct before sending to client
type postProcessInnerClientPayload struct {
	KeyPem  string `json:"key_pem"`
	CertPem string `json:"cert_pem"`
}

// postProcessClientPayload is the actual payload that is sent to the lego
// client
type postProcessClientPayload struct {
	// Payload is the base64 encoded string of the cipherData produced from encrypting innerPayload
	Payload string `json:"payload"`
}

// executePostProcessingLeGoClient sends a data payload to the LeGo CertHub client located
// at certificate's CN, using the encryption key specified on certificate
func (of *orderFulfiller) executePostProcessingLeGoClient(order Order) {
	// no-op if no client key
	if order.Certificate.PostProcessingClientKeyB64 == "" {
		of.logger.Debugf("post processing %d: skipping lego client notify (cert does not have a client key) (cert: %d, cn: %s)", order.ID, order.Certificate.ID, order.Certificate.Subject)
		return
	}

	of.logger.Infof("post processing %d: attempting to notify lego client (cert: %d, cn: %s)", order.ID, order.Certificate.ID, order.Certificate.Subject)

	// decode AES key
	aesKey, err := base64.RawURLEncoding.DecodeString(order.Certificate.PostProcessingClientKeyB64)
	if err != nil {
		of.logger.Errorf("post processing %d: notify lego client failed: invalid aes key (%s) (cert: %d, cn: %s)", order.ID, err, order.Certificate.ID, order.Certificate.Subject)
		return
	}

	// verify pem exists (should never trigger)
	if order.Pem == nil || order.FinalizedKey == nil {
		of.logger.Errorf("post processing %d: notify lego client failed: something really weird happened and pem content is nil (cert: %d, cn: %s)", order.ID, order.Certificate.ID, order.Certificate.Subject)
		return
	}

	// make inner payload for client
	innerPayload := postProcessInnerClientPayload{
		KeyPem:  order.FinalizedKey.Pem,
		CertPem: *order.Pem,
	}
	innerPayloadJson, err := json.Marshal(innerPayload)
	if err != nil {
		of.logger.Errorf("post processing %d: notify lego client failed: failed to marshal inner payload (%s) (cert: %d, cn: %s)", order.ID, err, order.Certificate.ID, order.Certificate.Subject)
		return
	}

	// make AES-GCM for encrypting
	aes, err := aes.NewCipher(aesKey)
	if err != nil {
		of.logger.Errorf("post processing %d: notify lego client failed: failed to make cipher (%s) (cert: %d, cn: %s)", order.ID, err, order.Certificate.ID, order.Certificate.Subject)
		return
	}

	gcm, err := cipher.NewGCM(aes)
	if err != nil {
		of.logger.Errorf("post processing %d: notify lego client failed: failed to make gcm AEAD (%s) (cert: %d, cn: %s)", order.ID, err, order.Certificate.ID, order.Certificate.Subject)
		return
	}

	// make nonce and encrypt
	nonce := make([]byte, gcm.NonceSize())
	_, err = rand.Read(nonce)
	if err != nil {
		of.logger.Errorf("post processing %d: notify lego client failed: failed to make nonce (%s) (cert: %d, cn: %s)", order.ID, err, order.Certificate.ID, order.Certificate.Subject)
		return
	}
	// note: dst==nonce on purpose (so nonce is prepended)
	encryptedInnerData := gcm.Seal(nonce, nonce, innerPayloadJson, nil)

	// make actual payload to send client
	payload := postProcessClientPayload{
		Payload: base64.RawURLEncoding.EncodeToString(encryptedInnerData),
	}

	dataPayload, err := json.Marshal(payload)
	if err != nil {
		of.logger.Errorf("post processing %d: notify lego client failed: failed to marshal outer payload (%s) (cert: %d, cn: %s)", order.ID, err, order.Certificate.ID, order.Certificate.Subject)
		return
	}

	// send post to client
	postTo := fmt.Sprintf("https://%s:%d%s", order.Certificate.Subject, postProcessClientPort, postProcessClientPostRoute)
	resp, err := of.httpClient.Post(postTo, "application/json", bytes.NewBuffer(dataPayload))
	if err != nil {
		of.logger.Errorf("post processing %d: notify lego client failed: failed to post to client (%s) (cert: %d, cn: %s)", order.ID, err, order.Certificate.ID, order.Certificate.Subject)
		return
	}

	// ensure body is read and closed
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)

	// error if not 200
	if resp.StatusCode != http.StatusOK {
		of.logger.Errorf("post processing %d: notify lego client failed: post status %d (cert: %d, cn: %s)", order.ID, resp.StatusCode, order.Certificate.ID, order.Certificate.Subject)
		return
	}

	of.logger.Infof("post processing %d: lego client notify completed", order.ID)
}

// executePostProcessingScript executes the certificate's post processing script, if the cert
// does not have a script, this is a no-op
func (of *orderFulfiller) executePostProcessingScript(order Order) {
	// no-op if no command
	if order.Certificate.PostProcessingCommand == "" {
		of.logger.Debugf("post processing %d: skipping script (cert does not have a script to run) (cert: %d, cn: %s)", order.ID, order.Certificate.ID, order.Certificate.Subject)
		return
	}

	of.logger.Infof("post processing %d: attempting to run script (cert: %d, cn: %s)", order.ID, order.Certificate.ID, order.Certificate.Subject)

	// if app failed to get suitable shell at startup, post processing is disabled
	if of.shellPath == "" {
		err := fmt.Errorf("post processing %d: script failed to run post processing (no suitable shell was found during app startup)", order.ID)
		of.logger.Error(err)
		return
	}

	// nil checks
	if order.Pem == nil {
		err := fmt.Errorf("post processing %d: script failed: order pem is nil (should never happen)", order.ID)
		of.logger.Error(err)
		return
	}
	if order.FinalizedKey == nil {
		err := fmt.Errorf("post processing %d: script failed: finalized key no longer exists", order.ID)
		of.logger.Error(err)
		return
	}

	// make environment; certain values are always automatically added
	// LEGO_PRIVATE_KEY_NAME				= the name of the private key used to finalize the order, as defined in LeGo
	// LEGO_PRIVATE_KEY_PEM					= the pem of the private key
	// LEGO_CERTIFICATE_NAME				= the name of the certificate as defined in LeGo
	// LEGO_CERTIFICATE_PEM					= the pem of the complete certificate chain for the order
	// LEGO_CERTIFICATE_COMMON_NAME = the common name of the certificate
	environ := []string{}

	// LEGO_PRIVATE_KEY_NAME
	environ = append(environ, fmt.Sprintf("LEGO_PRIVATE_KEY_NAME=%s", order.FinalizedKey.Name))

	// LEGO_PRIVATE_KEY_PEM
	environ = append(environ, fmt.Sprintf("LEGO_PRIVATE_KEY_PEM=%s", order.FinalizedKey.Pem))

	// LEGO_CERTIFICATE_NAME
	environ = append(environ, fmt.Sprintf("LEGO_CERTIFICATE_NAME=%s", order.Certificate.Name))

	// LEGO_CERTIFICATE_PEM
	environ = append(environ, fmt.Sprintf("LEGO_CERTIFICATE_PEM=%s", *order.Pem))

	// LEGO_CERTIFICATE_COMMON_NAME
	environ = append(environ, fmt.Sprintf("LEGO_CERTIFICATE_COMMON_NAME=%s", order.Certificate.Subject))

	// user specified environment can have placeholders for the above values (so user can set
	// their own name for the environment variable)
	// {{PRIVATE_KEY_NAME}}					= the name of the private key used to finalize the order, as defined in LeGo
	// {{PRIVATE_KEY_PEM}}					= the pem of the private key
	// {{CERTIFICATE_NAME}}					= the name of the certificate as defined in LeGo
	// {{CERTIFICATE_PEM}}					= the pem of the complete certificate chain for the order
	// {{CERTIFICATE_COMMON_NAME}}	= the common name of the certificate

	// update placeholders and set user values
	for i := range order.Certificate.PostProcessingEnvironment {
		envItemSplit := strings.SplitN(order.Certificate.PostProcessingEnvironment[i], "=", -1)

		// only append properly formatted vars (i.e. VAR=someValue; len after = split == 2)
		if len(envItemSplit) == 2 {
			// do replacements if user val is a value that should be replaced
			switch val := strings.ToUpper(envItemSplit[1]); val {
			case "{{PRIVATE_KEY_NAME}}":
				envItemSplit[1] = order.FinalizedKey.Name

			case "{{PRIVATE_KEY_PEM}}":
				envItemSplit[1] = order.FinalizedKey.Pem

			case "{{CERTIFICATE_NAME}}":
				envItemSplit[1] = order.Certificate.Name

			case "{{CERTIFICATE_PEM}}":
				envItemSplit[1] = *order.Pem

			case "{{CERTIFICATE_COMMON_NAME}}":
				envItemSplit[1] = order.Certificate.Subject

			default:
				// no-op - user specified some other value
			}

			// append to environment
			environ = append(environ, envItemSplit[0]+"="+envItemSplit[1])
		} else {
			of.logger.Errorf("post processing %d: %s is not a properly formatted environment variable, it will be skipped", order.ID, order.Certificate.PostProcessingEnvironment[i])
		}
	}

	// make args for command
	// 0 - script name (e.g. /path/to/script.sh)
	args := []string{order.Certificate.PostProcessingCommand}

	// make command
	cmd := exec.Command(of.shellPath, args...)

	// set command environment (default OS + environ from above)
	cmd.Env = append(os.Environ(), environ...)

	// run script command
	result, err := cmd.Output()
	of.logger.Debugf("post processing %d: script output: %s", order.ID, string(result))
	if err != nil {
		// try to get stderr and log it too
		exitErr := new(exec.ExitError)
		if errors.As(err, &exitErr) {
			of.logger.Errorf("post processing %d: script std err: %s", order.ID, exitErr.Stderr)
		}

		of.logger.Errorf("post processing %d: script failed: error: %s", order.ID, err)
		return
	}

	of.logger.Infof("post processing %d: script completed", order.ID)
}

// executePostProcessing executes the order's certificate's post processing script
// if the script field is blank, this is a no-op
func (of *orderFulfiller) executePostProcessing(order Order) {
	of.logger.Infof("post processing %d: begin attempt", order.ID)

	// run client post processing
	of.executePostProcessingLeGoClient(order)

	// run script post processing
	of.executePostProcessingScript(order)

	of.logger.Infof("post processing %d: end", order.ID)
}
