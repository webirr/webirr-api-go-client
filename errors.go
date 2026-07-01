package webirr

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
)

// HTTPError is returned for non-2xx HTTP responses.
type HTTPError struct {
	StatusCode int
	Status     string
}

type transientErrors struct{}

// TransientErrors classifies platform errors that are normally safe to retry.
var TransientErrors transientErrors

func (e *HTTPError) Error() string {
	if e.Status != "" {
		return fmt.Sprintf("webirr http error: %s", e.Status)
	}
	return fmt.Sprintf("webirr http error: %d", e.StatusCode)
}

// IsTransient reports whether the HTTP status is normally safe to retry.
func (e *HTTPError) IsTransient() bool {
	if e == nil {
		return false
	}
	return e.StatusCode == http.StatusRequestTimeout ||
		e.StatusCode == http.StatusTooManyRequests ||
		e.StatusCode >= 500
}

// IsTransient reports whether an SDK error is normally safe to retry.
func (transientErrors) IsTransient(err error) bool {
	if err == nil {
		return false
	}

	var httpError *HTTPError
	if errors.As(err, &httpError) {
		return httpError.IsTransient()
	}

	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}

	var netError net.Error
	return errors.As(err, &netError) && netError.Timeout()
}
