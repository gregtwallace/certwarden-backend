package orders

import (
	"legocerthub-backend/pkg/output"
	"net/http"
)

type allWorkStatusResponse struct {
	output.JsonResponse
	WorkerJobs  map[int]*orderFulfillerJobResponse `json:"worker_jobs"` // [workerid]
	JobsWaiting []orderFulfillerJobResponse        `json:"jobs_waiting"`
}

type orderFulfillerJobResponse struct {
	AddedToQueue int                  `json:"added_to_queue"` // unix time job was requested
	HighPriority bool                 `json:"high_priority"`
	Order        orderSummaryResponse `json:"order"`
}

func (j *orderFulfillerJob) summaryResponse(of *orderFulfiller) *orderFulfillerJobResponse {
	if j == nil {
		return nil
	}

	return &orderFulfillerJobResponse{
		AddedToQueue: j.addedToQueue,
		HighPriority: j.highPriority,
		Order:        j.order.summaryResponse(of),
	}
}

// allWorkStatus returns a summary of all work currently in the fulfiller queue
// or being worked by its workers
func (of *orderFulfiller) allWorkStatus() *allWorkStatusResponse {
	of.mu.RLock()
	defer of.mu.RUnlock()

	// convert waiting to response
	jobsWaiting := []orderFulfillerJobResponse{}
	for i := range of.jobsWaiting {
		jobsWaiting = append(jobsWaiting, *of.jobsWaiting[i].summaryResponse(of))
	}

	// convert workers to response
	workerJobs := make(map[int]*orderFulfillerJobResponse)
	for i := range of.workerJobs {
		workerJobs[i] = of.workerJobs[i].summaryResponse(of)
	}

	// make response
	response := &allWorkStatusResponse{}
	response.StatusCode = http.StatusOK
	response.Message = "ok"
	response.WorkerJobs = workerJobs
	response.JobsWaiting = jobsWaiting

	return response
}

// checkForOrderId returns the worker that is currently working the specified
// orderId. If the order is in the waiting queue, a negative int is
// returned. If the order is not with a worker or waiting, nil is returned.
func (of *orderFulfiller) checkForOrderId(orderId int) *int {
	of.mu.RLock()
	defer of.mu.RUnlock()

	// check workers
	for i := range of.workerJobs {
		if of.workerJobs[i] != nil && of.workerJobs[i].order.ID == orderId {
			return &i
		}
	}

	// check waiting
	for i := range of.jobsWaiting {
		if of.jobsWaiting[i].order.ID == orderId {
			result := -1
			return &result
		}
	}

	// not found
	return nil
}
