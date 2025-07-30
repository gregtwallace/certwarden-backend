package acme

import (
	"bytes"
	"crypto"
	"crypto/ecdsa"
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/json"
	"errors"
	"strings"
)

// acmeSignedMessage contains the unencoded structure of an ACME Message
// as well as the json Marshalled httpBody of the properly encoded message.
// This is done so a json.Marshal of this struct gives a human friendly output
// but `marshalled` gives the correct bytes for communication with an ACME
// server.
type acmeSignedMessage struct {
	ProtectedHeader struct {
		Algorithm  string      `json:"alg"`
		JsonWebKey *jsonWebKey `json:"jwk,omitempty"`
		KeyId      string      `json:"kid,omitempty"`
		Nonce      string      `json:"nonce,omitempty"`
		Url        string      `json:"url"`
	} `json:"protected"`
	Payload   any    `json:"payload"`
	Signature string `json:"signature"`

	marshalled []byte `json:"-"`
}

// setNonceAndSign sets the nonce on the acmeSignedMessage and then generates the
// new signature and httpBody
func (asm *acmeSignedMessage) setNonceAndSign(nonce string, accountKey AccountKey) error {
	// set nonce
	asm.ProtectedHeader.Nonce = nonce

	// Signing & Body
	encodedHeader, err := encodeJson(asm.ProtectedHeader)
	if err != nil {
		return err
	}

	// only encoded payload if it isn't empty
	encodedPayload := ""
	if asm.Payload != "" {
		switch payload := asm.Payload.(type) {

		// special payload handling if the payload is a nested acmeSignedMessage (e.g., key rollover)
		case *acmeSignedMessage:
			// encode string, not json, as the httpBody was already marshalled
			encodedPayload = encodeString(payload.marshalled)

		// special payload handling if the payload is new account (needed for EAB)
		case *acmeNewAccountPayload:
			var eabMarshalled []byte
			if payload.ExternalAccountBinding != nil {
				eabMarshalled = payload.ExternalAccountBinding.marshalled
			}

			jsonStruct := struct {
				Contact                []string        `json:"contact"`
				TosAgreed              bool            `json:"termsOfServiceAgreed"`
				ExternalAccountBinding json.RawMessage `json:"externalAccountBinding,omitempty"`
			}{
				Contact:                payload.Contact,
				TosAgreed:              payload.TosAgreed,
				ExternalAccountBinding: eabMarshalled,
			}

			encodedPayload, err = encodeJson(jsonStruct)
			if err != nil {
				return err
			}

		default:
			encodedPayload, err = encodeJson(asm.Payload)
			if err != nil {
				return err
			}
		}
	}

	dataToSign := []byte(strings.Join([]string{encodedHeader, encodedPayload}, "."))

	// sign appropriately based on key type
	switch key := accountKey.Key.(type) {
	case *rsa.PrivateKey:
		// all rsa use RS256
		hash := crypto.SHA256
		hashed256 := sha256.Sum256(dataToSign)
		hashed := hashed256[:]

		// sign using the key
		signature, err := rsa.SignPKCS1v15(rand.Reader, key, hash, hashed)
		if err != nil {
			return err
		}

		// Make signature in the format ACME expects. Padding should not be required
		// for RSA.
		asm.Signature = encodeString(signature)

	case *ecdsa.PrivateKey:
		// hash has to be generated based on the header.Algorithm or will error
		var hashed []byte
		bitSize := key.PublicKey.Params().BitSize
		switch bitSize {
		case 256:
			hashed256 := sha256.Sum256(dataToSign)
			hashed = hashed256[:]

		case 384:
			hashed384 := sha512.Sum384(dataToSign)
			hashed = hashed384[:]

		default:
			return errors.New("acme: failed to sign (unsupported ec bit size)")
		}

		// sign using the key
		r, s, err := ecdsa.Sign(rand.Reader, key, hashed)
		if err != nil {
			return err
		}

		// ACME expects these values to be zero padded
		rPadded := padBytes(r.Bytes(), bitSize)
		sPadded := padBytes(s.Bytes(), bitSize)

		// combine the buffers and encode
		asm.Signature = encodeString(append(rPadded, sPadded...))

	case []byte:
		// mac signature (for EAB)
		mac := hmac.New(sha256.New, key)
		_, err = mac.Write(dataToSign)
		if err != nil {
			return err
		}

		// encoded Signature
		asm.Signature = encodeString(mac.Sum(nil))

	default:
		// not supported
		return errors.New("acme: sign: unsupported private key type")
	}

	// marshal the message in the expected format
	jsonStruct := struct {
		ProtectedHeader string `json:"protected"`
		Payload         string `json:"payload"`
		Signature       string `json:"signature"`
	}{
		ProtectedHeader: encodedHeader,
		Payload:         encodedPayload,
		Signature:       asm.Signature,
	}

	asm.marshalled, err = json.Marshal(jsonStruct)
	if err != nil {
		return err
	}

	return nil
}

// makeAcmeSignedMessage creates an acmeSignedMessage struct which is mostly unencoded
// for logging. Encoding for the HTTP body is performed in a different function.
func makeAcmeSignedMessage(payload any, nonce string, url string, accountKey AccountKey) (*acmeSignedMessage, error) {
	msg := new(acmeSignedMessage)
	var err error

	// ProtectedHeader
	// alg
	switch privateKey := accountKey.Key.(type) {
	case *rsa.PrivateKey:
		// all rsa use RS256
		msg.ProtectedHeader.Algorithm = "RS256"

	case *ecdsa.PrivateKey:
		switch privateKey.Curve.Params().Name {
		case "P-256":
			msg.ProtectedHeader.Algorithm = "ES256"
		case "P-384":
			msg.ProtectedHeader.Algorithm = "ES384"
		default:
			return nil, errors.New("acme: signature algorithm: unsupported ecdsa curve")
		}

	case []byte:
		// Assume []byte is an External Account Binding key, which always uses HS256
		msg.ProtectedHeader.Algorithm = "HS256"

	default:
		return nil, errors.New("acme: signature algorithm: unsupported private key type")
	}

	// jwk or kid (use kid if it is available, otherwise jwk)
	if accountKey.Kid != "" {
		msg.ProtectedHeader.JsonWebKey = nil
		msg.ProtectedHeader.KeyId = accountKey.Kid
	} else {
		msg.ProtectedHeader.JsonWebKey, err = accountKey.jwk()
		if err != nil {
			return nil, err
		}
		// msg.ProtectedHeader.KeyId = ""
	}

	// nonce
	// defer until later

	// url
	msg.ProtectedHeader.Url = url

	// Payload
	msg.Payload = payload

	// Signature
	err = msg.setNonceAndSign(nonce, accountKey)
	if err != nil {
		return nil, err
	}

	return msg, nil
}

func (asm *acmeSignedMessage) SignedHTTPBody() *bytes.Buffer {
	return bytes.NewBuffer(asm.marshalled)
}

// padBytes pads data to an appropriate byte size based on the specified
// number of bits (which generally comes from the key bit size)
func padBytes(data []byte, bitSize int) (padded []byte) {
	// calculate byte length (bits rounded up to nearest 8)
	octetLength := (bitSize + 7) >> 3

	// make new buffer of byte length
	padded = make([]byte, octetLength-len(data), octetLength)

	// insert the data into the padded buffer
	padded = append(padded, data...)

	return padded
}
