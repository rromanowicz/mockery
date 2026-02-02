// Package routing
package routing

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/rromanowicz/mockery/context"
	"github.com/rromanowicz/mockery/model"

	"github.com/theory/jsonpath"
)

func RegisterRoutes(ctx context.Context, handler *RegexpHandler) {
	regHelp, _ := regexp.Compile("/help")
	regConfigList, _ := regexp.Compile("/config/list")
	regConfig, _ := regexp.Compile("/config.*")
	reg, _ := regexp.Compile("/.*")
	handler.HandleFunc(regHelp, handleHelp)
	handler.HandleFunc(regConfigList, handleConfigList(ctx))
	handler.HandleFunc(regConfig, handleConfig(ctx))
	handler.HandleFunc(reg, handleAll(ctx))
}

func handleHelp(rw http.ResponseWriter, req *http.Request) {
	data, err := os.ReadFile("./README.md")
	if err != nil {
		log.Println(err.Error())
	}
	rw.Header().Set("Content-Type", "text/html; charset=utf-8")
	rw.Write(data)
}

func handleConfig(ctx context.Context) func(rw http.ResponseWriter, req *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		switch req.Method {
		case "GET":
			mocks, err := ctx.MockService.Get(req.Method, req.URL.Path)
			if err != nil {
				rw.WriteHeader(http.StatusNotFound)
				rw.Write([]byte(err.Error()))
			} else {
				mock, err := filterMocks(mocks, req)
				if err != nil {
					rw.WriteHeader(http.StatusNotFound)
					rw.Write([]byte(err.Error()))
				} else {
					rw.WriteHeader(mock.ResponseStatus)
					rw.Write([]byte(mock.ResponseBody.(string)))
				}
			}
		case "POST":
			var requestMock model.Mock
			var err error
			defer req.Body.Close()
			err = json.NewDecoder(req.Body).Decode(&requestMock)
			if err != nil {
				rw.WriteHeader(http.StatusBadRequest)
				rw.Write([]byte(err.Error()))
				return
			}
			errors := requestMock.Validate()
			if len(errors) != 0 {
				rw.WriteHeader(http.StatusBadRequest)
				errorsJSON, _ := json.Marshal(errors)
				rw.Write(errorsJSON)
				return
			}
			mock, err := ctx.MockService.Add(requestMock)
			if err != nil {
				rw.WriteHeader(http.StatusBadRequest)
				rw.Write([]byte(err.Error()))
			} else {
				rw.Header().Set("Content-Type", "application/json")
				rw.WriteHeader(http.StatusCreated)
				jsonBody, _ := json.Marshal(mock)
				rw.Write([]byte(jsonBody))
				log.Printf("Created new mock for path [%s]", mock.Path)
			}
		case "DELETE":
			var id int64
			id, _ = strconv.ParseInt(req.URL.Query().Get("id"), 10, 64)
			err := ctx.MockService.Delete(id)
			if err != nil {
				rw.WriteHeader(http.StatusNotFound)
				rw.Write([]byte(err.Error()))
			} else {
				rw.WriteHeader(http.StatusOK)
			}
		}
	}
}

func handleConfigList(ctx context.Context) func(rw http.ResponseWriter, req *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		mocks, err := ctx.MockService.List()
		if err != nil {
			rw.WriteHeader(http.StatusNotFound)
			rw.Write([]byte(err.Error()))
		} else {
			rw.Header().Set("Content-Type", "application/json")
			rw.WriteHeader(http.StatusOK)
			response, _ := json.Marshal(mocks)
			rw.Write(response)
		}
	}
}

func handleAll(ctx context.Context) func(rw http.ResponseWriter, req *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		mocks, err := ctx.MockService.Get(req.Method, req.URL.Path)
		if err != nil {
			rw.WriteHeader(http.StatusNotFound)
			rw.Write([]byte(err.Error()))
		} else {
			mock, err := filterMocks(mocks, req)
			if err != nil {
				rw.WriteHeader(http.StatusNotFound)
				rw.Write([]byte(err.Error()))
			} else {
				log.Printf("Matched Mock[id=%v]", mock.ID)
				rw.Header().Set("Content-Type", "application/json")
				rw.WriteHeader(mock.ResponseStatus)
				response := strings.ReplaceAll(mock.ResponseBody.(string), "\\", "")
				rw.Write([]byte(response))
			}
		}
	}
}

func filterMocks(mocks []model.Mock, req *http.Request) (model.Mock, error) {
	if len(mocks) == 0 {
		return model.Mock{}, errors.New("not found")
	}

	requestBody, err := io.ReadAll(req.Body)
	if err != nil {
		log.Printf("Failed to read request body. %s", err.Error())
	}

	var matchedMocks []*model.Mock

	for i := range mocks {
		mock := &mocks[i]
		if isMatchingRequestBody(mock.RequestBodyMatchers, requestBody) &&
			isMatchingRequestQuery(mock.RequestQueryMatchers, req.URL.Query()) &&
			isMatchingRequestHeader(mock.RequestHeaderMatchers, &req.Header) {
			matchedMocks = append(matchedMocks, mock)
		}
	}

	if len(matchedMocks) == 0 {
		return model.Mock{}, errors.New("not matched")
	}

	return *matchedMocks[0], nil
}

func isMatchingRequestQuery(queryMatchers []model.QueryMatcher, requestQueryParams url.Values) bool {
	if len(queryMatchers) == 0 {
		return true
	}
	if len(requestQueryParams) == 0 {
		return false
	}

	for _, matcher := range queryMatchers {
		input := requestQueryParams.Get(matcher.ParamName)
		if len(input) == 0 || fmt.Sprintf("%v", matcher.ExpectedValue) != input {
			return false
		}
	}
	return true
}

func isMatchingRequestBody(bodyMatchers []model.BodyMatcher, requestBody []byte) bool {
	if len(bodyMatchers) == 0 {
		return true
	}
	if len(requestBody) == 0 {
		return false
	}

	for _, matcher := range bodyMatchers {
		if !isPathMatching(matcher, &requestBody) {
			return false
		}
	}
	return true
}

func isMatchingRequestHeader(headerMatchers []model.HeaderMatcher, requestHeaders *http.Header) bool {
	if len(headerMatchers) == 0 {
		return true
	}
	if len(*requestHeaders) == 0 {
		return false
	}

	for _, matcher := range headerMatchers {
		headervalue := requestHeaders.Get(matcher.HeaderName)
		if len(headervalue) == 0 || matcher.ExpectedValue != headervalue {
			return false
		}
	}
	return true
}

func isPathMatching(matcher model.BodyMatcher, string *[]byte) bool {
	var value any
	if err := json.Unmarshal(*string, &value); err != nil {
		log.Printf("Failed to marshal request body. %s", err.Error())
	}
	path, err := jsonpath.Parse(matcher.JSONPathString)
	if err != nil {
		log.Printf("Failed to parse JsonPath. %s", err.Error())
	}

	nodes := path.Select(value)
	for _, node := range nodes {
		if matcher.ExpectedValue == fmt.Sprintf("%v", node) {
			return true
		}
	}
	return false
}
