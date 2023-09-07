package providers

import "sync"

// Stop calls the Stop() command on each provider
func (ps *Providers) Stop() error {
	return ps.stopOrStart(false)
}

// Start calls the Start() command on each provider
func (ps *Providers) Start() error {
	return ps.stopOrStart(true)
}

// stopOrStart calls the Start() or Stop() command on each provider
func (ps *Providers) stopOrStart(isStart bool) error {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	// if stop, mark unusable
	if !isStart {
		ps.usable = false
	}

	// call shutdown on all async
	var wg sync.WaitGroup
	wgSize := len(ps.pD)

	wg.Add(wgSize)
	wgErrors := make(chan error, wgSize)

	for p := range ps.pD {
		// async shutdown on each
		go func(prov Service) {
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
		ps.usable = true
	}

	return nil
}
