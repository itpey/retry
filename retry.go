// Copyright 2024 itpey
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package retry

import (
	"context"
	"errors"
	"time"

	"github.com/bytedance/gopkg/lang/fastrand"
)

// Retry holds the configuration for retrying operations.
type Retry struct {
	cfg Config
}

// New creates a new Retry instance with default settings.
func New(config ...Config) *Retry {

	// Set default config
	cfg := configDefault(config...)

	return &Retry{
		cfg: cfg,
	}
}

// Do retries the provided function asynchronously until it either succeeds (returns nil error),
// decides not to retry (returns false), or reaches the maximum number of retries if specified.
func (r *Retry) Do(ctx context.Context, fn func() error) error {
	maxRetries := r.cfg.maxAttemptTimes
	initialBackoff := r.cfg.initialBackoff
	maxBackoff := r.cfg.maxBackoff
	maxJitter := r.cfg.maxJitter

	backoff := initialBackoff
	timer := time.NewTimer(0) // Initialize a timer
	defer timer.Stop()

	for attempt := uint(1); ; attempt++ {
		select {
		case <-ctx.Done():
			return ctx.Err() // Context cancelled or deadline exceeded
		case <-timer.C:
			// Proceed with retry attempt
		}

		err := fn()
		if err == nil {
			return nil // Success, no error
		}
		if attempt >= maxRetries {
			return errors.New("exceeded retry limit") // Exceeded max retries
		}

		// Calculate next backoff duration
		backoff = calculateBackoff(backoff, maxBackoff, maxJitter)
		timer.Reset(backoff) // Reset the timer for the next backoff duration
	}
}

// calculateBackoff calculates the next backoff duration using truncated binary exponential backoff with jitter.
func calculateBackoff(backoff, maxBackoff, maxJitter time.Duration) time.Duration {
	if backoff <= 0 {
		return 0 * time.Millisecond
	}

	// Double the backoff duration
	backoff *= 2

	// Ensure the backoff duration does not exceed the maximum allowed
	if maxBackoff > 0 && backoff > maxBackoff {
		return maxBackoff
	}

	// Calculate jitter
	if maxJitter > 0 {
		// Generate random jitter within [0, jitter)
		jitterValue := time.Duration(fastrand.Int63n(int64(maxJitter)))
		// Apply jitter to the backoff duration
		backoff += jitterValue
	}

	return backoff
}

// IsMaxRetriesError checks if the error indicates that the maximum number of retries has been reached.
func IsMaxRetriesError(err error) bool {
	return err != nil && err.Error() == "exceeded retry limit"
}
