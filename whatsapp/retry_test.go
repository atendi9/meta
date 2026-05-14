package whatsapp

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/atendi9/capivara/assert"
)

func TestRetry_SuccessOnFirstAttempt(t *testing.T) {
	ctx := context.Background()
	fn := func() error {
		return nil
	}

	retries, err := Retry(ctx, 3, time.Millisecond, fn)

	assert.NoError(t, err)
	assert.Equal(t, 0, retries)
}

func TestRetry_SuccessAfterFewAttempts(t *testing.T) {
	ctx := context.Background()
	attempts := 0
	fn := func() error {
		if attempts < 2 {
			attempts++
			return errors.New("temporary connection error")
		}
		return nil
	}

	retries, err := Retry(ctx, 3, time.Millisecond, fn)

	assert.NoError(t, err)
	assert.Equal(t, 2, retries)
}

func TestRetry_FailsAllAttempts(t *testing.T) {
	ctx := context.Background()
	expectedErr := errors.New("permanent failure")
	fn := func() error {
		return expectedErr
	}

	retries, err := Retry(ctx, 3, time.Millisecond, fn)

	assert.Error(t, err)
	assert.Equal(t, 3, retries)
}

func TestRetry_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	attempts := 0

	fn := func() error {
		attempts++
		if attempts == 1 {
			cancel()
		}
		return errors.New("some error")
	}

	retries, err := Retry(ctx, 5, time.Hour, fn)

	assert.Error(t, err)
	assert.Equal(t, 1, retries)
}
