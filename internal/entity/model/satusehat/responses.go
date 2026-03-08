package satusehat

// Response structures and error handling for Satu Sehat FHIR API
// Based on https://www.hl7.org/fhir/operationoutcome.html

// OperationOutcome - Information about the success/failure of an action
type OperationOutcome struct {
	DomainResource
	Issue []OperationOutcomeIssue `json:"issue"` // A single issue associated with the action (REQUIRED)
}

// OperationOutcomeIssue - A single issue associated with the action
type OperationOutcomeIssue struct {
	BackboneElement
	Severity    string           `json:"severity"`              // fatal | error | warning | information (REQUIRED)
	Code        string           `json:"code"`                  // Error or warning code (REQUIRED)
	Details     *CodeableConcept `json:"details,omitempty"`     // Additional details about the error
	Diagnostics string           `json:"diagnostics,omitempty"` // Additional diagnostic information about the issue
	Location    []string         `json:"location,omitempty"`    // Deprecated: Path of element(s) related to issue
	Expression  []string         `json:"expression,omitempty"`  // FHIRPath of element(s) related to issue
}

// TokenResponse - OAuth2 token response for authentication
type TokenResponse struct {
	AccessToken string `json:"access_token"` // The access token
	TokenType   string `json:"token_type"`   // Type of token (usually "Bearer")
	ExpiresIn   int    `json:"expires_in"`   // Seconds until the token expires
	Scope       string `json:"scope,omitempty"` // The scope of the token
}

// ErrorResponse - Standard error response wrapper
type ErrorResponse struct {
	ResourceType string                `json:"resourceType"` // Always "OperationOutcome"
	Issue        []OperationOutcomeIssue `json:"issue"`      // List of issues
}

// SearchResponse - Generic search response wrapper
// The actual Bundle should be used for search responses, this is just a convenience type
type SearchResponse struct {
	Bundle
}

// CreateResponse - Response wrapper for resource creation
// Contains the created resource with server-assigned id and meta
type CreateResponse struct {
	ResourceType string `json:"resourceType"` // Type of the created resource
	ID           string `json:"id"`           // Server-assigned resource ID
	Meta         *Meta  `json:"meta"`         // Metadata about the resource
}

// ValidationResponse - Response for resource validation
type ValidationResponse struct {
	OperationOutcome
	Valid bool `json:"-"` // Whether validation passed (derived from issues)
}

// BatchResponse - Response for batch operations
// Uses Bundle with type "batch-response"
type BatchResponse struct {
	Bundle
}

// TransactionResponse - Response for transaction operations
// Uses Bundle with type "transaction-response"
type TransactionResponse struct {
	Bundle
}
