package orders

import (
	"fmt"
)

// fulfillOrder queues the specified order ID with the specified priority level
// for fulfillment of that order with the ACME server
func (service *Service) fulfillOrder(orderID int, isHighPriority bool) (err error) {
	// make job
	newJob, err := service.makeFulfillingJob(orderID, isHighPriority)
	if err != nil {
		return err
	}

	// add to the Job Manager
	err = service.orderFulfilling.AddJob(newJob)
	if err != nil {
		return fmt.Errorf("orders: fulfilling: failed to add order id %d (%s)", orderID, err)
	}

	return nil
}
