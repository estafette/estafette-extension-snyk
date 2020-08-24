package main

import (
	"bytes"
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

type SnykApiClient interface {
	OrganizationProjects(ctx context.Context, orgID string) (Projects, error)
	ProjectVulnerabilities(ctx context.Context, orgID, projectID string, filters *projectFilters) (Vulnerabilities, error)
}

//NewApiClient returns a new SnykApiClient
func NewApiClient(apiToken string) SnykApiClient {
	return &apiClient{
		apiBaseURL: "https://snyk.io/api/v1/",
		apiToken:   apiToken,
	}
}

type apiClient struct {
	apiBaseURL string
	apiToken   string
}

func (c *apiClient) OrganizationProjects(ctx context.Context, orgID string) (Projects, error) {

	var wrapper struct {
		Projects `json:"projects"`
	}

	organizationProjectsURL := fmt.Sprintf("%v/org/%s/projects", c.apiBaseURL, orgID)

	headers := map[string]string{
		"Authorization": fmt.Sprintf("token %s", c.apiToken),
		"Content-Type":  "application/json",
	}

	response, err := c.getRequest(organizationProjectsURL, nil, headers, 200)
	if err != nil {
		log.Fatal().Msgf("Failed retrieving projects from snyk")
	}

	defer response.Body.Close()

	bodyBytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Info().Msgf(err.Error())
		log.Fatal().Msgf("Error reading body of snyk response")
	}

	// unmarshal json body
	err = json.Unmarshal(bodyBytes, &wrapper)
	if err != nil {
		log.Info().Msgf(err.Error())
		log.Fatal().Msgf("Failed unmarshalling snyk response body")
	}

	return wrapper.Projects, nil

}

func (c *apiClient) ProjectVulnerabilities(ctx context.Context, orgID, projectID string, filters *projectFilters) (Vulnerabilities, error) {
	pIssues := &projectIssuesWrapper{}

	if filters == nil {
		filters = defaultFilters()
	}

	data := struct {
		Filters *projectFilters `json:"filters"`
	}{
		filters,
	}

	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	body := bytes.NewReader(jsonBytes)

	projectVulnerabilitiesURL := fmt.Sprintf("%v/org/%s/project/%s/issues", c.apiBaseURL, orgID, projectID)

	headers := map[string]string{
		"Authorization": fmt.Sprintf("token %s", c.apiToken),
		"Content-Type":  "application/json",
	}

	response, err := c.postRequest(projectVulnerabilitiesURL, body, headers, 200)
	if err != nil {
		log.Fatal().Msgf("Failed retrieving vulnerabilities from snyk")
	}

	defer response.Body.Close()

	bodyBytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Info().Msgf(err.Error())
		log.Fatal().Msgf("Error reading body of snyk response")
	}

	err = json.Unmarshal(bodyBytes, &pIssues)
	if err != nil {
		log.Info().Msgf(err.Error())
		log.Fatal().Msgf("Failed unmarshalling snyk response body")
	}

	return pIssues.Issues.Vulnerabilities, nil
}

func (c *apiClient) getRequest(uri string, requestBody io.Reader, headers map[string]string, allowedStatusCodes ...int) (resp *http.Response, err error) {
	return c.makeRequest("GET", uri, requestBody, headers, allowedStatusCodes...)
}

func (c *apiClient) postRequest(uri string, requestBody io.Reader, headers map[string]string, allowedStatusCodes ...int) (resp *http.Response, err error) {
	return c.makeRequest("POST", uri, requestBody, headers, allowedStatusCodes...)
}

func (c *apiClient) makeRequest(method, uri string, requestBody io.Reader, headers map[string]string, allowedStatusCodes ...int) (resp *http.Response, err error) {

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

	if len(allowedStatusCodes) == 0 {
		allowedStatusCodes = []int{http.StatusOK}
	}

	if !foundation.IntArrayContains(allowedStatusCodes, response.StatusCode) {
		return nil, fmt.Errorf("%v %v responded with status code %v", method, uri, response.StatusCode)
	}

	return response, nil
}
