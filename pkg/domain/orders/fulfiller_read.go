package orders

type allWorkStatus struct {
	JobsWaiting []orderFulfillerJobResponse        `json:"jobs_waiting"`
	WorkerJobs  map[int]*orderFulfillerJobResponse `json:"worker_jobs"` // [workerid]
}

type orderFulfillerJobResponse struct {
	PlacedAt     int                  `json:"placed_at"` // unix time job was requested
	HighPriority bool                 `json:"high_priority"`
	Order        orderSummaryResponse `json:"order"`
}

func (j *orderFulfillerJob) summaryResponse() *orderFulfillerJobResponse {
	if j == nil {
		return nil
	}

	return &orderFulfillerJobResponse{
		PlacedAt:     j.placedAt,
		HighPriority: j.highPriority,
		Order:        j.order.summaryResponse(),
	}
}

// allWorkStatus returns a summary of all work currently in the fulfiller queue
// or being worked by its workers
func (of *orderFulfiller) allWorkStatus() allWorkStatus {
	of.mu.RLock()
	defer of.mu.RUnlock()

	// convert waiting to response
	jobsWaiting := []orderFulfillerJobResponse{}
	for i := range of.jobsWaiting {
		jobsWaiting = append(jobsWaiting, *of.jobsWaiting[i].summaryResponse())
	}

	// convert workers to response
	workerJobs := make(map[int]*orderFulfillerJobResponse)
	for i := range of.workerJobs {
		workerJobs[i] = of.workerJobs[i].summaryResponse()
	}

	return allWorkStatus{
		JobsWaiting: jobsWaiting,
		WorkerJobs:  workerJobs,
	}
}
