package orders

import (
	"fmt"
	"time"
)

// postProcessJob represents a post processing job, including all variables needed
// to actually Do the job
type postProcessJob struct {
	service *Service

	addedToQueue  time.Time
	highPriority  bool
	orderID       int
	certificateID int
}

// makeFulfillingJob makes an orderFulfillJob
func (service *Service) makePostProcessJob(orderID int, highPriority bool) (*postProcessJob, error) {
	// get order
	order, err := service.storage.GetOneOrder(orderID)
	if err != nil {
		return nil, fmt.Errorf("post processing: failed to make post process job for order id %d (%w)", orderID, err)
	}

	// fail add if order isn't valid
	if order.Status != "valid" {
		return nil, fmt.Errorf("post processing: failed to make post process job for order id %d (status is not 'valid')", orderID)
	}

	// confirm order actually has post processing to do
	if !order.hasPostProcessingToDo() {
		return nil, fmt.Errorf("post processing: failed to make post process job for order id %d (certificate %s has no post processing configured)", orderID, order.Certificate.Name)
	}

	return &postProcessJob{
		service: service,

		addedToQueue:  time.Now(),
		highPriority:  highPriority,
		orderID:       orderID,
		certificateID: order.Certificate.ID,
	}, nil
}

// Description implements part of the Job interface and returns a string
// that will be used for logging purposes
func (j *postProcessJob) Description() string {
	return fmt.Sprintf("certificate id: %d, order id: %d", j.certificateID, j.orderID)
}

// Equal implements part of the Job interface to determine if two jobs
// should be considered the same job
func (j *postProcessJob) Equal(j2 *postProcessJob) bool {
	return j != nil && j2 != nil && (j.orderID == j2.orderID || j.certificateID == j2.certificateID)
}

// IsHighPriority implements Job interface priority func
func (j *postProcessJob) IsHighPriority() bool {
	return j.highPriority
}
