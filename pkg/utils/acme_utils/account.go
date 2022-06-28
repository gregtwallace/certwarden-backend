package acme_utils

// // UpdateAccount updates an existing account
// //  used to change contact and to set deactivated status
// func (dir AcmeDirectory) UpdateAccount(acct AcmeAccount, keyPem string, kid string) (AcmeAccountResponse, error) {
// 	var err error

// 	// translate the pem into useful structs
// 	pemBlock, err := utils.PemDecodeAndNormalize(keyPem)
// 	if err != nil {
// 		return AcmeAccountResponse{}, err
// 	}

// 	algorithmValue, err := utils.PrivateKeyAlgorithmValue(pemBlock)
// 	if err != nil {
// 		return AcmeAccountResponse{}, err
// 	}
// 	alg, err := key_crypto.AlgorithmByValue(algorithmValue)
// 	if err != nil {
// 		return AcmeAccountResponse{}, err
// 	}

// 	var ecKey *ecdsa.PrivateKey
// 	// var rsaKey *rsa.PrivateKey
// 	message := acmeMessage{}

// 	//// header
// 	header := protectedHeader{}
// 	header.Url = kid
// 	header.Algorithm = alg.SignatureAlg

// 	/// kid
// 	header.KeyId = kid

// 	// Nonce
// 	// TO DO - use prior nonce from last request (instead of getting one each time)
// 	response, err := http.Get(dir.NewNonce)
// 	if err != nil {
// 		return AcmeAccountResponse{}, err
// 	}
// 	header.Nonce = response.Header.Get("Replay-Nonce")
// 	response.Body.Close()

// 	// encode the header and load it into the message
// 	message.ProtectedHeader, err = encodeAcmeData(header)
// 	if err != nil {
// 		return AcmeAccountResponse{}, err
// 	}
// 	//// header end

// 	//// payload
// 	// load the payload
// 	message.Payload, err = encodeAcmeData(acct)
// 	if err != nil {
// 		log.Println("Failed to encode payload.")
// 		return AcmeAccountResponse{}, err
// 	}

// 	//// signature
// 	if alg.KeyType == "EC" {
// 		ecKey, err = x509.ParseECPrivateKey(pemBlock.Bytes)
// 		if err != nil {
// 			return AcmeAccountResponse{}, err
// 		}

// 		message.Signature, err = acmeEcSignature(message, ecKey)
// 		if err != nil {
// 			return AcmeAccountResponse{}, err
// 		}
// 	} else if alg.KeyType == "RSA" {
// 		// TODO Deal with RSA keys sig
// 	} else {
// 		// Unsupported key type; should never happen here due to other if block
// 		return AcmeAccountResponse{}, errors.New("Unsupported key type")
// 	}

// 	// Post the account to LE
// 	messageJson, err := json.Marshal(message)
// 	if err != nil {
// 		return AcmeAccountResponse{}, err
// 	}

// 	response, err = http.Post(header.Url, "application/jose+json", bytes.NewBuffer(messageJson))
// 	if err != nil {
// 		return AcmeAccountResponse{}, err
// 	}
// 	defer response.Body.Close()

// 	body, err := ioutil.ReadAll(response.Body)

// 	// unmarshal the LE response into an Account
// 	var responseAccount AcmeAccountResponse
// 	err = UnmarshalAcmeResp(body, &responseAccount)
// 	if err != nil {
// 		return AcmeAccountResponse{}, err
// 	}
// 	// kid isn't in the header for updates
// 	responseAccount.Location = kid

// 	return responseAccount, nil

// }
