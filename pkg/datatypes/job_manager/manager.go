package job_manager

import (
	"context"
	"sync"

	"go.uber.org/zap"
)

// Job is the interface that the external job struct will need to satisfy
type Job[V any] interface {
	// Description should return information to help identify a specific job in the logs
	// (e.g. a certificate's CN)
	Description() string

	// IsHighPriority returns if the job should be considered high priority
	IsHighPriority() bool

	// Equal compares two jobs to determine if the job should be considered duplicate
	// and therefore not be added if it has already been added
	Equal(job V) bool

	// Do is the actual work the job does
	Do(workerID int)
}

// Manager manages jobs and their interaction with the workers
type Manager[V Job[V]] struct {
	// readable list of all jobs in the manager
	workingJobs map[int]V // workerID:job
	waitingJobs []V

	// channels to send work to workers
	highJobsChan chan V
	lowJobsChan  chan V

	sync.RWMutex
}

// NewManager creates the job manager and its workers
func NewManager[V Job[V]](workerCount int, workLabel string, shutdownCtx context.Context, shutdownWg *sync.WaitGroup, logger *zap.SugaredLogger) *Manager[V] {
	// if workerCount is invalid, return nil
	if workerCount <= 0 {
		return nil
	}

	// make manager
	mgr := &Manager[V]{
		workingJobs: make(map[int]V),

		highJobsChan: make(chan V),
		lowJobsChan:  make(chan V),
	}

	// make workers
	for i := 0; i < workerCount; i++ {
		// make entry on map for worker tracking
		var zeroVal V
		mgr.workingJobs[i] = zeroVal

		// start worker func w/ id
		shutdownWg.Add(1)
		go func(workerId int) {
			// spawn worker
			defer shutdownWg.Done()
			logger.Debugf("%s worker %d: started", workLabel, workerId)

		doingWork:
			for {
				select {
				case <-shutdownCtx.Done():
					// break to shutdown
					break doingWork

				case highJob := <-mgr.highJobsChan:
					logger.Debugf("%s worker %d: start high priority job (%s)", workLabel, workerId, highJob.Description())
					mgr.doJob(highJob, workerId)
					logger.Debugf("%s worker %d: end high priority job (%s)", workLabel, workerId, highJob.Description())

				case lowJob := <-mgr.lowJobsChan:
				lower:
					for {
						select {
						case <-shutdownCtx.Done():
							// break to shutdown
							break doingWork

						case highJob := <-mgr.highJobsChan:
							logger.Debugf("%s worker %d: start high priority job (%s)", workLabel, workerId, highJob.Description())
							mgr.doJob(highJob, workerId)
							logger.Debugf("%s worker %d: end high priority job (%s)", workLabel, workerId, highJob.Description())

						default:
							break lower
						}
					}

					logger.Debugf("%s worker %d: start low priority job (%s)", workLabel, workerId, lowJob.Description())
					mgr.doJob(lowJob, workerId)
					logger.Debugf("%s worker %d: end low priority job (%s)", workLabel, workerId, lowJob.Description())
				}
			}

			logger.Debugf("%s worker %d: shutdown complete", workLabel, workerId)
		}(i)

	}

	return mgr
}
