package httpwriter

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/faisalhardin/medilink/internal/library/common/commonerr"
	liblog "github.com/faisalhardin/medilink/internal/library/common/log"
	"github.com/faisalhardin/medilink/internal/library/common/log/logger"
	"github.com/faisalhardin/medilink/internal/library/util/requestinfo"
	"github.com/go-chi/chi/v5"
	"github.com/pkg/errors"
)

type key struct{}

var errCtxKey key

type Response struct {
	Data   interface{}   `json:"data,omitempty"`
	Errors []interface{} `json:"errors,omitempty"`
}

type ErrorMessage struct {
	ErrorMessage []*commonerr.ErrorFormat `json:"error_messages"`
}

func WriteJSONAPIData(w http.ResponseWriter, r *http.Request, status int, data interface{}) (int, error) {
	resp := Response{Data: data}
	return resp.Write(w, r, status)
}

func (resp *Response) Write(w http.ResponseWriter, r *http.Request, status int) (int, error) {
	if resp.Errors != nil {
		resp.Data = nil
		setError(r)
	}

	w.Header().Set("Content-Type", "application/json")
	responseInBytes, err := json.Marshal(resp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		wLen, wErr := w.Write([]byte(`{"errors:["Internal Server Error"]}`))
		if wErr != nil {
			return wLen, wErr
		}
	}

	w.WriteHeader(status)
	return w.Write(responseInBytes)
}

// setError marks that error is written in response
func setError(r *http.Request) {
	ctx := r.Context()
	ctx = context.WithValue(ctx, errCtxKey, true)
	(*r) = *r.WithContext(ctx)
}

// IsError checks whether error was written in response or not
// Can be used by post handler middleware
func IsError(r *http.Request) bool {
	ctx := r.Context()
	if val := ctx.Value(errCtxKey); val != nil {
		return val.(bool)
	}
	return false
}

func SetOKWithData(ctx context.Context, w http.ResponseWriter, data interface{}) (err error) {
	_, err = WriteJSONAPIData(w, nil, http.StatusOK, data)
	return err
}

func SetOKWithByte(ctx context.Context, w http.ResponseWriter, b []byte) (err error) {
	_, err = w.Write(b)
	return
}

func Redirect(ctx context.Context, w http.ResponseWriter, r *http.Request, url string, statusCode int) (err error) {
	http.Redirect(w, r, url, statusCode)

	return nil
}

func SetError(ctx context.Context, w http.ResponseWriter, errValue error) (err error) {

	requestInfo := getRequestInfo(ctx)
	switch errCause := errors.Cause(errValue).(type) {
	case *commonerr.ErrorMessage:
		requestInfo.SetHTTPStatus(errCause.Code)
		err = SetErrorFormat(ctx, w, errCause)
	default:
		requestInfo.SetHTTPStatus(http.StatusInternalServerError)
		_, err = WriteJSON(w, http.StatusInternalServerError, &ErrorMessage{
			ErrorMessage: commonerr.SetNewInternalError().ErrorList,
		})
		go postProcess(ctx, errValue)
	}
	return
}

// SetErrorFormat http
func SetErrorFormat(ctx context.Context, w http.ResponseWriter, errFormat *commonerr.ErrorMessage) (err error) {
	_, err = WriteJSON(w, errFormat.Code, &ErrorMessage{
		ErrorMessage: errFormat.ErrorList,
	})
	return
}

func getRequestInfo(ctx context.Context) requestinfo.RequestInfo {
	ctxData := chi.RouteContext(ctx)
	reqInfo := requestinfo.GetRequestInfo(ctx)
	if ctxData != nil {
		reqInfo.RequestURL = ctxData.RoutePattern()
	}
	return reqInfo
}

func WriteJSON(w http.ResponseWriter, status int, data interface{}) (int, error) {
	w.Header().Set("Content-Type", "application/json")
	b, err := json.Marshal(data)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		writeLen, writeErr := w.Write([]byte(`{"errors":["Internal Server Error"]}`))
		if writeErr != nil {
			return writeLen, writeErr
		}
		return writeLen, err
	}

	w.WriteHeader(status)
	return w.Write(b)
}

func postProcess(ctx context.Context, err error) {

	logData := logger.KV{}
	if err != nil {

		// Report errors
		liblog.ErrorWithFields(err.Error(), logData)
	}

}
