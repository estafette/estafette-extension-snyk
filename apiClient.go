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
	"github.com/sethgrid/pester"
)

type SnykAPIClient interface {
	OrganizationProjects(ctx context.Context, orgID string) (Projects, error)
	ProjectVulnerabilities(ctx context.Context, orgID, projectID string, filters *projectFilters) (Vulnerabilities, error)
}

//NewApiClient returns a new SnykApiClient
func NewSnykAPIClient(apiToken string) SnykAPIClient {
	return &apiClient{
		apiBaseURL: "https://snyk.io/api/v1/",
		apiHeaders: map[string]string{
			"Authorization": fmt.Sprintf("token %s", apiToken),
			"Content-Type":  "application/json",
		},
	}
}

type apiClient struct {
	apiBaseURL string
	apiHeaders map[string]string
}

func (c *apiClient) OrganizationProjects(ctx context.Context, orgID string) (Projects, error) {

	projects := &projectsWrapper{}

	organizationProjectsURL := fmt.Sprintf("%v/org/%s/projects", c.apiBaseURL, orgID)

	err := c.getRequest(organizationProjectsURL, nil, c.apiHeaders, projects, 200)
	if err != nil {
		return nil, fmt.Errorf("Failed retrieving projects from snyk: %s", err.Error())
	}

	return projects.Projects, nil

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

	err = c.postRequest(projectVulnerabilitiesURL, body, c.apiHeaders, pIssues, 200)
	if err != nil {
		return nil, fmt.Errorf("Failed retrieving vulnerabilities from snyk: %s", err.Error())
	}

	return pIssues.Issues.Vulnerabilities, nil
}

func (c *apiClient) getRequest(uri string, requestBody io.Reader, headers map[string]string, wrapper interface{}, allowedStatusCodes ...int) (err error) {
	return c.makeRequest("GET", uri, requestBody, headers, wrapper, allowedStatusCodes...)
}

func (c *apiClient) postRequest(uri string, requestBody io.Reader, headers map[string]string, wrapper interface{}, allowedStatusCodes ...int) (err error) {
	return c.makeRequest("POST", uri, requestBody, headers, wrapper, allowedStatusCodes...)
}

func (c *apiClient) makeRequest(method, uri string, requestBody io.Reader, headers map[string]string, wrapper interface{}, allowedStatusCodes ...int) (err error) {

	// create client, in order to add headers
	client := pester.NewExtendedClient(&http.Client{Transport: &nethttp.Transport{}})
	client.MaxRetries = 3
	client.Backoff = pester.ExponentialJitterBackoff
	client.KeepLog = true
	client.Timeout = time.Second * 10

	request, err := http.NewRequest(method, uri, requestBody)
	if err != nil {
		return err
	}

	// add headers
	for k, v := range headers {
		request.Header.Add(k, v)
	}

	// perform actual request
	response, err := client.Do(request)
	if err != nil {
		return err
	}

	defer response.Body.Close()

	if len(allowedStatusCodes) == 0 {
		allowedStatusCodes = []int{http.StatusOK}
	}

	if !foundation.IntArrayContains(allowedStatusCodes, response.StatusCode) {
		return fmt.Errorf("%v %v responded with status code %v", method, uri, response.StatusCode)
	}

	bodyBytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("Error reading body of snyk response: %s", err.Error())
	}

	err = json.Unmarshal(bodyBytes, &wrapper)
	if err != nil {
		return fmt.Errorf("Failed unmarshalling snyk response body: %s", err.Error())
	}

	return nil
}
