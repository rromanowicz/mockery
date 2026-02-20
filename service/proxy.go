package service

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/rromanowicz/mockery/model"
)

type ProxyInt interface {
	CallExternal(mock *model.Mock, req *http.Request) (model.Mock, error)
}

type ProxyService struct{}

func (ps ProxyService) CallExternal(mock *model.Mock, req *http.Request) (model.Mock, error) {
	request, err := BuildRequest(mock, req)
	if err != nil {
		return *mock, err
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return *mock, err
	}
	defer response.Body.Close()

	var jsonb model.JSONB
	respBody, err := io.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}
	json.Unmarshal(respBody, &jsonb)

	return model.Mock{Response: model.Response{Status: response.StatusCode, Body: jsonb}}, nil
}

func BuildRequest(mock *model.Mock, req *http.Request) (*http.Request, error) {
	var request *http.Request
	targetURL := fmt.Sprintf("%s%s", mock.Response.Proxy.HostURL, req.URL)
	log.Printf("Calling proxy target: %s %s", req.Method, targetURL)
	request, err := http.NewRequest(req.Method, targetURL, req.Body)
	if err != nil {
		log.Printf("Failed to prepare request. Error: %v", err.Error())
		return request, err
	}

	for k, v := range req.Header {
		for i := range v {
			request.Header.Add(k, v[i])
		}
	}

	return request, nil
}
