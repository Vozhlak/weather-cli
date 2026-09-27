package retry

import (
	"context"
	"fmt"
	"time"
)

func Do(ctx context.Context, attempts int, baseDelay time.Duration, fn func() error) error {
	if attempts <= 0 {
		return fmt.Errorf("attempts must be positive")
	}

	delay := baseDelay
	var err error

	for i := 0; i < attempts; i++ {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		err = fn()
		if err == nil {
			return nil
		}

		if i < attempts-1 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
			}

			delay *= 2
		}

	}

	return err
}
