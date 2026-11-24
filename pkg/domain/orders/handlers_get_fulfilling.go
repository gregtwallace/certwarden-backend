package orders

import (
	"certwarden-backend/pkg/output"
	"fmt"
	"net/http"
)

// orderJobResponse contains the json response struct for one order job
type orderJobResponse struct {
	AddedToQueue int                  `json:"added_to_queue"` // unix time job was requested
	HighPriority bool                 `json:"high_priority"`
	Order        orderSummaryResponse `json:"order"`
}

// orderWorkStatusResponse contains the full response to a GET request for the status
// of the a work service
type orderWorkStatusResponse struct {
	output.JsonResponse

	JobsWorking map[int]*orderJobResponse `json:"jobs_working"` // [workerid]
	JobsWaiting []orderJobResponse        `json:"jobs_waiting"`
}

// GetFulfillWorkStatus returns all fulfilling jobs with workers and waiting in queue
func (service *Service) GetFulfillWorkStatus(w http.ResponseWriter, r *http.Request) *output.JsonError {
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

	// lookup all orders in db, if any jobs exist
	orders := []*Order{}
	var err error
	if len(orderIDs) > 0 {
		orders, err = service.storage.GetOrders(orderIDs)
		if err != nil {
			err = fmt.Errorf("orders: failed to convert fulfilling jobs to order objects (%w)", err)
			service.logger.Error(err)
			return output.ErrorJsonErrInternal(err)
		}
	}

	// build working part of response
	workingResp := make(map[int]*orderJobResponse)
	for workerID, mgrWorkingJob := range mgrJobs.WorkingJobs {
		// find order that matches this workerID, and then make response
		if mgrWorkingJob == nil {
			workingResp[workerID] = nil
		} else {
			for i := range orders {
				if mgrWorkingJob.orderID == orders[i].ID {
					workingResp[workerID] = &orderJobResponse{
						AddedToQueue: int(mgrWorkingJob.addedToQueue.Unix()),
						HighPriority: mgrWorkingJob.IsHighPriority(),
						Order:        orders[i].summaryResponse(service),
					}
					break
				}
			}
		}
	}

	// build waiting part of response
	waitingResp := []orderJobResponse{}
	for _, mgrWaitingJob := range mgrJobs.WaitingJobs {
		for i := range orders {
			if mgrWaitingJob.orderID == orders[i].ID {
				waitingResp = append(waitingResp, orderJobResponse{
					AddedToQueue: int(mgrWaitingJob.addedToQueue.Unix()),
					HighPriority: mgrWaitingJob.IsHighPriority(),
					Order:        orders[i].summaryResponse(service),
				})
				break
			}
		}
	}

	// final response
	jobsResp := &orderWorkStatusResponse{
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
		service.logger.Errorf("orders: failed to write json (%s)", err)
		return output.ErrorJsonErrWriteJsonError(err)
	}
	return nil
}
