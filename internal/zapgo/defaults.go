package zapgo

import (
	"time"
)

// Wait durations for a response from the Docker daemon before returning an error to the caller.
const (
	// TimeoutInS is the wait duration for common operations.
	TimeoutInS = 30 * time.Second

	// LongTimeoutInS is the wait duration for long operations such as PullImage.
	LongTimeoutInS = 120 * time.Second
)
