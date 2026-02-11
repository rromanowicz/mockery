// Package model
package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"regexp"
	"slices"
	"strings"

	"github.com/theory/jsonpath"
)

type Mock struct {
	ID                    int64    `json:"id"`
	Method                string   `json:"method" validate:"notEmpty,httpMethod"`
	Path                  string   `json:"path,omitempty"`
	RegexPath             string   `json:"regexPath,omitempty"`
	RequestHeaderMatchers Matchers `json:"requestHeaderMatchers,omitempty" gorm:"type:jsonb"`
	RequestQueryMatchers  Matchers `json:"requestQueryMatchers,omitempty" gorm:"type:jsonb"`
	RequestBodyMatchers   Matchers `json:"requestBodyMatchers,omitempty" gorm:"type:jsonb"`
	ResponseStatus        int      `json:"responseStatus" validate:"httpStatus"`
	ResponseBody          JSONB    `json:"responseBody" gorm:"type:jsonb"`
}

type Matchers []Matcher

type Matcher struct {
	Key   string `json:"key"`
	Value any    `json:"value"`
}

type RegexMatcher struct {
	ID        int64
	Method    string
	RegexPath RegexPath `gorm:"name:regex_path"`
}

type RegexPath string

func (rp RegexPath) Compile() *regexp.Regexp {
	compiled, _ := regexp.Compile(string(rp))
	return compiled
}

func (m Mock) Validate() (bool, []string) {
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
	return len(validationErrors) == 0, validationErrors
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
		if len(matcher.Key) == 0 && len(fmt.Sprint(matcher.Value)) == 0 {
			*validationErrors = append(*validationErrors, "Invalid BodyMatcher. Both values must be provided.")
			break
		}
		_, err := jsonpath.Parse(matcher.Key)
		if err != nil {
			*validationErrors = append(*validationErrors, "Invalid BodyMatcher. Cannot parse key value as JsonPath.")
			break
		}
	}
}

func validateQueryMatchers(mock Mock, validationErrors *[]string) {
	for i := range mock.RequestQueryMatchers {
		matcher := mock.RequestQueryMatchers[i]
		if len(matcher.Key) == 0 && len(fmt.Sprint(matcher.Value)) == 0 {
			*validationErrors = append(*validationErrors, "Invalid QueryMatcher. Both values must be provided.")
			break
		}
	}
}

func validateHeaderMatchers(mock Mock, validationErrors *[]string) {
	for i := range mock.RequestHeaderMatchers {
		matcher := mock.RequestHeaderMatchers[i]
		if len(matcher.Key) == 0 && len(fmt.Sprint(matcher.Value)) == 0 {
			*validationErrors = append(*validationErrors, "Invalid HeaderMatcher. Both values must be provided.")
			break
		}
	}
}

func validatePath(mock Mock, validationErrors *[]string) {
	if (len(mock.Path) == 0 && len(mock.RegexPath) == 0) || (len(mock.Path) != 0 && len(mock.RegexPath) != 0) {
		*validationErrors = append(*validationErrors, "Invalid path. Either 'Path' or 'RegexPath' must be provided.")
	}
	if len(mock.RegexPath) != 0 {
		_, err := regexp.Compile(mock.RegexPath)
		if err != nil {
			*validationErrors = append(*validationErrors, fmt.Sprintf("Invalid RegexPath. %s is not a valid expression.", mock.RegexPath))
		}
	}
}

type JSONB map[string]any

func (a JSONB) Value() (driver.Value, error) {
	return json.Marshal(a)
}

func (a *JSONB) Scan(value any) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(b, &a)
}

func (a Matchers) Value() (driver.Value, error) {
	return json.Marshal(a)
}

func (a *Matchers) Scan(value any) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(b, &a)
}
