package orders

import (
	"fmt"
	"time"
)

// orderFulfillerJob contains information the order worker will use to do work
type orderFulfillerJob struct {
	placedAt     int // unix time job was requested
	highPriority bool
	order        Order
}

// addJob adds the specified job to the queue and completes the job when a worker
// is available to do so
func (of *orderFulfiller) addJob(orderId int, highPriority bool) (err error) {
	of.mu.Lock()
	defer of.mu.Unlock()

	// fetch the relevant order
	orderDb, err := of.storage.GetOneOrder(orderId)
	if err != nil {
		of.logger.Errorf("cannot add order id %d (%s)", orderId, err)
		return err
	}

	// make job
	j := orderFulfillerJob{
		placedAt:     int(time.Now().Unix()),
		highPriority: highPriority,
		order:        orderDb,
	}

	// check if currently waiting
	errOrderAlreadyProcessing := fmt.Errorf("cannot add order id %d (already in process)", j.order.ID)
	for i := range of.jobsWaiting {
		if of.jobsWaiting[i].order.ID == j.order.ID {
			return errOrderAlreadyProcessing
		}
	}

	// check if currently with a worker
	for i := range of.workerJobs {
		if of.workerJobs[i] != nil && of.workerJobs[i].order.ID == j.order.ID {
			return errOrderAlreadyProcessing
		}
	}

	// if not, add to waiting
	of.jobsWaiting = append(of.jobsWaiting, j)

	// send job to worker channel (based on priority)
	// async required as this will block until other end of channel reads the job
	go func() {
		if j.highPriority {
			of.highJobs <- j
		} else {
			of.lowJobs <- j
		}
	}()

	return nil
}
