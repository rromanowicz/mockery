package server

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/rromanowicz/mockery/model"
)

func runTestServer() *httptest.Server {
	config := model.Config{DBType: "InMemory", AutoImport: false}
	_, handler := SetupServer(&config)

	return httptest.NewServer(handler)
}

func Test_Api_Integration(t *testing.T) {
	ts := runTestServer()
	defer ts.Close()

	t.Run("GET /health", func(t *testing.T) {
		resp, err := http.Get(fmt.Sprintf("%s/health", ts.URL))
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		assert.Equal(t, 200, resp.StatusCode)
	})

	setupTests := []struct {
		testName       string
		expectedStatus int
		expectedResult string
		expectedError  string
		input          string
	}{
		{"POST /config missing path", 400, "", "", postConfigMissingPath},
		{"POST /config simple path", 201, "", "", postConfigRegexPath},
		{"POST /config regex path", 201, "", "", postConfigSimplePath},
		{"POST /config query matcher", 201, "", "", postConfigQueryMatcher},
		{"POST /config header matcher", 201, "", "", postConfigHeaderMatcher},
		{"POST /config body matcher", 201, "", "", postConfigBodyMatcher},
	}

	for _, tt := range setupTests {
		t.Run(tt.testName, func(t *testing.T) {
			io := bytes.NewBuffer([]byte(tt.input))
			resp, _ := http.Post(fmt.Sprintf("%s/config", ts.URL), "application/json", io)
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
		})
	}

	pathTests := []struct {
		testName       string
		expectedStatus int
		expectedResult string
		requestMethod  string
		requestPath    string
		requestBody    string
		query          []string
		header         []string
	}{
		{"GET /config/list", 200, "", "GET", "/config/list", "", []string{}, []string{}},
		{"GET /foo/bar/1/baz", 200, `{"bar":{"id":1},"foo":true}`, "GET", "/foo/bar/1/baz", "", []string{}, []string{}},
		{"GET /foo", 200, `{"bar":{"id":2},"foo":true}`, "GET", "/foo", "", []string{}, []string{}},
		{"GET /bar", 200, `{"bar":{"id":3},"foo":true}`, "GET", "/bar", "", []string{"id", "3"}, []string{}},
		{"GET /bar", 200, `{"bar":{"id":4},"foo":true}`, "GET", "/bar", "", []string{}, []string{"foo", "bar"}},
		{"POST /bar", 201, `{"bar":{"id":5},"foo":true}`, "POST", "/bar", `{"foo": "bar"}`, []string{}, []string{}},
	}

	for _, tt := range pathTests {
		t.Run(tt.testName, func(t *testing.T) {
			var reqBody io.Reader = nil
			if len(tt.requestBody) != 0 {
				reqBody = bytes.NewBuffer([]byte(tt.requestBody))
			}
			req, _ := http.NewRequest(tt.requestMethod, fmt.Sprintf("%s%s", ts.URL, tt.requestPath), reqBody)
			if len(tt.query) != 0 {
				q := req.URL.Query()
				q.Add(tt.query[0], tt.query[1])
				req.URL.RawQuery = q.Encode()
			}
			if len(tt.header) != 0 {
				req.Header.Add(tt.header[0], tt.header[1])
			}
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				panic(err)
			}

			buf := new(bytes.Buffer)
			defer resp.Body.Close()
			_, _ = buf.ReadFrom(resp.Body)

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
			if len(tt.expectedResult) != 0 {
				assert.Equal(t, tt.expectedResult, buf.String())
			}
		})
	}
}

var (
	postConfigMissingPath = `{
		"method": "GET",
		"responseStatus": 200,
		"responseBody": { "foo": true, "bar": { "id": null } }
	}`
	postConfigRegexPath = `{
		"method": "GET",
		"regexPath": "\/foo\/bar\/\\d+\/baz",
		"responseStatus": 200,
		"responseBody": { "foo": true, "bar": { "id": 1 } }
	}`
	postConfigSimplePath = `{
		"method": "GET",
		"path": "/foo",
		"responseStatus": 200,
		"responseBody": { "foo": true, "bar": { "id": 2 } }
	}`
	postConfigQueryMatcher = `{
		"method": "GET",
		"path": "/bar",
		"requestQueryMatchers": [ { "key": "id", "value": 3 } ],
		"responseStatus": 200,
		"responseBody": { "foo": true, "bar": { "id": 3 } }
	}`
	postConfigHeaderMatcher = `{
		"method": "GET",
		"path": "/bar",
		"requestHeaderMatchers": [ { "key": "foo", "value": "bar" } ],
		"responseStatus": 200,
		"responseBody": { "foo": true, "bar": { "id": 4 } }
	}`
	postConfigBodyMatcher = `{
		"method": "POST",
		"path": "/bar",
		"requestBodyMatchers": [ { "key": "$.foo", "value": "bar" } ],
		"responseStatus": 201,
		"responseBody": { "foo": true, "bar": { "id": 5 } }
	}`
)
