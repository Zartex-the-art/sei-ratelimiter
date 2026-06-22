package handlers

import (
	"context"
	"strings"
)

func isInfraError(err error) bool {
	if err == nil {
		return false
	}

	if isErr(err, context.DeadlineExceeded) {
		return true
	}

	if isErr(err, context.Canceled) {
		return true
	}

	s := err.Error()

	infraSignals := []string{
		"connection refused",
		"dial tcp",
		"i/o timeout",
		"EOF",
		"broken pipe",
		"no such host",
		"redis: client is closed",
		"NOSCRIPT",
	}

	for _, sig := range infraSignals {
		if strings.Contains(s, sig) {
			return true
		}
	}

	return false
}

func isErr(err, target error) bool {
	if err == nil {
		return false
	}

	return err.Error() == target.Error()
}
