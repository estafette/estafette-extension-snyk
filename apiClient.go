package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	foundation "github.com/estafette/estafette-foundation"
	"github.com/opentracing-contrib/go-stdlib/nethttp"
	"github.com/rs/zerolog/log"
	"github.com/sethgrid/pester"
)

type ApiClient interface {
	GetStatus(ctx context.Context, repoSource, repoOwner, repoName string) (status string, err error)
}

// NewApiClient returns a new ApiClient
func NewApiClient(apiToken string) ApiClient {
	return &apiClient{
		apiBaseURL: "https://snyk.io/api/v1/",
		apiToken:   apiToken,
	}
}

type apiClient struct {
	apiBaseURL string
	apiToken   string
}

func (c *apiClient) GetStatus(ctx context.Context, repoSource, repoOwner, repoName string) (status string, err error) {

	getStatusURL := fmt.Sprintf("%v/...", c.apiBaseURL)

	headers := map[string]string{
		"Content-Type": "application/json",
	}

	responseBody, err := c.getRequest(getStatusURL, nil, headers)
	if err != nil {
		log.Error().Err(err).Str("url", getStatusURL).Msgf("Failed retrieving snyk status")
		return
	}

	var statusResponse struct {
		Status string `json:"status"`
	}

	// unmarshal json body
	err = json.Unmarshal(responseBody, &statusResponse)
	if err != nil {
		log.Error().Err(err).Str("body", string(responseBody)).Str("url", getStatusURL).Msgf("Failed unmarshalling snyk status response")
		return
	}

	return statusResponse.Status, nil
}

func (c *apiClient) getRequest(uri string, requestBody io.Reader, headers map[string]string, allowedStatusCodes ...int) (responseBody []byte, err error) {
	return c.makeRequest("GET", uri, requestBody, headers, allowedStatusCodes...)
}

func (c *apiClient) postRequest(uri string, requestBody io.Reader, headers map[string]string, allowedStatusCodes ...int) (responseBody []byte, err error) {
	return c.makeRequest("POST", uri, requestBody, headers, allowedStatusCodes...)
}

func (c *apiClient) putRequest(uri string, requestBody io.Reader, headers map[string]string, allowedStatusCodes ...int) (responseBody []byte, err error) {
	return c.makeRequest("PUT", uri, requestBody, headers, allowedStatusCodes...)
}

func (c *apiClient) deleteRequest(uri string, requestBody io.Reader, headers map[string]string, allowedStatusCodes ...int) (responseBody []byte, err error) {
	return c.makeRequest("DELETE", uri, requestBody, headers, allowedStatusCodes...)
}

func (c *apiClient) makeRequest(method, uri string, requestBody io.Reader, headers map[string]string, allowedStatusCodes ...int) (responseBody []byte, err error) {

	// create client, in order to add headers
	client := pester.NewExtendedClient(&http.Client{Transport: &nethttp.Transport{}})
	client.MaxRetries = 3
	client.Backoff = pester.ExponentialJitterBackoff
	client.KeepLog = true
	client.Timeout = time.Second * 10

	request, err := http.NewRequest(method, uri, requestBody)
	if err != nil {
		return nil, err
	}

	// add headers
	containsAuthorizationHeader := false
	for k, v := range headers {
		request.Header.Add(k, v)
		if k == "Authorization" {
			containsAuthorizationHeader = true
		}
	}
	if !containsAuthorizationHeader {
		request.Header.Add("Authorization", fmt.Sprintf("token %v", c.apiToken))
	}

	// perform actual request
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if len(allowedStatusCodes) == 0 {
		allowedStatusCodes = []int{http.StatusOK}
	}

	if !foundation.IntArrayContains(allowedStatusCodes, response.StatusCode) {
		return nil, fmt.Errorf("%v %v responded with status code %v", method, uri, response.StatusCode)
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return
	}

	return body, nil
}
