package test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// PageURL serves a local HTML page for URL conversion tests.
func PageURL(t *testing.T) string {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if _, err := fmt.Fprint(w, "<!doctype html><html><body>Test page</body></html>"); err != nil {
			t.Errorf("write test page: %v", err)
		}
	}))
	t.Cleanup(srv.Close)
	return srv.URL
}

// AssertStatusCode checks if the given request
// returns the expected status code.
func AssertStatusCode(t *testing.T, expectedStatusCode int, srv http.Handler, req *http.Request) {
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	assert.Equal(t, expectedStatusCode, rec.Code)
}
