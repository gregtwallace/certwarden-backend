package job_manager

// unsafeJobExists searches for an Equal job in manager. If one is found,
// the worker number it is associated with is returned. If the job is in
// queue without a worker, a negative number is returned. If the job is not
// found, nil is returned.
// Manager MUST be AT LEAST RLocked before callign this func.
func (mgr *Manager[V]) unsafeJobExists(job V) *int {
	// zero value job will never be in manager
	var zeroVal V
	if job.Equal(zeroVal) {
		return nil
	}

	// check workers
	for workerID, mgrJ := range mgr.workingJobs {
		if !mgrJ.Equal(zeroVal) && job.Equal(mgrJ) {
			// copy int so direct access to mgr's int isn't given out
			retVal := new(int)
			*retVal = workerID
			return retVal
		}
	}

	// check waiting
	for i, mgrJ := range mgr.waitingJobs {
		if !mgrJ.Equal(zeroVal) && job.Equal(mgrJ) {
			i *= -1
			return &i
		}
	}

	return nil
}

// JobExists searches for an Equal job in manager. If one is found,
// the worker number it is associated with is returned. If the job is in
// queue without a worker, a negative number is returned. If the job is not
// found, nil is returned.
func (mgr *Manager[V]) JobExists(job V) *int {
	// zero value job will never be in manager
	var zeroVal V
	if job.Equal(zeroVal) {
		return nil
	}

	mgr.RLock()
	defer mgr.RUnlock()

	return mgr.unsafeJobExists(job)
}

// allManagerJobs is a struct to return all of the jobs currently in Manager
type AllManagerJobs[V Job[V]] struct {
	WorkingJobs map[int]V // workerID:job
	WaitingJobs []V
}

// AllCurrentJobs returns all of the jobs in manager. Jobs are separated by those
// currently being worked on, and those waiting in the queue.
func (mgr *Manager[V]) AllCurrentJobs() *AllManagerJobs[V] {
	mgr.RLock()
	defer mgr.RUnlock()

	// working jobs
	workingJobs := make(map[int]V)
	for i, mgrJob := range mgr.workingJobs {
		// copy jobs to new map
		workingJobs[i] = mgrJob
	}

	// waiting (queue) jobs
	waitingJobs := make([]V, len(mgr.waitingJobs))
	_ = copy(waitingJobs, mgr.waitingJobs)

	// return result
	return &AllManagerJobs[V]{
		WorkingJobs: workingJobs,
		WaitingJobs: waitingJobs,
	}
}
