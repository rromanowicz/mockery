package model_test

import (
	"strings"
	"testing"

	"github.com/rromanowicz/mockery/model"
)

func TestMock_Validate(t *testing.T) {
	tests := []struct {
		testName       string
		expectedResult bool
		expectedError  string
		input          model.Mock
	}{
		{"Valid Simple Mock", true, "", validSimple},
		{"Valid Full Mock", true, "", validFull},
		{"Invalid Path", false, model.InvalidPath, invalidPath},
		{"Invalid Regex", false, model.InvalidRegex, invalidRegex},
		{"Invalid BodyMatcher missing field", false, model.InvalidBodyMatcher, bodyMatcherMissingField},
		{"Invalid BodyMatcher invalid JsonPath", false, model.InvalidBodyMatcherJSONPath, bodyMatcherInvalidJSONPath},
		{"Invalid QueryMatcher missing field", false, model.InvalidQueryMatcher, queryMatcherMissingField},
		{"Invalid HeaderMatcher missing field", false, model.InvalidHeaderMatcher, headerMatcherMissingField},
		{"Missing method", false, model.CanNotBeEmpty, missingMethod},
		{"Invalid method", false, model.InvalidValue, invalidMethod},
		{"Invalid status", false, model.InvalidValue, invalidStatus},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			ok, errors := tt.input.Validate()
			if ok != tt.expectedResult {
				t.Errorf("Validate() = %v, want %v", ok, tt.expectedResult)
			}
			if !containsError(tt.expectedError, errors) {
				t.Errorf("Validate() = %v, want %v", errors, tt.expectedError)
			}
		})
	}
}

func containsError(errStr string, errors []string) bool {
	if len(errStr) == 0 {
		return true
	}
	for i := range errors {
		if strings.Contains(errors[i], errStr) {
			return true
		}
	}
	return false
}

var (
	validSimple = model.Mock{
		Method: "POST",
		Path:   "/test",
		Response: model.Response{
			Status: 200,
			Body:   make(model.JSONB),
		},
	}
	validFull = model.Mock{
		Method:                "POST",
		RegexPath:             "\\/test\\/\\d+",
		RequestBodyMatchers:   []model.Matcher{{"$.test", "test"}, {"$.foo", "bar"}},
		RequestQueryMatchers:  []model.Matcher{{"test", "test"}},
		RequestHeaderMatchers: []model.Matcher{{"test", "test"}},
		Response: model.Response{
			Status: 200,
			Body:   make(model.JSONB),
		},
	}
	invalidPath = model.Mock{
		Method:    "POST",
		Path:      "/test",
		RegexPath: "\\/test\\/\\d+",
		Response: model.Response{
			Status: 200,
			Body:   make(model.JSONB),
		},
	}
	invalidRegex = model.Mock{
		Method:    "POST",
		RegexPath: "\\/test\\/[asd",
		Response: model.Response{
			Status: 200,
			Body:   make(model.JSONB),
		},
	}
	bodyMatcherMissingField = model.Mock{
		Method:              "POST",
		Path:                "/test",
		RequestBodyMatchers: []model.Matcher{{"test", ""}},
		Response: model.Response{
			Status: 200,
			Body:   make(model.JSONB),
		},
	}
	bodyMatcherInvalidJSONPath = model.Mock{
		Method:              "POST",
		Path:                "/test",
		RequestBodyMatchers: []model.Matcher{{"test", "test"}},
		Response: model.Response{
			Status: 200,
			Body:   make(model.JSONB),
		},
	}
	queryMatcherMissingField = model.Mock{
		Method:               "POST",
		Path:                 "/test",
		RequestQueryMatchers: []model.Matcher{{"test", ""}},
		Response: model.Response{
			Status: 200,
			Body:   make(model.JSONB),
		},
	}
	headerMatcherMissingField = model.Mock{
		Method:                "POST",
		Path:                  "/test",
		RequestHeaderMatchers: []model.Matcher{{"test", ""}},
		Response: model.Response{
			Status: 200,
			Body:   make(model.JSONB),
		},
	}
	missingMethod = model.Mock{
		Path: "/test",
		Response: model.Response{
			Status: 200,
			Body:   make(model.JSONB),
		},
	}
	invalidMethod = model.Mock{
		Method: "TEST",
		Path:   "/test",
		Response: model.Response{
			Status: 200,
			Body:   make(model.JSONB),
		},
	}
	invalidStatus = model.Mock{
		Method: "POST",
		Path:   "/test",
		Response: model.Response{
			Status: 123,
			Body:   make(model.JSONB),
		},
	}
)
