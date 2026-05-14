package whatsapp

import (
	"context"
	"time"
)

// DefaultRetryDuration represents the default duration of type [time.Duration] to wait between retry attempts.
const DefaultRetryDuration time.Duration = 2 * time.Second

// Retry executes the provided function up to maxRetries times.
//   - It waits for the specified [time.Duration] between attempts.
//   - It returns the number of failed attempts and the last error encountered.
//   - It respects the [context.Context] for cancellation, stopping immediately if the context is done.
func Retry(
	ctx context.Context,
	maxRetries int,
	duration time.Duration,
	fn func() error,
) (int, error) {
	var err error
	retryCount := 0

	for retryCount < maxRetries {
		err = fn()
		if err == nil {
			return retryCount, nil
		}

		retryCount++
		if retryCount >= maxRetries {
			break
		}

		select {
		case <-ctx.Done():
			return retryCount, ctx.Err()
		case <-time.After(duration):
			continue
		}
	}

	return retryCount, err
}
