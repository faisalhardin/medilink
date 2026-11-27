package requestinfo

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSetRequestContextNeverStoresRawPath(t *testing.T) {
	nikPath := "/v1/patients/3201012345670001"
	ctx := SetRequestContext(httptest.NewRequest(http.MethodGet, nikPath, nil).Context(), httptest.NewRequest(http.MethodGet, nikPath, nil))
	info := GetRequestInfo(ctx)
	if info.RequestURL != "unknown" {
		t.Fatalf("RequestURL %q", info.RequestURL)
	}
	if strings.Contains(info.RequestURL, "320101") {
		t.Fatalf("NIK leaked: %q", info.RequestURL)
	}
}
