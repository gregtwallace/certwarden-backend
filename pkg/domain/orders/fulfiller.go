package orders

import (
	"context"
	"fmt"
	"legocerthub-backend/pkg/domain/acme_servers"
	"legocerthub-backend/pkg/domain/authorizations"
	"legocerthub-backend/pkg/httpclient"
	"os/exec"
	"sync"

	"go.uber.org/zap"
)

// orderFulfiller works Orders with the ACME server in order to get
// certificates issued. It tracks which orders have been requested, which
// are being worked, and with which worker the orders are with.
type orderFulfiller struct {
	shutdownContext   context.Context
	logger            *zap.SugaredLogger
	storage           Storage
	acmeServerService *acme_servers.Service
	authorizations    *authorizations.Service

	isHttps               bool
	serverCertificateName *string
	loadHttpsCertificate  func() error
	httpClient            *httpclient.Client
	shellPath             string

	highJobs chan orderFulfillerJob
	lowJobs  chan orderFulfillerJob
	mu       sync.RWMutex

	// actual data
	jobsWaiting []orderFulfillerJob
	workerJobs  map[int]*orderFulfillerJob // [workerid]
}

// CreateManager creates the manager with the requested number of workers
func createOrderFulfiller(app App, workerCount int) *orderFulfiller {
	logger := app.GetLogger()

	// determine shell (os dependent)
	// powershell
	var shellPath string
	var err error
	shellPath, err = exec.LookPath("powershell.exe")
	if err != nil {
		logger.Debugf("unable to find powershell (%s)", err)
		// then try bash
		shellPath, err = exec.LookPath("bash")
		if err != nil {
			logger.Debugf("unable to find bash (%s)", err)
			// then try zshell
			shellPath, err = exec.LookPath("zsh")
			if err != nil {
				logger.Debugf("unable to find zshell (%s)", err)
				// then try sh
				shellPath, err = exec.LookPath("sh")
				if err != nil {
					logger.Debugf("unable to find sh (%s)", err)
					// failed - disable post processing
					logger.Errorf("unable to find a suitable shell for certificate post processing actions, post processing disabled (%s)")
					shellPath = ""
				}
			}
		}
	}

	of := &orderFulfiller{
		shutdownContext:   app.GetShutdownContext(),
		logger:            logger,
		storage:           app.GetOrderStorage(),
		acmeServerService: app.GetAcmeServerService(),
		authorizations:    app.GetAuthsService(),

		isHttps:               app.IsHttps(),
		serverCertificateName: app.HttpsCertificateName(),
		loadHttpsCertificate:  app.LoadHttpsCertificate,
		httpClient:            app.GetHttpClient(),
		shellPath:             shellPath,

		workerJobs: make(map[int]*orderFulfillerJob),
		highJobs:   make(chan orderFulfillerJob),
		lowJobs:    make(chan orderFulfillerJob),
	}

	wg := app.GetShutdownWaitGroup()
	shutdownContext := app.GetShutdownContext()

	// make workers
	for i := 0; i < workerCount; i++ {
		// make entry on map for work tracking
		of.workerJobs[i] = nil

		// add to wg
		wg.Add(1)
		defer wg.Done()

		// start worker func w/ id
		go func(workerId int) {
			// spawn worker
			of.logger.Debugf("worker %d: starting", workerId)

		doingWork:
			for {
				select {
				case <-shutdownContext.Done():
					// break to shutdown
					break doingWork
				case highJob := <-of.highJobs:
					of.doJob(highJob, workerId)
					of.logger.Debugf("worker %d: end of high priority order fulfiller for order id %d (certificate name: %s, subject: %s)", workerId, highJob.order.ID, highJob.order.Certificate.Name, highJob.order.Certificate.Subject)
				case lowJob := <-of.lowJobs:
				lower:
					for {
						select {
						case <-shutdownContext.Done():
							// break to shutdown
							break doingWork
						case highJob := <-of.highJobs:
							of.doJob(highJob, workerId)
							of.logger.Debugf("worker %d: end of high priority order fulfiller for order id %d (certificate name: %s, subject: %s)", workerId, highJob.order.ID, highJob.order.Certificate.Name, highJob.order.Certificate.Subject)
						default:
							break lower
						}
					}
					of.doJob(lowJob, workerId)
					of.logger.Debugf("worker %d: end of low priority order fulfiller for order id %d (certificate name: %s, subject: %s)", workerId, lowJob.order.ID, lowJob.order.Certificate.Name, lowJob.order.Certificate.Subject)
				}
			}

			of.logger.Debugf("worker %d: shutdown complete", workerId)
		}(i)

	}

	return of
}

// moveJobToWorking moves the specified job from waiting to the tracking map for workers
func (of *orderFulfiller) moveJobToWorking(j orderFulfillerJob, workerId int) (err error) {
	of.mu.Lock()
	defer of.mu.Unlock()

	// verify in waiting
	for i := range of.jobsWaiting {
		if of.jobsWaiting[i].order.ID == j.order.ID {

			// verify worker is free to take the job (should never have this error)
			if of.workerJobs[workerId] != nil {
				return fmt.Errorf("cannot add order id %d to worker id %d (worker already has order id %d)", j.order.ID, workerId, of.workerJobs[workerId].order.ID)
			}

			// update worker
			of.workerJobs[workerId] = &j

			// and remove from waiting
			of.jobsWaiting[i] = of.jobsWaiting[(len(of.jobsWaiting) - 1)]
			of.jobsWaiting = of.jobsWaiting[:len(of.jobsWaiting)-1]

			return nil
		}
	}

	// did not find job in waiting (should also never happen)
	return fmt.Errorf("cannot move order id %d to worker id %d (job not in waiting)", j.order.ID, workerId)
}

// removeJob removes the specified job from the worker tracking map
func (of *orderFulfiller) removeJob(j orderFulfillerJob, workerId int) (err error) {
	of.mu.Lock()
	defer of.mu.Unlock()

	// if job id is with specified worker id, remove it
	if of.workerJobs[workerId] != nil && of.workerJobs[workerId].order.ID == j.order.ID {
		of.workerJobs[workerId] = nil

		return nil
	}

	return fmt.Errorf("cannot remove order id %d from worker %d (worker doesn't have specified order)", j.order.ID, workerId)
}
