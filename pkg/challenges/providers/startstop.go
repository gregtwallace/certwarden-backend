package providers

import "sync"

// Stop calls the Stop() command on each provider
func (mgr *Manager) Stop() error {
	return mgr.stopOrStart(false)
}

// Start calls the Start() command on each provider
func (mgr *Manager) Start() error {
	return mgr.stopOrStart(true)
}

// stopOrStart calls the Start() or Stop() command on each provider
func (mgr *Manager) stopOrStart(isStart bool) error {
	mgr.mu.RLock()
	defer mgr.mu.RUnlock()

	// if stop, mark unusable
	if !isStart {
		mgr.usable = false
	}

	// call shutdown on all async
	var wg sync.WaitGroup
	wgSize := len(mgr.pD)

	wg.Add(wgSize)
	wgErrors := make(chan error, wgSize)

	for p := range mgr.pD {
		// async shutdown on each
		go func(prov *provider) {
			defer wg.Done()
			var err error

			// start or stop
			if isStart {
				err = prov.Start()
			} else {
				err = prov.Stop()
			}
			if err != nil {
				wgErrors <- err
			}
		}(p)
	}

	// wait for all
	wg.Wait()

	// check for errors
	close(wgErrors)
	for err := range wgErrors {
		if err != nil {
			return err
		}
	}

	// if success starting, is usable
	if isStart {
		mgr.usable = true
	}

	return nil
}
