package challenges

// UpdateProvider updates a specific provider with a new configuration. It also
// allows for adding a new provider when using the new id
// func (service *Service) UpdateProvider(w http.ResponseWriter, r *http.Request) error {
// 	// parse payload
// 	var payload providers.UpdatePayload
// 	err := json.NewDecoder(r.Body).Decode(&payload)
// 	if err != nil {
// 		service.logger.Debug(err)
// 		return output.ErrValidationFailed
// 	}

// 	service.logger.Debug(payload.Config)

// 	// params (set payload ID)
// 	idParam := httprouter.ParamsFromContext(r.Context()).ByName("id")
// 	payload.ID, err = strconv.Atoi(idParam)
// 	if err != nil {
// 		service.logger.Debug(err)
// 		return output.ErrValidationFailed
// 	}

// 	// do update (providers/manager handles validation)
// 	err = service.providers.UpdateProvider(payload)
// 	if err != nil {
// 		// UpdateProvider provides a client friendly error
// 		return err
// 	}

// 	// return response to client
// 	response := output.JsonResponse{
// 		Status:  http.StatusOK,
// 		ID:      payload.ID,
// 		Message: "provider updated",
// 	}

// 	err = service.output.WriteJSON(w, response.Status, response, "response")
// 	if err != nil {
// 		return err
// 	}

// 	return nil
// }
