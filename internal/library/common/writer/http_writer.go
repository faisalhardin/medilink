package httpwriter

import (
	"context"
	"encoding/json"
	"net/http"
)

type key struct{}

var errCtxKey key

type Response struct {
	Data   interface{}   `json:"data,omitempty"`
	Errors []interface{} `json:"errors,omitempty"`
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
