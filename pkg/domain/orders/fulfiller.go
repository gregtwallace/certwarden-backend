package orders

// orderFromAcme updates inProcess to indicate the specified order is being
// processed and launches a go routine that sends the order job to a worker.
// Priority allows high priority orders to always be processed before low priority
// orders. The intent is for automated tasks to be low priority, vs manual user
// initiated tasks being high priority.
func (service *Service) orderFromAcme(orderId int, highPriority bool) (err error) {
	// add to working to indicate order is being worked
	err = service.inProcess.add(orderId)
	// error indicates already inProcess
	if err != nil {
		return err
	}

	// send the order job to a worker which will try and fulfill order and then remove the order
	// from inProcess after it is complete
	go func(service *Service, orderId int, highPriority bool) {
		job := orderJob{
			orderId: orderId,
		}

		// add job, based on priority
		if highPriority {
			service.highJobs <- job
		} else {
			service.lowJobs <- job
		}
	}(service, orderId, highPriority)

	return nil
}
