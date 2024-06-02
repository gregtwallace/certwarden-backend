package orders

import (
	"fmt"
	"time"
)

// orderFulfillJob represents a job that fulfills ACME orders, including all
// variables needed to actually Do the job
type orderFulfillJob struct {
	service *Service

	addedToQueue time.Time
	highPriority bool
	orderID      int
}

// makeFulfillingJob makes an orderFulfillJob
func (service *Service) makeFulfillingJob(orderID int, highPriority bool) (*orderFulfillJob, error) {
	// get order (i.e. validate it exists)
	order, err := service.storage.GetOneOrder(orderID)
	if err != nil {
		return nil, fmt.Errorf("orders: fulfilling: failed to make fulfill job for order id %d (%s)", orderID, err)
	}

	// cant fulfill if already in a final state
	if order.Status == "valid" || order.Status == "invalid" {
		return nil, fmt.Errorf("orders: fulfilling: failed to make fulfill job for order id %d (already in final state %s)", orderID, order.Status)
	}

	return &orderFulfillJob{
		service: service,

		addedToQueue: time.Now(),
		highPriority: highPriority,
		orderID:      orderID,
	}, nil
}

// Description implements part of the Job interface and returns a string
// that will be used for logging purposes
func (j *orderFulfillJob) Description() string {
	return fmt.Sprintf("order id: %d", j.orderID)
}

// Equal implements part of the Job interface to determine if two jobs
// should be considered the same job
func (j *orderFulfillJob) Equal(j2 *orderFulfillJob) bool {
	return j != nil && j2 != nil && j.orderID == j2.orderID
}

// IsHighPriority implements Job interface priority func
func (j *orderFulfillJob) IsHighPriority() bool {
	return j.highPriority
}
