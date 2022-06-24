package acme

import (
	"bytes"
	"crypto"
	"encoding/json"
	"io/ioutil"
	"net/http"
)

// acmeSignedMessage is the ACME signed message payload
type acmeSignedMessage struct {
	Payload         string `json:"payload"`
	ProtectedHeader string `json:"protected"`
	Signature       string `json:"signature"`
}

// ProtectedHeader piece of the ACME payload
type protectedHeader struct {
	Algorithm  string      `json:"alg"`
	JsonWebKey *jsonWebKey `json:"jwk,omitempty"`
	KeyId      string      `json:"kid,omitempty"`
	Nonce      string      `json:"nonce"`
	Url        string      `json:"url"`
}

// jsonWebKey for the ACME protectedHeader
type jsonWebKey struct {
	KeyType        string `json:"kty,omitempty"`
	PublicExponent string `json:"e,omitempty"`   // RSA
	Modulus        string `json:"n,omitempty"`   // RSA
	CurveName      string `json:"crv,omitempty"` // EC
	CurvePointX    string `json:"x,omitempty"`   // EC
	CurvePointY    string `json:"y,omitempty"`   // EC
}

// AccountKey is the necessary account / key information for signed message generation
type AccountKey struct {
	Key crypto.PrivateKey
	Kid string
}

// postToUrlSigned posts the payload to the specified url, using the specified AccountKeyInfo
// and returns the response from ACME
func (service *Service) postToUrlSigned(payload any, url string, accountKey AccountKey) (acmeResponse any, err error) {
	// message is what will ultimately be posted to ACME
	var message acmeSignedMessage

	/// header
	var header protectedHeader

	// alg
	header.Algorithm, err = accountKey.signingAlg()
	if err != nil {
		return nil, err
	}

	// key or kid
	// use kid if available, otherwise use jsonWebKey
	if accountKey.Kid != "" {
		header.JsonWebKey = nil
		header.KeyId = accountKey.Kid
	} else {
		header.JsonWebKey, err = accountKey.jwk()
		header.KeyId = ""
	}

	// nonce
	// TODO - implement nonce manager
	response, err := http.Get(service.dir.NewNonce)
	if err != nil {
		return nil, err
	}
	header.Nonce = response.Header.Get("Replay-Nonce")
	response.Body.Close()
	// TODO (end)

	// url
	header.Url = url

	// encord and insert into message
	message.ProtectedHeader, err = encodeJson(header)
	if err != nil {
		return nil, err
	}
	/// header (end)

	/// payload
	message.Payload, err = encodeJson(payload)

	/// signature
	message.Signature, err = accountKey.Sign(message)

	/// post
	messageJson, err := json.Marshal(message)
	if err != nil {
		return nil, err
	}

	response, err = http.Post(url, "application/jose+json", bytes.NewBuffer(messageJson))
	if err != nil {
		return nil, err
	}
	// TODO: Add new nonce to nonce manager
	_ = response.Header.Get("Replay-Nonce")
	defer response.Body.Close()

	// TODO: Remove and switch to returning response (i.e. remove this line)
	_, err = ioutil.ReadAll(response.Body)

	// body, err := ioutil.ReadAll(response.Body)

	// TODO: return response
	// unmarshal the LE response into an Account
	// var responseAccount AcmeAccountResponse
	// err = UnmarshalAcmeResp(body, &responseAccount)
	// if err != nil {
	// 	return AcmeAccountResponse{}, err
	// }
	// kid isn't part of the JSON response, fetch it from the header
	// responseAccount.Location = response.Header.Get("Location")
	// instead of just Location can return all headers to the calling func
	// and then they can be parsed generally as appropriate by the caller

	// TODO: return response
	return
}
