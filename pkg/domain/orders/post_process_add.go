package orders

import (
	"fmt"
)

// postProcess queues a post processing job for the specified order ID with the specified
// priority level
func (service *Service) postProcess(orderID int, isHighPriority bool) (err error) {
	// make job
	newJob, err := service.makePostProcessJob(orderID, isHighPriority)
	if err != nil {
		return err
	}

	// add to the Job Manager
	err = service.postProcessing.AddJob(newJob)
	if err != nil {
		return fmt.Errorf("post processing: failed to add order id %d (%s)", orderID, err)
	}

	return nil
}
