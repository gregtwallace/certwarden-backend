package job_manager

import (
	"errors"
)

var (
	ErrAddDuplicateJob = errors.New("job manager: cant add job (already exists)")
	ErrAddZeroValueJob = errors.New("job manager: cant add job with zero value")
)

// AddJob adds the specified job to the manager. It uses the underlying job's
// Equal() function to determine if the job already exists in manager. If the
// job already exists, it is not added again.
func (mgr *Manager[V]) AddJob(job V) error {
	// fail if zeroVal
	var zeroVal V
	if job.Equal(zeroVal) {
		return ErrAddZeroValueJob
	}

	mgr.Lock()
	defer mgr.Unlock()

	// check for equivelant job using job's equal func
	workerNumb := mgr.unsafeJobExists(job)
	if workerNumb != nil {
		return ErrAddDuplicateJob
	}

	// add to work queue
	mgr.waitingJobs = append(mgr.waitingJobs, job)

	// send ID to the appropriate channel
	// async required as this will block until other end of channel reads the job
	go func() {
		if job.IsHighPriority() {
			mgr.highJobsChan <- job
		} else {
			mgr.lowJobsChan <- job
		}
	}()

	return nil
}
