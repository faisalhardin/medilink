package satusehat

import (
	"fmt"

	"github.com/faisalhardin/medilink/internal/entity/model/satusehat"
)

// Common Satu Sehat API error types
var (
	ErrUnauthorized       = fmt.Errorf("unauthorized: invalid or expired credentials")
	ErrForbidden          = fmt.Errorf("forbidden: insufficient permissions")
	ErrNotFound           = fmt.Errorf("resource not found")
	ErrBadRequest         = fmt.Errorf("bad request: invalid input")
	ErrConflict           = fmt.Errorf("conflict: resource already exists")
	ErrServerError        = fmt.Errorf("server error")
	ErrRateLimitExceeded  = fmt.Errorf("rate limit exceeded")
	ErrInvalidToken       = fmt.Errorf("invalid or expired access token")
	ErrMissingCredentials = fmt.Errorf("missing client credentials")
)

// FHIRError represents a FHIR OperationOutcome error
type FHIRError struct {
	StatusCode int
	Severity   string
	Code       string
	Details    string
	Diagnostic string
	Expression []string
}

func (e *FHIRError) Error() string {
	if e.Diagnostic != "" {
		return fmt.Sprintf("FHIR %s (%d): %s - %s", e.Severity, e.StatusCode, e.Code, e.Diagnostic)
	}
	return fmt.Sprintf("FHIR %s (%d): %s", e.Severity, e.StatusCode, e.Code)
}

// ParseOperationOutcome converts an OperationOutcome to a FHIRError
func ParseOperationOutcome(outcome *satusehat.OperationOutcome, statusCode int) error {
	if outcome == nil || len(outcome.Issue) == 0 {
		return fmt.Errorf("unknown error with status code %d", statusCode)
	}

	// Get the first (most severe) issue
	issue := outcome.Issue[0]

	return &FHIRError{
		StatusCode: statusCode,
		Severity:   issue.Severity,
		Code:       issue.Code,
		Diagnostic: issue.Diagnostics,
		Expression: issue.Expression,
	}
}

// IsRetryable checks if an error is retryable
func IsRetryable(err error) bool {
	if err == nil {
		return false
	}

	fhirErr, ok := err.(*FHIRError)
	if !ok {
		return false
	}

	// Retry on server errors (5xx) and rate limiting (429)
	return fhirErr.StatusCode >= 500 || fhirErr.StatusCode == 429
}

// IsAuthError checks if an error is authentication-related
func IsAuthError(err error) bool {
	if err == nil {
		return false
	}

	fhirErr, ok := err.(*FHIRError)
	if !ok {
		return false
	}

	return fhirErr.StatusCode == 401 || fhirErr.StatusCode == 403
}

// IsNotFoundError checks if an error indicates a resource was not found
func IsNotFoundError(err error) bool {
	if err == nil {
		return false
	}

	fhirErr, ok := err.(*FHIRError)
	if !ok {
		return false
	}

	return fhirErr.StatusCode == 404
}

// IsValidationError checks if an error is a validation error
func IsValidationError(err error) bool {
	if err == nil {
		return false
	}

	fhirErr, ok := err.(*FHIRError)
	if !ok {
		return false
	}

	return fhirErr.StatusCode == 400 || fhirErr.StatusCode == 422
}
