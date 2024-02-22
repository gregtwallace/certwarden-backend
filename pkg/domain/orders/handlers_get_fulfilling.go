package orders

import (
	"legocerthub-backend/pkg/output"
	"net/http"
)

// orderFulfillerJobResponse contains the json response struct for one fulfilling job
type orderFulfillingJobResponse struct {
	AddedToQueue int                  `json:"added_to_queue"` // unix time job was requested
	HighPriority bool                 `json:"high_priority"`
	Order        orderSummaryResponse `json:"order"`
}

// fulfillWorkStatusResponse contains the full response to a GET request for the status
// of the fulfilling service
type fulfillWorkStatusResponse struct {
	output.JsonResponse
	JobsWorking map[int]*orderFulfillingJobResponse `json:"jobs_working"` // [workerid]
	JobsWaiting []orderFulfillingJobResponse        `json:"jobs_waiting"`
}

// GetFulfillWorkStatus returns all fulfilling jobs with workers and waiting in queue
func (service *Service) GetFulfillWorkStatus(w http.ResponseWriter, r *http.Request) *output.Error {
	// get jobs from manager
	mgrJobs := service.orderFulfilling.AllCurrentJobs()

	// get Order IDs for all jobs (to query db)
	orderIDs := []int{}
	for _, mgrWorkingJob := range mgrJobs.WorkingJobs {
		// only add working if work isn't idle (i.e. it has a job)
		if mgrWorkingJob != nil {
			orderIDs = append(orderIDs, mgrWorkingJob.orderID)
		}
	}
	for _, mgrWaitingJob := range mgrJobs.WaitingJobs {
		orderIDs = append(orderIDs, mgrWaitingJob.orderID)
	}

	// lookup all orders in db
	orders, err := service.storage.GetOrders(orderIDs)
	if err != nil {
		service.logger.Errorf("orders: failed to convert fulfilling jobs to order objects (%w)", err)
		return output.ErrInternal
	}

	// build working part of response
	workingResp := make(map[int]*orderFulfillingJobResponse)
	for workerID, mgrWorkingJob := range mgrJobs.WorkingJobs {
		// find order that matches this workerID, and then make response
		if mgrWorkingJob == nil {
			workingResp[workerID] = nil
		} else {
			for _, order := range orders {
				if mgrWorkingJob.orderID == order.ID {
					workingResp[workerID] = &orderFulfillingJobResponse{
						AddedToQueue: int(mgrWorkingJob.addedToQueue.Unix()),
						HighPriority: mgrWorkingJob.IsHighPriority(),
						Order:        order.summaryResponse(service),
					}
					break
				}
			}
		}
	}

	// build waiting part of response
	waitingResp := []orderFulfillingJobResponse{}
	for _, mgrWaitingJob := range mgrJobs.WaitingJobs {
		for _, order := range orders {
			if mgrWaitingJob.orderID == order.ID {
				waitingResp = append(waitingResp, orderFulfillingJobResponse{
					AddedToQueue: int(mgrWaitingJob.addedToQueue.Unix()),
					HighPriority: mgrWaitingJob.IsHighPriority(),
					Order:        order.summaryResponse(service),
				})
				break
			}
		}
	}

	// final response
	jobsResp := &fulfillWorkStatusResponse{
		JsonResponse: output.JsonResponse{
			StatusCode: http.StatusOK,
			Message:    "ok",
		},
		JobsWorking: workingResp,
		JobsWaiting: waitingResp,
	}

	// serve final response
	err = service.output.WriteJSON(w, jobsResp)
	if err != nil {
		service.logger.Errorf("failed to write json (%s)", err)
		return output.ErrWriteJsonError
	}
	return nil
}
