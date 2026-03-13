package satusehat

import "encoding/json"

// FHIR Framework types
// Based on https://satusehat.kemkes.go.id/platform/docs/id/fhir/framework/

// Resource - Base Resource
type Resource struct {
	ResourceType  string      `json:"resourceType"`            // Type name of the resource
	ID            string      `json:"id,omitempty"`            // Logical id of this artifact
	Meta          *Meta       `json:"meta,omitempty"`          // Metadata about the resource
	ImplicitRules string      `json:"implicitRules,omitempty"` // A set of rules under which this content was created
	Language      string      `json:"language,omitempty"`      // Language of the resource content (BCP-47)
}

// DomainResource - Parent type for domain resources (resources with narrative)
type DomainResource struct {
	Resource
	Text              *Narrative        `json:"text,omitempty"`              // Text summary of the resource, for human interpretation
	Contained         []json.RawMessage `json:"contained,omitempty"`         // Contained, inline Resources
	Extension         []Extension       `json:"extension,omitempty"`         // Additional content defined by implementations
	ModifierExtension []Extension       `json:"modifierExtension,omitempty"` // Extensions that cannot be ignored
}

// Bundle - Contains a collection of resources
type Bundle struct {
	Resource
	Identifier *Identifier    `json:"identifier,omitempty"` // Persistent identifier for the bundle
	Type       string         `json:"type"`                 // document | message | transaction | transaction-response | batch | batch-response | history | searchset | collection
	Timestamp  string         `json:"timestamp,omitempty"`  // When the bundle was assembled
	Total      *int           `json:"total,omitempty"`      // If search, the total number of matches
	Link       []BundleLink   `json:"link,omitempty"`       // Links related to this Bundle
	Entry      []BundleEntry  `json:"entry,omitempty"`      // Entry in the bundle - will have a resource or information
	Signature  *Signature     `json:"signature,omitempty"`  // Digital Signature
}

// BundleLink - Links related to this Bundle
type BundleLink struct {
	Relation string `json:"relation"` // See http://www.iana.org/assignments/link-relations/link-relations.xhtml
	URL      string `json:"url"`      // Reference details for the link
}

// BundleEntry - Entry in the bundle - will have a resource or information
type BundleEntry struct {
	Link     []BundleLink         `json:"link,omitempty"`     // Links related to this entry
	FullURL  string               `json:"fullUrl,omitempty"`  // URI for resource (Absolute URL server address or URI for UUID/OID)
	Resource json.RawMessage      `json:"resource,omitempty"` // A resource in the bundle
	Search   *BundleEntrySearch   `json:"search,omitempty"`   // Search related information
	Request  *BundleEntryRequest  `json:"request,omitempty"`  // Additional execution information (transaction/batch/history)
	Response *BundleEntryResponse `json:"response,omitempty"` // Results of execution (transaction/batch/history)
}

// BundleEntrySearch - Search related information
type BundleEntrySearch struct {
	Mode  string   `json:"mode,omitempty"`  // match | include | outcome - why this is in the result set
	Score *float64 `json:"score,omitempty"` // Search ranking (between 0 and 1)
}

// BundleEntryRequest - Additional execution information (transaction/batch/history)
type BundleEntryRequest struct {
	Method          string `json:"method"`                    // GET | HEAD | POST | PUT | DELETE | PATCH
	URL             string `json:"url"`                       // URL for HTTP equivalent of this entry
	IfNoneMatch     string `json:"ifNoneMatch,omitempty"`     // For managing cache currency
	IfModifiedSince string `json:"ifModifiedSince,omitempty"` // For managing cache currency
	IfMatch         string `json:"ifMatch,omitempty"`         // For managing update contention
	IfNoneExist     string `json:"ifNoneExist,omitempty"`     // For conditional creates
}

// BundleEntryResponse - Results of execution (transaction/batch/history)
type BundleEntryResponse struct {
	Status       string          `json:"status"`                 // Status response code (text optional)
	Location     string          `json:"location,omitempty"`     // The location (if the operation returns a location)
	Etag         string          `json:"etag,omitempty"`         // The Etag for the resource (if relevant)
	LastModified string          `json:"lastModified,omitempty"` // Server's date time modified
	Outcome      json.RawMessage `json:"outcome,omitempty"`      // OperationOutcome with hints and warnings (for batch/transaction)
}
