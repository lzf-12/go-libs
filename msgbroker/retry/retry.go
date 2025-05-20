package retry

import (
	"math"
	"time"
)

type RetryPolicy struct {
	MaxRetries      int
	InitialInterval time.Duration
	Multiplier      float64
	MaxInterval     time.Duration
}

func WithBackoff(policy RetryPolicy, fn func() error) error {
	var err error
	for i := 0; i < policy.MaxRetries; i++ {
		err = fn()
		if err == nil {
			return nil
		}

		if i == policy.MaxRetries-1 {
			break
		}

		wait := calculateWait(policy, i)
		time.Sleep(wait)
	}
	return err
}

func calculateWait(policy RetryPolicy, attempt int) time.Duration {
	wait := float64(policy.InitialInterval) * math.Pow(policy.Multiplier, float64(attempt))
	max := float64(policy.MaxInterval)
	if wait > max {
		wait = max
	}
	return time.Duration(wait)
}
