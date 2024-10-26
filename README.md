[//]: # "Title: RETRY"
[//]: # "Author: itpey"
[//]: # "Attendees: itpey"
[//]: # "Tags: #RETRY #go #Retry #golang #go-lang #try"

<h1 align="center">
RETRY
</h1>

<p align="center">
Flexible and configurable retry mechanism for Go. With RETRY, you can easily implement retry logic with customizable backoff strategies to enhance the resilience of your applications.
</p>

<p align="center">
  <a href="https://pkg.go.dev/github.com/itpey/retry">
    <img src="https://pkg.go.dev/badge/github.com/itpey/retry.svg" alt="Go Reference">
  </a>
  <a href="https://github.com/retry/blob/main/LICENSE">
    <img src="https://img.shields.io/github/license/itpey/retry" alt="license">
  </a>
</p>

# Features

- **Configurable Retries**: Set the maximum number of retries to control how many times an operation is attempted.
- **Backoff Strategies**: Define initial and maximum backoff durations with optional jitter to avoid synchronized retries.
- **Jitter Support**: Add randomness to backoff durations to prevent thundering herd problems.
- **Context Support**: Utilize Go contexts for cancellation and timeouts, ensuring operations can be gracefully terminated.
- **Thread-Safe**: Concurrently modify retry configurations safely across multiple goroutines.

# Installation

Make sure you have Go installed and configured on your system. Use go install to install `RETRY`:

```bash
go get -u github.com/itpey/retry
```

# Usage

## Simple Retry

```go
package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/itpey/retry"
)

func main() {
	r := retry.New(
		retry.Config{
			MaxAttemptTimes: 10,
		})

	fn := func() error {
		return errors.New("temporary error")
	}
	ctx := context.Background()
	if err := r.Do(ctx, fn); err != nil {
		if retry.IsMaxRetriesError(err) {
			fmt.Println("Failed after maximum number of retries")
		} else {
			fmt.Printf("Failed with error: %v\n", err)
		}
	} else {
		fmt.Println("Operation succeeded")
	}
}

```

## Retry with Backoff

```go
package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/itpey/retry"
)

func main() {
	r := retry.New(
		retry.Config{
			MaxAttemptTimes: 5,
			InitialBackoff:  100 * time.Millisecond,
			MaxBackoff:      1 * time.Second,
			MaxJitter:       50 * time.Millisecond,
		})

	fn := func() error {
		return errors.New("temporary error")
	}

	ctx := context.Background()
	if err := r.Do(ctx, fn); err != nil {
		if retry.IsMaxRetriesError(err) {
			fmt.Println("Failed after maximum number of retries")
		} else {
			fmt.Printf("Failed with error: %v\n", err)
		}
	} else {
		fmt.Println("Operation succeeded")
	}

}

```

# Feedback and Contributions

If you encounter any issues or have suggestions for improvement, please [open an issue](https://github.com/itpey/retry/issues) on GitHub.

We welcome contributions! Fork the repository, make your changes, and submit a pull request.

# License

RETRY is open-source software released under the Apache License, Version 2.0. You can find a copy of the license in the [LICENSE](https://github.com/itpey/retry/blob/main/LICENSE) file.

# Author

RETRY was created by [itpey](https://github.com/itpey)
