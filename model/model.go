// Package model
package model

import (
	"fmt"
	"net/http"
	"reflect"
	"regexp"
	"slices"
	"strings"
)

type Mock struct {
	ID                    int64           `json:"id"`
	Method                string          `json:"method" validate:"notEmpty,httpMethod"`
	Path                  string          `json:"path,omitempty"`
	RegexPath             string          `json:"regexPath,omitempty"`
	RequestHeaderMatchers []HeaderMatcher `json:"requestHeaderMatchers,omitempty"`
	RequestQueryMatchers  []QueryMatcher  `json:"requestQueryMatchers,omitempty"`
	RequestBodyMatchers   []BodyMatcher   `json:"requestBodyMatchers,omitempty"`
	ResponseStatus        int             `json:"responseStatus" validate:"httpStatus"`
	ResponseBody          any             `json:"responseBody"`
}

type QueryMatcher struct {
	ParamName     string `json:"param"`
	ExpectedValue any    `json:"value"`
}

type BodyMatcher struct {
	JSONPathString string `json:"jsonPath"`
	ExpectedValue  any    `json:"value"`
}

type HeaderMatcher struct {
	HeaderName    string `json:"name"`
	ExpectedValue string `json:"value"`
}

type RegexMatcher struct {
	ID     int64
	Method string
	Regexp *regexp.Regexp
}

func (m Mock) Validate() []string {
	val := reflect.ValueOf(m)
	var validationErrors []string
	for i := 0; i < val.NumField(); i++ {
		tags := val.Type().Field(i).Tag.Get("validate")
		fieldValue := val.Field(i)
		for tag := range strings.SplitSeq(tags, ",") {
			switch tag {
			case "notEmpty":
				if fieldValue.Len() == 0 {
					validationErrors = append(validationErrors, fmt.Sprintf("'%v' can not be empty", val.Type().Field(i).Name))
				}
			case "httpMethod":
				methods := []string{http.MethodGet, http.MethodHead, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodConnect, http.MethodOptions, http.MethodTrace}
				if !slices.Contains(methods, fieldValue.String()) {
					validationErrors = append(validationErrors, fmt.Sprintf("%v Invalid value [%v]", val.Type().Field(i).Name, fieldValue.String()))
				}
			case "httpStatus":
				if len(http.StatusText(int(fieldValue.Int()))) == 0 {
					validationErrors = append(validationErrors, fmt.Sprintf("%v Invalid value: [%v]", val.Type().Field(i).Name, fieldValue.Int()))
				}
			}
		}
	}
	validateMissingData(m, &validationErrors)
	return validationErrors
}

func validateMissingData(mock Mock, validationErrors *[]string) {
	validatePath(mock, validationErrors)
	validateHeaderMatchers(mock, validationErrors)
	validateQueryMatchers(mock, validationErrors)
	validateBodyMatchers(mock, validationErrors)
}

func validateBodyMatchers(mock Mock, validationErrors *[]string) {
	for i := range mock.RequestBodyMatchers {
		matcher := mock.RequestBodyMatchers[i]
		if len(matcher.JSONPathString) == 0 && len(fmt.Sprint(matcher.ExpectedValue)) == 0 {
			*validationErrors = append(*validationErrors, "Invalid BodyMatcher. Both values must be provided.")
			break
		}
	}
}

func validateQueryMatchers(mock Mock, validationErrors *[]string) {
	for i := range mock.RequestQueryMatchers {
		matcher := mock.RequestQueryMatchers[i]
		if len(matcher.ParamName) == 0 && len(fmt.Sprint(matcher.ExpectedValue)) == 0 {
			*validationErrors = append(*validationErrors, "Invalid QueryMatcher. Both values must be provided.")
			break
		}
	}
}

func validateHeaderMatchers(mock Mock, validationErrors *[]string) {
	for i := range mock.RequestHeaderMatchers {
		matcher := mock.RequestHeaderMatchers[i]
		if len(matcher.HeaderName) == 0 && len(fmt.Sprint(matcher.ExpectedValue)) == 0 {
			*validationErrors = append(*validationErrors, "Invalid HeaderMatcher. Both values must be provided.")
			break
		}
	}
}

func validatePath(mock Mock, validationErrors *[]string) {
	if len(mock.Path) == 0 && len(mock.RegexPath) == 0 {
		*validationErrors = append(*validationErrors, "Invalid path. Either 'Path' or 'RegexPath' must be provided.")
	}
	if len(mock.RegexPath) != 0 {
		_, err := regexp.Compile(mock.RegexPath)
		if err != nil {
			*validationErrors = append(*validationErrors, fmt.Sprintf("Invalid RegexPath. %s is not a valid expression.", mock.RegexPath))
		}
	}
}
