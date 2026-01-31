package service

import (
	"fmt"
	"net/http"
	"reflect"
	"slices"
	"strings"
)

type Mock struct {
	ID                    int64           `json:"id"`
	Method                string          `json:"method" validate:"notEmpty,httpMethod"`
	Path                  string          `json:"path" validate:"notEmpty"`
	RequestHeaderMatchers []HeaderMatcher `json:"requestHeaderMatchers"`
	RequestQueryMatchers  []QueryMatcher  `json:"requestQueryMatchers"`
	RequestBodyMatchers   []BodyMatcher   `json:"requestBodyMatchers"`
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
	return validationErrors
}
