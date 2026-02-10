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
	regConfigImport, _ := regexp.Compile("/config/import")
	regConfigExport, _ := regexp.Compile("/config/export")
	regConfig, _ := regexp.Compile("/config.*")
	reg, _ := regexp.Compile("/.*")
	handler.HandleFunc(regHelp, handleHelp)
	handler.HandleFunc(regConfigList, handleConfigList(ctx))
	handler.HandleFunc(regConfigImport, handleConfigImport(ctx))
	handler.HandleFunc(regConfigExport, handleConfigExport(ctx))
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
					resp, _ := mock.ResponseBody.Value()
					rw.Write(resp.([]byte))
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

func handleConfigImport(ctx context.Context) func(rw http.ResponseWriter, req *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		files, err := ctx.MockService.Import()
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			rw.Write([]byte(err.Error()))
		} else {
			rw.Header().Set("Content-Type", "application/json")
			rw.WriteHeader(http.StatusOK)
			resp, _ := json.Marshal(files)
			rw.Write(resp)
		}
	}
}

func handleConfigExport(ctx context.Context) func(rw http.ResponseWriter, req *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		files, err := ctx.MockService.Export()
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			rw.Write([]byte(err.Error()))
		} else {
			rw.Header().Set("Content-Type", "application/json")
			rw.WriteHeader(http.StatusOK)
			resp, _ := json.Marshal(files)
			rw.Write(resp)
		}
	}
}

func handleAll(ctx context.Context) func(rw http.ResponseWriter, req *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		mocks, err := fetchMocks(ctx, req.Method, req.URL.Path)
		if err != nil {
			rw.WriteHeader(http.StatusNotFound)
			rw.Write([]byte(err.Error()))
			return
		}
		mock, err := filterMocks(mocks, req)
		if err != nil {
			rw.WriteHeader(http.StatusNotFound)
			rw.Write([]byte(err.Error()))
		} else {
			log.Printf("Matched Mock[id=%v]", mock.ID)
			rw.Header().Set("Content-Type", "application/json")
			rw.WriteHeader(mock.ResponseStatus)
			response, _ := json.Marshal(mock.ResponseBody)
			rw.Write(response)
		}
	}
}

func fetchMocks(ctx context.Context, method string, path string) ([]model.Mock, error) {
	var mocks []model.Mock
	var err error
	mocks, err = ctx.MockService.Get(method, path)
	if err != nil {
		return []model.Mock{}, err
	}

	if len(mocks) == 0 {
		regexMatchers, err := ctx.MockService.GetRegexpMatchers(method)
		if err != nil {
			return []model.Mock{}, err
		}
		var ids []int64
		for i := range regexMatchers {
			if regexMatchers[i].RegexPath.Compile().MatchString(path) {
				ids = append(ids, regexMatchers[i].ID)
			}
		}
		if len(ids) == 0 {
			return []model.Mock{}, fmt.Errorf("no mocks found")
		}
		mocks, err = ctx.MockService.GetByIds(ids)
		if err != nil {
			return []model.Mock{}, err
		}
	}

	return mocks, nil
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
	if len(matchedMocks) > 1 {
		var ids []string
		for i := range matchedMocks {
			ids = append(ids, fmt.Sprint(matchedMocks[i].ID))
		}
		log.Printf("Multiple mocks matched [%s].", strings.Join(ids, ","))
	}

	return *matchedMocks[0], nil
}

func isMatchingRequestQuery(queryMatchers model.Matchers, requestQueryParams url.Values) bool {
	if len(queryMatchers) == 0 {
		return true
	}
	if len(requestQueryParams) == 0 {
		return false
	}

	for _, matcher := range queryMatchers {
		input := requestQueryParams.Get(matcher.Key)
		if len(input) == 0 || fmt.Sprintf("%v", matcher.Value) != input {
			return false
		}
	}
	return true
}

func isMatchingRequestBody(bodyMatchers model.Matchers, requestBody []byte) bool {
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

func isMatchingRequestHeader(headerMatchers model.Matchers, requestHeaders *http.Header) bool {
	if len(headerMatchers) == 0 {
		return true
	}
	if len(*requestHeaders) == 0 {
		return false
	}

	for _, matcher := range headerMatchers {
		headervalue := requestHeaders.Get(matcher.Key)
		if len(headervalue) == 0 || matcher.Value != headervalue {
			return false
		}
	}
	return true
}

func isPathMatching(matcher model.Matcher, string *[]byte) bool {
	var value any
	if err := json.Unmarshal(*string, &value); err != nil {
		log.Printf("Failed to marshal request body. %s", err.Error())
	}
	path, err := jsonpath.Parse(matcher.Key)
	if err != nil {
		log.Printf("Failed to parse JsonPath. %s", err.Error())
	}

	nodes := path.Select(value)
	for _, node := range nodes {
		if matcher.Value == fmt.Sprintf("%v", node) {
			return true
		}
	}
	return false
}
