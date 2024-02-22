package job_manager

// do updates the job to assign it to a worker and then executes the internal 'real'
// job
func (mgr *Manager[V]) doJob(job V, workerID int) {
	// move job from waiting to working
	mgr.Lock()
	for i, waitingJ := range mgr.waitingJobs {
		if waitingJ.Equal(job) {
			// remove from waiting
			mgr.waitingJobs[i] = mgr.waitingJobs[len(mgr.waitingJobs)-1]
			mgr.waitingJobs = mgr.waitingJobs[:len(mgr.waitingJobs)-1]

			// add to worker
			mgr.workingJobs[workerID] = job
		}

	}
	mgr.Unlock()

	// run job
	job.Do(workerID)

	// after job completes, remove it from worker
	mgr.Lock()
	var zeroVal V
	mgr.workingJobs[workerID] = zeroVal
	mgr.Unlock()
}
