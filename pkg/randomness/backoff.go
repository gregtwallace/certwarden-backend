package randomness

import (
	"context"
	"time"

	"github.com/cenkalti/backoff/v4"
)

// BackoffACME returns a backoff to be used for ACME operations while waiting for a
// remote ACME service to finish working on some task
// RFC8555 mentions 5 - 10 seconds when discussing waiting on challenges, so use 7
// seconds as the starting point
func BackoffACME(maxElapsedTime time.Duration, shutdownCtx context.Context) backoff.BackOffContext {
	bo := backoff.NewExponentialBackOff()
	bo.InitialInterval = 7 * time.Second
	bo.RandomizationFactor = 0.4
	bo.Multiplier = 1.4
	bo.MaxInterval = 60 * time.Second
	bo.MaxElapsedTime = maxElapsedTime

	boWithContext := backoff.WithContext(bo, shutdownCtx)

	return boWithContext
}
