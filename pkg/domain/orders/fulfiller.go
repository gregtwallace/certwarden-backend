package orders

// orderFromAcmeHigh updates inProcess to indicate the specified order is being
// processed and launches a go routine that sends the order job to a worker.
// High priority is used for the specified order.
func (service *Service) orderFromAcmeHigh(orderId int) (err error) {
	// add to working to indicate order is being worked
	err = service.inProcess.add(orderId)
	// error indicates already inProcess
	if err != nil {
		return err
	}

	// send the order job to a worker which will try and fulfill order and then remove the order
	// from inProcess after it is complete
	go func(orderId int, service *Service) {
		job := orderJob{
			orderId: orderId,
		}

		service.highJobs <- job
	}(orderId, service)

	return nil
}
