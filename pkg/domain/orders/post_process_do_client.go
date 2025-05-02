package orders

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const postProcessClientPostRoute = "/certwardenclient/api/v1/install"
const postProcessClientPort = 5055

// postProcessInnerClientPayload is the data that will be marshalled and
// encrypted, then encoded, then embedded in the outer struct before sending to client
type postProcessInnerClientPayload struct {
	KeyPem  string `json:"key_pem"`
	CertPem string `json:"cert_pem"`
}

// postProcessClientPayload is the actual payload that is sent to the client
type postProcessClientPayload struct {
	// Payload is the base64 encoded string of the cipherData produced from encrypting innerPayload
	Payload string `json:"payload"`
}

// doClientPostProcess sends a data payload to the client located
// at certificate's CN, using the encryption key specified on certificate
func (j *postProcessJob) doClientPostProcess(order Order, workerID int) {
	// no-op if no client key
	if order.Certificate.PostProcessingClientKeyB64 == "" || order.Certificate.PostProcessingClientAddress == "" {
		j.service.logger.Debugf("orders: post processing worker %d: order %d: skipping client notify (cert does not have a client address and/or client key) (cert: %d, cn: %s, addr: %s)", workerID, order.ID, order.Certificate.ID, order.Certificate.Subject, order.Certificate.PostProcessingClientAddress)
		return
	}

	j.service.logger.Infof("orders: post processing worker %d: order %d: attempting to notify client (cert: %d, cn: %s, addr: %s)", workerID, order.ID, order.Certificate.ID, order.Certificate.Subject, order.Certificate.PostProcessingClientAddress)

	// decode AES key
	aesKey, err := base64.RawURLEncoding.DecodeString(order.Certificate.PostProcessingClientKeyB64)
	if err != nil {
		j.service.logger.Errorf("orders: post processing worker %d: order %d: notify client failed: invalid aes key (%s) (cert: %d, cn: %s, addr: %s)", workerID, order.ID, err, order.Certificate.ID, order.Certificate.Subject, order.Certificate.PostProcessingClientAddress)
		return
	}

	// verify pem exists (should never trigger)
	if order.Pem == nil || order.FinalizedKey == nil {
		j.service.logger.Errorf("orders: post processing worker %d: order %d: notify client failed: something really weird happened and pem content is nil (cert: %d, cn: %s, addr: %s)", workerID, order.ID, order.Certificate.ID, order.Certificate.Subject, order.Certificate.PostProcessingClientAddress)
		return
	}

	// make inner payload for client
	innerPayload := postProcessInnerClientPayload{
		KeyPem:  order.FinalizedKey.Pem,
		CertPem: *order.Pem,
	}
	innerPayloadJson, err := json.Marshal(innerPayload)
	if err != nil {
		j.service.logger.Errorf("orders: post processing worker %d: order %d: notify client failed: failed to marshal inner payload (%s) (cert: %d, cn: %s, addr: %s)", workerID, order.ID, err, order.Certificate.ID, order.Certificate.Subject, order.Certificate.PostProcessingClientAddress)
		return
	}

	// make AES-GCM for encrypting
	aes, err := aes.NewCipher(aesKey)
	if err != nil {
		j.service.logger.Errorf("orders: post processing worker %d: order %d: notify client failed: failed to make cipher (%s) (cert: %d, cn: %s, addr: %s)", workerID, order.ID, err, order.Certificate.ID, order.Certificate.Subject, order.Certificate.PostProcessingClientAddress)
		return
	}

	gcm, err := cipher.NewGCM(aes)
	if err != nil {
		j.service.logger.Errorf("orders: post processing worker %d: order %d: notify client failed: failed to make gcm AEAD (%s) (cert: %d, cn: %s, addr: %s)", workerID, order.ID, err, order.Certificate.ID, order.Certificate.Subject, order.Certificate.PostProcessingClientAddress)
		return
	}

	// make nonce and encrypt
	nonce := make([]byte, gcm.NonceSize())
	_, err = rand.Read(nonce)
	if err != nil {
		j.service.logger.Errorf("orders: post processing worker %d: order %d: notify client failed: failed to make nonce (%s) (cert: %d, cn: %s, addr: %s)", workerID, order.ID, err, order.Certificate.ID, order.Certificate.Subject, order.Certificate.PostProcessingClientAddress)
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
		j.service.logger.Errorf("orders: post processing worker %d: order %d: notify client failed: failed to marshal outer payload (%s) (cert: %d, cn: %s, addr: %s)", workerID, order.ID, err, order.Certificate.ID, order.Certificate.Subject, order.Certificate.PostProcessingClientAddress)
		return
	}

	// send post to client
	postTo := fmt.Sprintf("https://%s:%d%s", order.Certificate.PostProcessingClientAddress, postProcessClientPort, postProcessClientPostRoute)
	resp, err := j.service.httpClient.Post(postTo, "application/json", bytes.NewBuffer(dataPayload))
	if err != nil {
		j.service.logger.Errorf("orders: post processing worker %d: order %d: notify client failed: failed to post to client (%s) (cert: %d, cn: %s, addr: %s)", workerID, order.ID, err, order.Certificate.ID, order.Certificate.Subject, order.Certificate.PostProcessingClientAddress)
		return
	}

	// ensure body is read and closed
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)

	// error if not 200
	if resp.StatusCode != http.StatusOK {
		j.service.logger.Errorf("orders: post processing worker %d: order %d: notify client failed: post status %d (cert: %d, cn: %s, addr: %s)", workerID, order.ID, resp.StatusCode, order.Certificate.ID, order.Certificate.Subject, order.Certificate.PostProcessingClientAddress)
		return
	}

	j.service.logger.Infof("orders: post processing worker %d: order %d: client notify completed", workerID, order.ID)
}
