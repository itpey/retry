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

import "time"

// Config defines the config for Retry.
type Config struct {
	MaxAttemptTimes uint          // Maximum number of retries
	InitialBackoff  time.Duration // Initial backoff duration
	MaxBackoff      time.Duration // Maximum backoff duration
	MaxJitter       time.Duration // Jitter duration to add randomness to backoff
}

var ConfigDefault = Config{
	MaxAttemptTimes: 3,                    // Default max retries
	InitialBackoff:  0 * time.Millisecond, // Default initial backoff
	MaxBackoff:      0 * time.Millisecond, // Default max backoff
	MaxJitter:       0 * time.Millisecond, // Default Jitter
}

// Helper function to set default values
func configDefault(config ...Config) Config {
	// Return default config if nothing provided
	if len(config) < 1 {
		return ConfigDefault
	}

	// Override default config
	cfg := config[0]

	return cfg
}
