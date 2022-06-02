package acme_utils

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/json"
	"errors"
	"io/ioutil"
	"legocerthub-backend/utils"
	"log"
	"net/http"
)

func (dir AcmeDirectory) CreateAccount(acct AcmeAccount, keyPem string) (AcmeAccountResponse, error) {
	var err error

	// translate the pem into useful structs
	pemBlock, err := utils.PemDecodeAndNormalize(keyPem)
	if err != nil {
		return AcmeAccountResponse{}, err
	}

	algorithmValue, err := utils.PrivateKeyAlgorithmValue(pemBlock)
	if err != nil {
		return AcmeAccountResponse{}, err
	}
	alg := utils.AlgorithmByValue(algorithmValue)

	var ecKey *ecdsa.PrivateKey
	// var rsaKey *rsa.PrivateKey
	message := acmeMessage{}

	//// header
	header := protectedHeader{}
	header.Url = dir.NewAccount
	header.Algorithm = alg.SignatureAlg

	/// JWK
	// decode the pem

	if alg.KeyType == "EC" {
		ecKey, err = x509.ParseECPrivateKey(pemBlock.Bytes)
		if err != nil {
			return AcmeAccountResponse{}, err
		}

		header.JsonWebKey = JwkEcKey(ecKey)
	} else if alg.KeyType == "RSA" {
		// TODO Deal with RSA keys to JWK
	} else {
		// Unsupported key type
		return AcmeAccountResponse{}, errors.New("Unsupported key type")
	}
	/// JWK end

	// Nonce
	// TO DO - use prior nonce from last request (instead of getting one each time)
	response, err := http.Get(dir.NewNonce)
	if err != nil {
		return AcmeAccountResponse{}, err
	}
	header.Nonce = response.Header.Get("Replay-Nonce")
	response.Body.Close()

	// encode the header and load it into the message
	message.ProtectedHeader, err = encodeAcmeData(header)
	if err != nil {
		return AcmeAccountResponse{}, err
	}
	//// header end

	//// payload
	// load the payload (for account this is just contact and tos)
	message.Payload, err = encodeAcmeData(acct)
	if err != nil {
		log.Println("Failed to encode payload.")
		return AcmeAccountResponse{}, err
	}

	//// signature
	if alg.KeyType == "EC" {
		message.Signature, err = acmeEcSignature(message, ecKey)
		if err != nil {
			return AcmeAccountResponse{}, err
		}
	} else if alg.KeyType == "RSA" {
		// TODO Deal with RSA keys sig
	} else {
		// Unsupported key type; should never happen here due to other if block
		return AcmeAccountResponse{}, errors.New("Unsupported key type")
	}

	// Post the account to LE
	messageJson, err := json.Marshal(message)
	if err != nil {
		return AcmeAccountResponse{}, err
	}

	response, err = http.Post(dir.NewAccount, "application/jose+json", bytes.NewBuffer(messageJson))
	if err != nil {
		return AcmeAccountResponse{}, err
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)

	// unmarshal the LE response into an Account
	var responseAccount AcmeAccountResponse
	err = json.Unmarshal(body, &responseAccount)
	if err != nil {
		return AcmeAccountResponse{}, err
	}
	// kid isn't part of the JSON response, fetch it from the header
	responseAccount.Location = response.Header.Get("Location")

	return responseAccount, nil
}

// UpdateAccount is the same function as creating an account, but uses a kid instead
//  of the key
// func UpdateAccount(dir AcmeDirectory, account AcmeAccount) (AcmeAccount, error) {
// 	return CreateAccount(dir, account)
// }
