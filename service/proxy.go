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
	var request *http.Request
	targetURL := fmt.Sprintf("%s%s", mock.Response.Proxy.HostURL, req.URL)
	log.Printf("Calling proxy target: %s %s", req.Method, targetURL)
	request, _ = http.NewRequest(req.Method, targetURL, nil)

	// TODO: pass headers/query/body from original request to proxy call

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return *mock, nil
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
