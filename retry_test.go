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
	"math/rand"
	"sync"
	"testing"
	"time"
)

// TestNew tests the default values of a new Retry instance.
func TestNew(t *testing.T) {
	r := New()
	if r.cfg.maxAttemptTimes != 3 {
		t.Errorf("Expected maxRetries to be 3, got %d", r.cfg.maxAttemptTimes)
	}
	if r.cfg.initialBackoff != 0 {
		t.Errorf("Expected initialBackoff to be 0, got %d", r.cfg.initialBackoff)
	}
	if r.cfg.maxBackoff != 0 {
		t.Errorf("Expected maxBackoff to be 0, got %d", r.cfg.maxBackoff)
	}
	if r.cfg.maxJitter != 0 {
		t.Errorf("Expected jitter to be 0, got %d", r.cfg.maxJitter)
	}
}

// TestWithMaxRetries tests setting the maximum number of retries.
func TestWithMaxAttemptTimes(t *testing.T) {
	r := New(Config{maxAttemptTimes: 5})
	if r.cfg.maxAttemptTimes != 5 {
		t.Errorf("Expected maxAttemptTimes to be 5, got %d", r.cfg.maxAttemptTimes)
	}
}

// TestWithBackoff tests setting the backoff parameters.
func TestWithBackoff(t *testing.T) {
	r := New(
		Config{initialBackoff: 100 * time.Millisecond, maxBackoff: 2 * time.Second, maxJitter: 50 * time.Millisecond})
	if r.cfg.initialBackoff != 100*time.Millisecond {
		t.Errorf("Expected initialBackoff to be 100ms, got %d", r.cfg.initialBackoff)
	}
	if r.cfg.maxBackoff != 2*time.Second {
		t.Errorf("Expected maxBackoff to be 2s, got %d", r.cfg.maxBackoff)
	}
	if r.cfg.maxJitter != 50*time.Millisecond {
		t.Errorf("Expected jitter to be 50ms, got %d", r.cfg.maxJitter)
	}
}

// TestDoSuccess tests the retry mechanism with a function that succeeds.
func TestDoSuccess(t *testing.T) {
	r := New()
	fn := func() error {
		return nil
	}
	ctx := context.Background()
	err := r.Do(ctx, fn)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

// TestDoMaxRetries tests the retry mechanism when the maximum number of retries is reached.
func TestDoMaxRetries(t *testing.T) {
	r := New(Config{maxAttemptTimes: 3})
	fn := func() error {
		return errors.New("some error")
	}
	ctx := context.Background()
	err := r.Do(ctx, fn)
	if !IsMaxRetriesError(err) {
		t.Errorf("Expected max retries error, got %v", err)
	}
}

// TestDoFailsTwiceThenSucceeds tests the retry mechanism when the function fails twice before succeeding.
func TestDoFailsTwiceThenSucceeds(t *testing.T) {

	r := New(
		Config{maxAttemptTimes: 5, initialBackoff: 100 * time.Millisecond, maxBackoff: 1 * time.Second, maxJitter: 50 * time.Millisecond})
	attempts := 0
	fn := func() error {
		attempts++
		if attempts <= 2 {
			return errors.New("some error")
		}
		return nil
	}
	ctx := context.Background()
	err := r.Do(ctx, fn)
	if err != nil {
		t.Errorf("Expected no error after retries, got %v", err)
	}
	if attempts != 3 {
		t.Errorf("Expected 3 attempts, got %d", attempts)
	}
}

// TestCalculateBackoff tests the calculateBackoff function.
func TestCalculateBackoff(t *testing.T) {
	backoff := 100 * time.Millisecond
	maxBackoff := 1 * time.Second
	jitter := 50 * time.Millisecond
	newBackoff := calculateBackoff(backoff, maxBackoff, jitter)
	if newBackoff <= 100*time.Millisecond || newBackoff > 1*time.Second+50*time.Millisecond {
		t.Errorf("Backoff is out of expected range, got %v", newBackoff)
	}
}

// TestBackoffDisabled tests the retry mechanism with backoff disabled.
func TestBackoffDisabled(t *testing.T) {
	r := New(Config{maxAttemptTimes: 3})
	fn := func() error {
		return errors.New("some error")
	}
	ctx := context.Background()
	start := time.Now()
	err := r.Do(ctx, fn)
	duration := time.Since(start)
	if !IsMaxRetriesError(err) {
		t.Errorf("Expected max retries error, got %v", err)
	}
	if duration >= 100*time.Millisecond {
		t.Errorf("Expected duration to be less than 100ms, got %v", duration)
	}
}

// TestBackoffEnabled tests the retry mechanism with backoff enabled.
func TestBackoffEnabled(t *testing.T) {
	r := New(Config{maxAttemptTimes: 3, initialBackoff: 100 * time.Millisecond, maxBackoff: 1 * time.Second, maxJitter: 50 * time.Millisecond})
	fn := func() error {
		return errors.New("some error")
	}
	ctx := context.Background()
	start := time.Now()
	err := r.Do(ctx, fn)
	duration := time.Since(start)
	if !IsMaxRetriesError(err) {
		t.Errorf("Expected max retries error, got %v", err)
	}
	if duration < 200*time.Millisecond || duration > 1*time.Second+100*time.Millisecond {
		t.Errorf("Expected duration to be within range, got %v", duration)
	}
}

// TestContextCancelled tests the retry mechanism when the context is cancelled.
func TestContextCancelled(t *testing.T) {
	r := New(Config{maxAttemptTimes: 10, initialBackoff: 100 * time.Millisecond, maxBackoff: 1 * time.Second, maxJitter: 50 * time.Millisecond})
	fn := func() error {
		return errors.New("some error")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()
	err := r.Do(ctx, fn)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("Expected context deadline exceeded error, got %v", err)
	}
}

// Mock function to simulate a task that may fail randomly
func mockTask(successAfter int) func() error {
	attempts := 0
	return func() error {
		attempts++
		if attempts < successAfter {
			return errors.New("temporary error")
		}
		return nil // Success after the specified number of attempts
	}
}

func TestRetryThreadSafety(t *testing.T) {
	const numGoroutines = 100
	const maxAttempts = 5
	var wg sync.WaitGroup
	successCount := 0
	failureCount := 0

	// Create a Retry instance with a maximum of maxAttempts
	retry := New(Config{
		maxAttemptTimes: maxAttempts,
		initialBackoff:  10 * time.Millisecond,
		maxBackoff:      100 * time.Millisecond,
		maxJitter:       10 * time.Millisecond,
	})

	// Use a WaitGroup to wait for all goroutines to finish
	wg.Add(numGoroutines)

	results := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			ctx := context.Background()
			// Randomly decide how many attempts it will take to succeed
			successAfter := rand.Intn(maxAttempts) + 1
			err := retry.Do(ctx, mockTask(successAfter))
			results <- err
		}()
	}

	// Wait for all goroutines to finish
	wg.Wait()
	close(results)

	// Analyze results
	for err := range results {
		if err == nil {
			successCount++
		} else if IsMaxRetriesError(err) {
			failureCount++
		} else {
			t.Errorf("unexpected error: %v", err)
		}
	}

	// Check that the number of successes and failures is as expected
	if successCount+failureCount != numGoroutines {
		t.Errorf("expected %d total results, got %d (successes: %d, failures: %d)", numGoroutines, successCount+failureCount, successCount, failureCount)
	}
}
