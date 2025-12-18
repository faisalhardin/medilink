package odontogram

import (
	"net/http"
	"strconv"

	"github.com/faisalhardin/medilink/internal/entity/model"
	odontogramuc "github.com/faisalhardin/medilink/internal/entity/usecase/odontogram"
	commonwriter "github.com/faisalhardin/medilink/internal/library/common/writer"
	"github.com/faisalhardin/medilink/internal/library/util/common/binding"
)

var (
	bindingBind = binding.Bind
)

// OdontogramHandler handles HTTP requests for odontogram operations
type OdontogramHandler struct {
	OdontogramUC odontogramuc.OdontogramUC
}

// New creates a new odontogram HTTP handler
func New(handler *OdontogramHandler) *OdontogramHandler {
	return handler
}

// CreateEvents handles POST /odontogram/logs
func (h *OdontogramHandler) CreateEvents(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse request body
	var requests model.BulkCreateOdontogramEventRequest
	err := bindingBind(r, &requests)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	// Create events
	_, err = h.OdontogramUC.CreateEvents(ctx, requests.Events)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	commonwriter.SetOKWithData(ctx, w, "OK")
}

// GetEvents handles GET /odontogram/logs
func (h *OdontogramHandler) GetEvents(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse query parameters
	var params model.GetOdontogramEventsParams
	if err := bindingBind(r, &params); err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	// Get events
	response, err := h.OdontogramUC.GetEvents(ctx, params)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	commonwriter.SetOKWithData(ctx, w, response)
}

// GetSnapshot handles GET /odontogram
func (h *OdontogramHandler) GetSnapshot(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse query parameters
	var params model.GetOdontogramSnapshotParams
	if err := bindingBind(r, &params); err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	// Check If-None-Match header from client
	ifNoneMatch := r.Header.Get("If-None-Match")

	// Get snapshot
	response, err := h.OdontogramUC.GetSnapshot(ctx, params)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	// Set ETag header based on max_sequence_number
	currentETag := strconv.FormatInt(response.MaxSequenceNumber, 10)
	w.Header().Set("ETag", currentETag)

	// If client's ETag matches current ETag, return 304 Not Modified
	if ifNoneMatch != "" && ifNoneMatch == currentETag {
		w.WriteHeader(http.StatusNotModified)
		return // No body sent - saves bandwidth!
	}

	commonwriter.SetOKWithData(ctx, w, response)
}
