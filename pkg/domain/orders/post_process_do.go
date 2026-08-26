package orders

// Do actually runs the post processing task(s)
func (j *postProcessJob) Do(workerID int) {
	// get order
	order, err := j.service.storage.GetOneOrder(j.orderID)
	if err != nil {
		j.service.logger.Errorf("orders: ost processing worker %d: failed to get order %d from db for post processing (%s)", workerID, j.orderID, err)
		return // done, failed
	}

	// run client post processing
	j.doClientPostProcess(order, workerID)

	// run command post processing
	j.doScriptOrBinaryPostProcess(order, workerID)
}
