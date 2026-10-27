package xhttp

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thecodingmachine/gotenberg/internal/app/xhttp/pkg/resource"
	"github.com/thecodingmachine/gotenberg/internal/pkg/conf"
	"github.com/thecodingmachine/gotenberg/test"
)

func TestPingHandler(t *testing.T) {
	// should return 200.
	config := conf.DefaultConfig()
	srv := New(config)
	endpoint := pingEndpoint(config)
	req := httptest.NewRequest(http.MethodGet, endpoint, nil)
	test.AssertStatusCode(t, http.StatusOK, srv, req)
	// should return 405 as Method is wrong.
	req = httptest.NewRequest(http.MethodPost, endpoint, nil)
	test.AssertStatusCode(t, http.StatusMethodNotAllowed, srv, req)
}

func TestHTMLHandler(t *testing.T) {
	config := conf.DefaultConfig()
	srv := New(config)
	endpoint := htmlEndpoint(config)
	// should return 200.
	body, contentType := test.HTMLMultipartForm(t, nil)
	req := httptest.NewRequest(http.MethodPost, endpoint, body)
	req.Header.Set(echo.HeaderContentType, contentType)
	test.AssertStatusCode(t, http.StatusOK, srv, req)
	// should return 405 as Method is wrong.
	req = httptest.NewRequest(http.MethodGet, endpoint, nil)
	test.AssertStatusCode(t, http.StatusMethodNotAllowed, srv, req)
	// should return 415 as Content-Type is wrong.
	body, _ = test.HTMLMultipartForm(t, nil)
	req = httptest.NewRequest(http.MethodPost, endpoint, body)
	test.AssertStatusCode(t, http.StatusUnsupportedMediaType, srv, req)
	// should return 400 as "waitTimeout" form field
	// value is < 0.
	body, contentType = test.HTMLMultipartForm(t, map[string]string{string(resource.WaitTimeoutArgKey): "-1"})
	req = httptest.NewRequest(http.MethodPost, endpoint, body)
	req.Header.Set(echo.HeaderContentType, contentType)
	test.AssertStatusCode(t, http.StatusBadRequest, srv, req)
	// should return 400 as "waitTimeout" form field
	// value is is > config.MaximumWaitTimeout().
	body, contentType = test.HTMLMultipartForm(t, map[string]string{string(resource.WaitTimeoutArgKey): "31"})
	req = httptest.NewRequest(http.MethodPost, endpoint, body)
	req.Header.Set(echo.HeaderContentType, contentType)
	test.AssertStatusCode(t, http.StatusBadRequest, srv, req)
	// should return 400 as "waitTimeout" form field
	// value is invalid.
	body, contentType = test.HTMLMultipartForm(t, map[string]string{string(resource.WaitTimeoutArgKey): "not a float"})
	req = httptest.NewRequest(http.MethodPost, endpoint, body)
	req.Header.Set(echo.HeaderContentType, contentType)
	test.AssertStatusCode(t, http.StatusBadRequest, srv, req)
	// should return 504.
	body, contentType = test.HTMLMultipartForm(t, map[string]string{string(resource.WaitTimeoutArgKey): "0"})
	req = httptest.NewRequest(http.MethodPost, endpoint, body)
	req.Header.Set(echo.HeaderContentType, contentType)
	test.AssertStatusCode(t, http.StatusGatewayTimeout, srv, req)
	// should return 400 as "waitDelay" form field
	// value is < 0.
	body, contentType = test.HTMLMultipartForm(t, map[string]string{string(resource.WaitDelayArgKey): "-1"})
	req = httptest.NewRequest(http.MethodPost, endpoint, body)
	req.Header.Set(echo.HeaderContentType, contentType)
	test.AssertStatusCode(t, http.StatusBadRequest, srv, req)
	// should return 400 as "waitDelay" form field
	// value is is > config.MaximumWaitDelay().
	body, contentType = test.HTMLMultipartForm(t, map[string]string{string(resource.WaitDelayArgKey): "31"})
	req = httptest.NewRequest(http.MethodPost, endpoint, body)
	req.Header.Set(echo.HeaderContentType, contentType)
	test.AssertStatusCode(t, http.StatusBadRequest, srv, req)
	// should return 400 as "waitDelay" form field
	// value is invalid.
	body, contentType = test.HTMLMultipartForm(t, map[string]string{string(resource.WaitDelayArgKey): "not a float"})
	req = httptest.NewRequest(http.MethodPost, endpoint, body)
	req.Header.Set(echo.HeaderContentType, contentType)
	test.AssertStatusCode(t, http.StatusBadRequest, srv, req)
	// should return 400 as "paperWidth" form field
	// value is < 0.
	body, contentType = test.HTMLMultipartForm(t, map[string]string{string(resource.PaperWidthArgKey): "-1"})
	req = httptest.NewRequest(http.MethodPost, endpoint, body)
	req.Header.Set(echo.HeaderContentType, contentType)
	test.AssertStatusCode(t, http.StatusBadRequest, srv, req)
	// should return 400 as "paperWidth" form field
	// value is invalid.
	body, contentType = test.HTMLMultipartForm(t, map[string]string{string(resource.PaperWidthArgKey): "not a float"})
	req = httptest.NewRequest(http.MethodPost, endpoint, body)
	req.Header.Set(echo.HeaderContentType, contentType)
	test.AssertStatusCode(t, http.StatusBadRequest, srv, req)
	// should return 400 as "paperHeight" form field
	// value is < 0.
	body, contentType = test.HTMLMultipartForm(t, map[string]string{string(resource.PaperHeightArgKey): "-1"})
	req = httptest.NewRequest(http.MethodPost, endpoint, body)
	req.Header.Set(echo.HeaderContentType, contentType)
	test.AssertStatusCode(t, http.StatusBadRequest, srv, req)
	// should return 400 as "paperHeight" form field
	// value is invalid.
	body, contentType = test.HTMLMultipartForm(t, map[string]string{string(resource.PaperHeightArgKey): "not a float"})
	req = httptest.NewRequest(http.MethodPost, endpoint, body)
	req.Header.Set(echo.HeaderContentType, contentType)
	test.AssertStatusCode(t, http.StatusBadRequest, srv, req)
	// should return 400 as "marginTop" form field
	// value is < 0.
	body, contentType = test.HTMLMultipartForm(t, map[string]string{string(resource.MarginTopArgKey): "-1"})
	req = httptest.NewRequest(http.MethodPost, endpoint, body)
	req.Header.Set(echo.HeaderContentType, contentType)
	test.AssertStatusCode(t, http.StatusBadRequest, srv, req)
	// should return 400 as "marginTop" form field
	// value is invalid.
	body, contentType = test.HTMLMultipartForm(t, map[string]string{string(resource.MarginTopArgKey): "not a float"})
	req = httptest.NewRequest(http.MethodPost, endpoint, body)
	req.Header.Set(echo.HeaderContentType, contentType)
	test.AssertStatusCode(t, http.StatusBadRequest, srv, req)
	// should return 400 as "marginBottom" form field
	// value is < 0.
	body, contentType = test.HTMLMultipartForm(t, map[string]string{string(resource.MarginBottomArgKey): "-1"})
	req = httptest.NewRequest(http.MethodPost, endpoint, body)
	req.Header.Set(echo.HeaderContentType, contentType)
	test.AssertStatusCode(t, http.StatusBadRequest, srv, req)
	// should return 400 as "marginBottom" form field
	// value is invalid.
	body, contentType = test.HTMLMultipartForm(t, map[string]string{string(resource.MarginBottomArgKey): "not a float"})
	req = httptest.NewRequest(http.MethodPost, endpoint, body)
	req.Header.Set(echo.HeaderContentType, contentType)
	test.AssertStatusCode(t, http.StatusBadRequest, srv, req)
	// should return 400 as "marginLeft" form field
	// value is < 0.
	body, contentType = test.HTMLMultipartForm(t, map[string]string{string(resource.MarginLeftArgKey): "-1"})
	req = httptest.NewRequest(http.MethodPost, endpoint, body)
	req.Header.Set(echo.HeaderContentType, contentType)
	test.AssertStatusCode(t, http.StatusBadRequest, srv, req)
	// should return 400 as "marginLeft" form field
	// value is invalid.
	body, contentType = test.HTMLMultipartForm(t, map[string]string{string(resource.MarginLeftArgKey): "not a float"})
	req = httptest.NewRequest(http.MethodPost, endpoint, body)
	req.Header.Set(echo.HeaderContentType, contentType)
	test.AssertStatusCode(t, http.StatusBadRequest, srv, req)
	// should return 400 as "marginRight" form field
	// value is < 0.
	body, contentType = test.HTMLMultipartForm(t, map[string]string{string(resource.MarginRightArgKey): "-1"})
	req = httptest.NewRequest(http.MethodPost, endpoint, body)
	req.Header.Set(echo.HeaderContentType, contentType)
	test.AssertStatusCode(t, http.StatusBadRequest, srv, req)
	// should return 400 as "marginRight" form field
	// value is invalid.
	body, contentType = test.HTMLMultipartForm(t, map[string]string{string(resource.MarginRightArgKey): "not a float"})
	req = httptest.NewRequest(http.MethodPost, endpoint, body)
	req.Header.Set(echo.HeaderContentType, contentType)
	test.AssertStatusCode(t, http.StatusBadRequest, srv, req)
	// should return 400 as "landscape" form field
	// value is invalid.
	body, contentType = test.HTMLMultipartForm(t, map[string]string{string(resource.LandscapeArgKey): "not a boolean"})
	req = httptest.NewRequest(http.MethodPost, endpoint, body)
	req.Header.Set(echo.HeaderContentType, contentType)
	test.AssertStatusCode(t, http.StatusBadRequest, srv, req)
	// should return 400 as "googleChromeRpccBufferSize" form field
	// value is < 0.
	body, contentType = test.HTMLMultipartForm(t, map[string]string{string(resource.GoogleChromeRpccBufferSizeArgKey): "-1"})
	req = httptest.NewRequest(http.MethodPost, endpoint, body)
	req.Header.Set(echo.HeaderContentType, contentType)
	test.AssertStatusCode(t, http.StatusBadRequest, srv, req)
	// should return 400 as "googleChromeRpccBufferSize" form field
	// value is is > config.MaximumGoogleChromeRpccBufferSize().
	body, contentType = test.HTMLMultipartForm(t, map[string]string{string(resource.GoogleChromeRpccBufferSizeArgKey): "104857601"})
	req = httptest.NewRequest(http.MethodPost, endpoint, body)
	req.Header.Set(echo.HeaderContentType, contentType)
	test.AssertStatusCode(t, http.StatusBadRequest, srv, req)
	// should return 400 as "googleChromeRpccBufferSize" form field
	// value is invalid.
	body, contentType = test.HTMLMultipartForm(t, map[string]string{string(resource.GoogleChromeRpccBufferSizeArgKey): "not an int"})
	req = httptest.NewRequest(http.MethodPost, endpoint, body)
	req.Header.Set(echo.HeaderContentType, contentType)
	test.AssertStatusCode(t, http.StatusBadRequest, srv, req)
}

func TestURLHandler(t *testing.T) {
	config := conf.DefaultConfig()
	srv := New(config)
	endpoint := urlEndpoint(config)
	// should return 200.
	body, contentType := test.URLMultipartForm(t, nil)
	req := httptest.NewRequest(http.MethodPost, endpoint, body)
	req.Header.Set(echo.HeaderContentType, contentType)
	test.AssertStatusCode(t, http.StatusOK, srv, req)
	// should return 405 as Method is wrong.
	req = httptest.NewRequest(http.MethodGet, endpoint, nil)
	test.AssertStatusCode(t, http.StatusMethodNotAllowed, srv, req)
	// should return 415 as Content-Type is wrong.
	body, _ = test.URLMultipartForm(t, nil)
	req = httptest.NewRequest(http.MethodPost, endpoint, body)
	test.AssertStatusCode(t, http.StatusUnsupportedMediaType, srv, req)
	// should return 400 as "waitTimeout" form field
	// value is < 0.
	body, contentType = test.URLMultipartForm(t, map[string]string{string(resource.WaitTimeoutArgKey): "-1"})
	req = httptest.NewRequest(http.MethodPost, endpoint, body)
	req.Header.Set(echo.HeaderContentType, contentType)
	test.AssertStatusCode(t, http.StatusBadRequest, srv, req)
	// should return 400 as "waitTimeout" form field
	// value is is > config.MaximumWaitTimeout().
	body, contentType = test.URLMultipartForm(t, map[string]string{string(resource.WaitTimeoutArgKey): "31"})
	req = httptest.NewRequest(http.MethodPost, endpoint, body)
	req.Header.Set(echo.HeaderContentType, contentType)
	test.AssertStatusCode(t, http.StatusBadRequest, srv, req)
	// should return 400 as "waitTimeout" form field
	// value is invalid.
	body, contentType = test.URLMultipartForm(t, map[string]string{string(resource.WaitTimeoutArgKey): "not a float"})
	req = httptest.NewRequest(http.MethodPost, endpoint, body)
	req.Header.Set(echo.HeaderContentType, contentType)
	test.AssertStatusCode(t, http.StatusBadRequest, srv, req)
	// should return 504.
	body, contentType = test.URLMultipartForm(t, map[string]string{string(resource.WaitTimeoutArgKey): "0"})
	req = httptest.NewRequest(http.MethodPost, endpoint, body)
	req.Header.Set(echo.HeaderContentType, contentType)
	test.AssertStatusCode(t, http.StatusGatewayTimeout, srv, req)
	// should return 400 as "waitDelay" form field
	// value is < 0.
	body, contentType = test.URLMultipartForm(t, map[string]string{string(resource.WaitDelayArgKey): "-1"})
	req = httptest.NewRequest(http.MethodPost, endpoint, body)
	req.Header.Set(echo.HeaderContentType, contentType)
	test.AssertStatusCode(t, http.StatusBadRequest, srv, req)
	// should return 400 as "waitDelay" form field
	// value is is > config.MaximumWaitDelay().
	body, contentType = test.URLMultipartForm(t, map[string]string{string(resource.WaitDelayArgKey): "31"})
	req = httptest.NewRequest(http.MethodPost, endpoint, body)
	req.Header.Set(echo.HeaderContentType, contentType)
	test.AssertStatusCode(t, http.StatusBadRequest, srv, req)
	// should return 400 as "waitDelay" form field
	// value is invalid.
	body, contentType = test.URLMultipartForm(t, map[string]string{string(resource.WaitDelayArgKey): "not a float"})
	req = httptest.NewRequest(http.MethodPost, endpoint, body)
	req.Header.Set(echo.HeaderContentType, contentType)
	test.AssertStatusCode(t, http.StatusBadRequest, srv, req)
	// should return 400 as "paperWidth" form field
	// value is < 0.
	body, contentType = test.URLMultipartForm(t, map[string]string{string(resource.PaperWidthArgKey): "-1"})
	req = httptest.NewRequest(http.MethodPost, endpoint, body)
	req.Header.Set(echo.HeaderContentType, contentType)
	test.AssertStatusCode(t, http.StatusBadRequest, srv, req)
	// should return 400 as "paperWidth" form field
	// value is invalid.
	body, contentType = test.URLMultipartForm(t, map[string]string{string(resource.PaperWidthArgKey): "not a float"})
	req = httptest.NewRequest(http.MethodPost, endpoint, body)
	req.Header.Set(echo.HeaderContentType, contentType)
	test.AssertStatusCode(t, http.StatusBadRequest, srv, req)
	// should return 400 as "paperHeight" form field
	// value is < 0.
	body, contentType = test.URLMultipartForm(t, map[string]string{string(resource.PaperHeightArgKey): "-1"})
	req = httptest.NewRequest(http.MethodPost, endpoint, body)
	req.Header.Set(echo.HeaderContentType, contentType)
	test.AssertStatusCode(t, http.StatusBadRequest, srv, req)
	// should return 400 as "paperHeight" form field
	// value is invalid.
	body, contentType = test.URLMultipartForm(t, map[string]string{string(resource.PaperHeightArgKey): "not a float"})
	req = httptest.NewRequest(http.MethodPost, endpoint, body)
	req.Header.Set(echo.HeaderContentType, contentType)
	test.AssertStatusCode(t, http.StatusBadRequest, srv, req)
	// should return 400 as "marginTop" form field
	// value is < 0.
	body, contentType = test.URLMultipartForm(t, map[string]string{string(resource.MarginTopArgKey): "-1"})
	req = httptest.NewRequest(http.MethodPost, endpoint, body)
	req.Header.Set(echo.HeaderContentType, contentType)
	test.AssertStatusCode(t, http.StatusBadRequest, srv, req)
	// should return 400 as "marginTop" form field
	// value is invalid.
	body, contentType = test.URLMultipartForm(t, map[string]string{string(resource.MarginTopArgKey): "not a float"})
	req = httptest.NewRequest(http.MethodPost, endpoint, body)
	req.Header.Set(echo.HeaderContentType, contentType)
	test.AssertStatusCode(t, http.StatusBadRequest, srv, req)
	// should return 400 as "marginBottom" form field
	// value is < 0.
	body, contentType = test.URLMultipartForm(t, map[string]string{string(resource.MarginBottomArgKey): "-1"})
	req = httptest.NewRequest(http.MethodPost, endpoint, body)
	req.Header.Set(echo.HeaderContentType, contentType)
	test.AssertStatusCode(t, http.StatusBadRequest, srv, req)
	// should return 400 as "marginBottom" form field
	// value is invalid.
	body, contentType = test.URLMultipartForm(t, map[string]string{string(resource.MarginBottomArgKey): "not a float"})
	req = httptest.NewRequest(http.MethodPost, endpoint, body)
	req.Header.Set(echo.HeaderContentType, contentType)
	test.AssertStatusCode(t, http.StatusBadRequest, srv, req)
	// should return 400 as "marginLeft" form field
	// value is < 0.
	body, contentType = test.URLMultipartForm(t, map[string]string{string(resource.MarginLeftArgKey): "-1"})
	req = httptest.NewRequest(http.MethodPost, endpoint, body)
	req.Header.Set(echo.HeaderContentType, contentType)
	test.AssertStatusCode(t, http.StatusBadRequest, srv, req)
	// should return 400 as "marginLeft" form field
	// value is invalid.
	body, contentType = test.URLMultipartForm(t, map[string]string{string(resource.MarginLeftArgKey): "not a float"})
	req = httptest.NewRequest(http.MethodPost, endpoint, body)
	req.Header.Set(echo.HeaderContentType, contentType)
	test.AssertStatusCode(t, http.StatusBadRequest, srv, req)
	// should return 400 as "marginRight" form field
	// value is < 0.
	body, contentType = test.URLMultipartForm(t, map[string]string{string(resource.MarginRightArgKey): "-1"})
	req = httptest.NewRequest(http.MethodPost, endpoint, body)
	req.Header.Set(echo.HeaderContentType, contentType)
	test.AssertStatusCode(t, http.StatusBadRequest, srv, req)
	// should return 400 as "marginRight" form field
	// value is invalid.
	body, contentType = test.URLMultipartForm(t, map[string]string{string(resource.MarginRightArgKey): "not a float"})
	req = httptest.NewRequest(http.MethodPost, endpoint, body)
	req.Header.Set(echo.HeaderContentType, contentType)
	test.AssertStatusCode(t, http.StatusBadRequest, srv, req)
	// should return 400 as "landscape" form field
	// value is invalid.
	body, contentType = test.URLMultipartForm(t, map[string]string{string(resource.LandscapeArgKey): "not a boolean"})
	req = httptest.NewRequest(http.MethodPost, endpoint, body)
	req.Header.Set(echo.HeaderContentType, contentType)
	test.AssertStatusCode(t, http.StatusBadRequest, srv, req)
	// should return 400 as "googleChromeRpccBufferSize" form field
	// value is < 0.
	body, contentType = test.URLMultipartForm(t, map[string]string{string(resource.GoogleChromeRpccBufferSizeArgKey): "-1"})
	req = httptest.NewRequest(http.MethodPost, endpoint, body)
	req.Header.Set(echo.HeaderContentType, contentType)
	test.AssertStatusCode(t, http.StatusBadRequest, srv, req)
	// should return 400 as "googleChromeRpccBufferSize" form field
	// value is is > config.MaximumGoogleChromeRpccBufferSize().
	body, contentType = test.URLMultipartForm(t, map[string]string{string(resource.GoogleChromeRpccBufferSizeArgKey): "104857601"})
	req = httptest.NewRequest(http.MethodPost, endpoint, body)
	req.Header.Set(echo.HeaderContentType, contentType)
	test.AssertStatusCode(t, http.StatusBadRequest, srv, req)
	// should return 400 as "googleChromeRpccBufferSize" form field
	// value is invalid.
	body, contentType = test.URLMultipartForm(t, map[string]string{string(resource.GoogleChromeRpccBufferSizeArgKey): "not an int"})
	req = httptest.NewRequest(http.MethodPost, endpoint, body)
	req.Header.Set(echo.HeaderContentType, contentType)
	test.AssertStatusCode(t, http.StatusBadRequest, srv, req)
}

func TestMarkdownHandler(t *testing.T) {
	config := conf.DefaultConfig()
	srv := New(config)
	endpoint := markdownEndpoint(config)
	// should return 200.
	body, contentType := test.MarkdownMultipartForm(t, nil)
	req := httptest.NewRequest(http.MethodPost, endpoint, body)
	req.Header.Set(echo.HeaderContentType, contentType)
	test.AssertStatusCode(t, http.StatusOK, srv, req)
	// should return 405 as Method is wrong.
	req = httptest.NewRequest(http.MethodGet, endpoint, nil)
	test.AssertStatusCode(t, http.StatusMethodNotAllowed, srv, req)
	// should return 415 as Content-Type is wrong.
	body, _ = test.MarkdownMultipartForm(t, nil)
	req = httptest.NewRequest(http.MethodPost, endpoint, body)
	test.AssertStatusCode(t, http.StatusUnsupportedMediaType, srv, req)
	// should return 400 as "waitTimeout" form field
	// value is < 0.
	body, contentType = test.MarkdownMultipartForm(t, map[string]string{string(resource.WaitTimeoutArgKey): "-1"})
	req = httptest.NewRequest(http.MethodPost, endpoint, body)
	req.Header.Set(echo.HeaderContentType, contentType)
	test.AssertStatusCode(t, http.StatusBadRequest, srv, req)
	// should return 400 as "waitTimeout" form field
	// value is is > config.MaximumWaitTimeout().
	body, contentType = test.MarkdownMultipartForm(t, map[string]string{string(resource.WaitTimeoutArgKey): "31"})
	req = httptest.NewRequest(http.MethodPost, endpoint, body)
	req.Header.Set(echo.HeaderContentType, contentType)
	test.AssertStatusCode(t, http.StatusBadRequest, srv, req)
	// should return 400 as "waitTimeout" form field
	// value is invalid.
	body, contentType = test.MarkdownMultipartForm(t, map[string]string{string(resource.WaitTimeoutArgKey): "not a float"})
	req = httptest.NewRequest(http.MethodPost, endpoint, body)
	req.Header.Set(echo.HeaderContentType, contentType)
	test.AssertStatusCode(t, http.StatusBadRequest, srv, req)
	// should return 504.
	body, contentType = test.MarkdownMultipartForm(t, map[string]string{string(resource.WaitTimeoutArgKey): "0"})
	req = httptest.NewRequest(http.MethodPost, endpoint, body)
	req.Header.Set(echo.HeaderContentType, contentType)
	test.AssertStatusCode(t, http.StatusGatewayTimeout, srv, req)
	// should return 400 as "waitDelay" form field
	// value is < 0.
	body, contentType = test.MarkdownMultipartForm(t, map[string]string{string(resource.WaitDelayArgKey): "-1"})
	req = httptest.NewRequest(http.MethodPost, endpoint, body)
	req.Header.Set(echo.HeaderContentType, contentType)
	test.AssertStatusCode(t, http.StatusBadRequest, srv, req)
	// should return 400 as "waitDelay" form field
	// value is is > config.MaximumWaitDelay().
	body, contentType = test.MarkdownMultipartForm(t, map[string]string{string(resource.WaitDelayArgKey): "31"})
	req = httptest.NewRequest(http.MethodPost, endpoint, body)
	req.Header.Set(echo.HeaderContentType, contentType)
	test.AssertStatusCode(t, http.StatusBadRequest, srv, req)
	// should return 400 as "waitDelay" form field
	// value is invalid.
	body, contentType = test.MarkdownMultipartForm(t, map[string]string{string(resource.WaitDelayArgKey): "not a float"})
	req = httptest.NewRequest(http.MethodPost, endpoint, body)
	req.Header.Set(echo.HeaderContentType, contentType)
	test.AssertStatusCode(t, http.StatusBadRequest, srv, req)
	// should return 400 as "paperWidth" form field
	// value is < 0.
	body, contentType = test.MarkdownMultipartForm(t, map[string]string{string(resource.PaperWidthArgKey): "-1"})
	req = httptest.NewRequest(http.MethodPost, endpoint, body)
	req.Header.Set(echo.HeaderContentType, contentType)
	test.AssertStatusCode(t, http.StatusBadRequest, srv, req)
	// should return 400 as "paperWidth" form field
	// value is invalid.
	body, contentType = test.MarkdownMultipartForm(t, map[string]string{string(resource.PaperWidthArgKey): "not a float"})
	req = httptest.NewRequest(http.MethodPost, endpoint, body)
	req.Header.Set(echo.HeaderContentType, contentType)
	test.AssertStatusCode(t, http.StatusBadRequest, srv, req)
	// should return 400 as "paperHeight" form field
	// value is < 0.
	body, contentType = test.MarkdownMultipartForm(t, map[string]string{string(resource.PaperHeightArgKey): "-1"})
	req = httptest.NewRequest(http.MethodPost, endpoint, body)
	req.Header.Set(echo.HeaderContentType, contentType)
	test.AssertStatusCode(t, http.StatusBadRequest, srv, req)
	// should return 400 as "paperHeight" form field
	// value is invalid.
	body, contentType = test.MarkdownMultipartForm(t, map[string]string{string(resource.PaperHeightArgKey): "not a float"})
	req = httptest.NewRequest(http.MethodPost, endpoint, body)
	req.Header.Set(echo.HeaderContentType, contentType)
	test.AssertStatusCode(t, http.StatusBadRequest, srv, req)
	// should return 400 as "marginTop" form field
	// value is < 0.
	body, contentType = test.MarkdownMultipartForm(t, map[string]string{string(resource.MarginTopArgKey): "-1"})
	req = httptest.NewRequest(http.MethodPost, endpoint, body)
	req.Header.Set(echo.HeaderContentType, contentType)
	test.AssertStatusCode(t, http.StatusBadRequest, srv, req)
	// should return 400 as "marginTop" form field
	// value is invalid.
	body, contentType = test.MarkdownMultipartForm(t, map[string]string{string(resource.MarginTopArgKey): "not a float"})
	req = httptest.NewRequest(http.MethodPost, endpoint, body)
	req.Header.Set(echo.HeaderContentType, contentType)
	test.AssertStatusCode(t, http.StatusBadRequest, srv, req)
	// should return 400 as "marginBottom" form field
	// value is < 0.
	body, contentType = test.MarkdownMultipartForm(t, map[string]string{string(resource.MarginBottomArgKey): "-1"})
	req = httptest.NewRequest(http.MethodPost, endpoint, body)
	req.Header.Set(echo.HeaderContentType, contentType)
	test.AssertStatusCode(t, http.StatusBadRequest, srv, req)
	// should return 400 as "marginBottom" form field
	// value is invalid.
	body, contentType = test.MarkdownMultipartForm(t, map[string]string{string(resource.MarginBottomArgKey): "not a float"})
	req = httptest.NewRequest(http.MethodPost, endpoint, body)
	req.Header.Set(echo.HeaderContentType, contentType)
	test.AssertStatusCode(t, http.StatusBadRequest, srv, req)
	// should return 400 as "marginLeft" form field
	// value is < 0.
	body, contentType = test.MarkdownMultipartForm(t, map[string]string{string(resource.MarginLeftArgKey): "-1"})
	req = httptest.NewRequest(http.MethodPost, endpoint, body)
	req.Header.Set(echo.HeaderContentType, contentType)
	test.AssertStatusCode(t, http.StatusBadRequest, srv, req)
	// should return 400 as "marginLeft" form field
	// value is invalid.
	body, contentType = test.MarkdownMultipartForm(t, map[string]string{string(resource.MarginLeftArgKey): "not a float"})
	req = httptest.NewRequest(http.MethodPost, endpoint, body)
	req.Header.Set(echo.HeaderContentType, contentType)
	test.AssertStatusCode(t, http.StatusBadRequest, srv, req)
	// should return 400 as "marginRight" form field
	// value is < 0.
	body, contentType = test.MarkdownMultipartForm(t, map[string]string{string(resource.MarginRightArgKey): "-1"})
	req = httptest.NewRequest(http.MethodPost, endpoint, body)
	req.Header.Set(echo.HeaderContentType, contentType)
	test.AssertStatusCode(t, http.StatusBadRequest, srv, req)
	// should return 400 as "marginRight" form field
	// value is invalid.
	body, contentType = test.MarkdownMultipartForm(t, map[string]string{string(resource.MarginRightArgKey): "not a float"})
	req = httptest.NewRequest(http.MethodPost, endpoint, body)
	req.Header.Set(echo.HeaderContentType, contentType)
	test.AssertStatusCode(t, http.StatusBadRequest, srv, req)
	// should return 400 as "landscape" form field
	// value is invalid.
	body, contentType = test.MarkdownMultipartForm(t, map[string]string{string(resource.LandscapeArgKey): "not a boolean"})
	req = httptest.NewRequest(http.MethodPost, endpoint, body)
	req.Header.Set(echo.HeaderContentType, contentType)
	test.AssertStatusCode(t, http.StatusBadRequest, srv, req)
	// should return 400 as "googleChromeRpccBufferSize" form field
	// value is < 0.
	body, contentType = test.MarkdownMultipartForm(t, map[string]string{string(resource.GoogleChromeRpccBufferSizeArgKey): "-1"})
	req = httptest.NewRequest(http.MethodPost, endpoint, body)
	req.Header.Set(echo.HeaderContentType, contentType)
	test.AssertStatusCode(t, http.StatusBadRequest, srv, req)
	// should return 400 as "googleChromeRpccBufferSize" form field
	// value is is > config.MaximumGoogleChromeRpccBufferSize().
	body, contentType = test.MarkdownMultipartForm(t, map[string]string{string(resource.GoogleChromeRpccBufferSizeArgKey): "104857601"})
	req = httptest.NewRequest(http.MethodPost, endpoint, body)
	req.Header.Set(echo.HeaderContentType, contentType)
	test.AssertStatusCode(t, http.StatusBadRequest, srv, req)
	// should return 400 as "googleChromeRpccBufferSize" form field
	// value is invalid.
	body, contentType = test.MarkdownMultipartForm(t, map[string]string{string(resource.GoogleChromeRpccBufferSizeArgKey): "not an int"})
	req = httptest.NewRequest(http.MethodPost, endpoint, body)
	req.Header.Set(echo.HeaderContentType, contentType)
	test.AssertStatusCode(t, http.StatusBadRequest, srv, req)
}

func TestWebhook(t *testing.T) {
	customHeaderRealKey := http.CanonicalHeaderKey("MyCustomHeader")
	customHeaderKey := fmt.Sprintf("%s%s", resource.WebhookURLCustomHTTPHeaderCanonicalBaseKey, customHeaderRealKey)
	customHeaderValue := "foo"
	status := make(chan error, 2)
	rcv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			status <- fmt.Errorf("wrong method: got '%s' want '%s'", r.Method, http.MethodPost)
			return
		}
		if r.Header.Get(echo.HeaderContentType) != "application/pdf" {
			status <- fmt.Errorf("wrong Content-type: got '%s' want '%s'", r.Header.Get(echo.HeaderContentType), "application/pdf")
			return
		}
		if r.Header.Get(customHeaderRealKey) != customHeaderValue {
			status <- fmt.Errorf("wrong '%s': got '%s' want '%s'", customHeaderRealKey, r.Header.Get(customHeaderRealKey), customHeaderValue)
			return
		}
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			status <- err
			return
		}
		if !bytes.HasPrefix(body, []byte("%PDF-")) {
			status <- errors.New("callback body is not a PDF")
			return
		}
		status <- nil
	}))
	defer rcv.Close()
	config := conf.DefaultConfig()
	srv := New(config)
	// our custom server should receive the PDF.
	body, contentType := test.HTMLMultipartForm(t, map[string]string{string(resource.WebhookURLArgKey): rcv.URL})
	req := httptest.NewRequest(http.MethodPost, htmlEndpoint(config), body)
	req.Header.Set(echo.HeaderContentType, contentType)
	req.Header.Set(customHeaderKey, customHeaderValue)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
	select {
	case err := <-status:
		assert.NoError(t, err)
	case <-time.After(10 * time.Second):
		t.Fatal("timed out waiting for webhook callback")
	}
}

func TestResultFilename(t *testing.T) {
	config := conf.DefaultConfig()
	srv := New(config)
	body, contentType := test.HTMLMultipartForm(t, map[string]string{string(resource.ResultFilenameArgKey): "foo.pdf"})
	req := httptest.NewRequest(http.MethodPost, htmlEndpoint(config), body)
	req.Header.Set(echo.HeaderContentType, contentType)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "attachment; filename=\"foo.pdf\"", rec.Header().Get(echo.HeaderContentDisposition))
	assert.True(t, bytes.HasPrefix(rec.Body.Bytes(), []byte("%PDF-")))
}
