package trakt

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const apiURL = "https://api.trakt.tv"
const apiVersion = "2"

const apiDefaultDatapath = "/data"

// TraktAPI exposes the trakt API globally
var TraktAPI API

// API represents the Trakt API
type API struct {
	DataPath        string
	ClientID        string
	ClientSecret    string
	IsAuthenticated bool
	accessToken     string
}

type apiResponse struct {
	StatusCode int
	Body       []byte
}

func (api *API) validate() error {
	if api.DataPath == "" {
		api.DataPath = apiDefaultDatapath
	}

	return nil
}

func (api *API) sendRequest(method, url string, payload interface{}) (*apiResponse, error) {
	if method == "" {
		method = "GET"
	}

	err := api.validate()
	if err != nil {
		return nil, err
	}

	requestURL := apiURL + url
	var reqPayload []byte

	if payload != nil {
		reqPayload, err = json.Marshal(payload)
		if err != nil {
			return nil, err
		}
	}

	apiClient := http.Client{
		Timeout: time.Second * 5, // Timeout after 5 seconds
	}

	req, err := http.NewRequest(method, requestURL, bytes.NewReader(reqPayload))
	if err != nil {
		return nil, err
	}

	req.Header.Set("content-type", "application/json")
	req.Header.Set("trakt-api-version", apiVersion)
	req.Header.Set("trakt-api-key", api.ClientID)
	if api.accessToken != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", api.accessToken))
	}

	response, err := apiClient.Do(req)
	if err != nil {
		return nil, err
	}

	if response.Body != nil {
		defer response.Body.Close()
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	var returnVal = apiResponse{}
	returnVal.StatusCode = response.StatusCode
	returnVal.Body = body

	return &returnVal, nil
}
