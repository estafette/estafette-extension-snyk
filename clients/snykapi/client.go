package snykapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/estafette/estafette-extension-snyk/api"
	foundation "github.com/estafette/estafette-foundation"
	"github.com/opentracing-contrib/go-stdlib/nethttp"
	"github.com/rs/zerolog/log"
	"github.com/sethgrid/pester"
)

type Client interface {
	GetOrganizations(ctx context.Context) (organizations []api.Organization, err error)
	GetProjects(ctx context.Context, organization api.Organization) (projects []api.Project, err error)
	GetStatus(ctx context.Context, repoSource, repoOwner, repoName string) (status string, err error)
}

// NewClient returns a new snykapi.Client
func NewClient(apiToken string) Client {
	return &client{
		apiBaseURL: "https://snyk.io/api/v1",
		apiToken:   apiToken,
	}
}

type client struct {
	apiBaseURL string
	apiToken   string
}

func (c *client) GetOrganizations(ctx context.Context) (organizations []api.Organization, err error) {

	// https://snyk.docs.apiary.io/#reference/organizations/list-all-the-organizations-a-user-belongs-to
	getOrgsURL := fmt.Sprintf("%v/orgs", c.apiBaseURL)

	headers := map[string]string{
		"Content-Type": "application/json",
	}

	responseBody, err := c.getRequest(getOrgsURL, nil, headers)
	if err != nil {
		return
	}

	var response struct {
		Organizations []api.Organization `json:"orgs"`
	}

	// unmarshal json body
	err = json.Unmarshal(responseBody, &response)
	if err != nil {
		log.Error().Err(err).Str("body", string(responseBody)).Str("url", getOrgsURL).Msgf("Failed unmarshalling snyk getOrgsURL response")
		return
	}

	return response.Organizations, nil
}

func (c *client) GetProjects(ctx context.Context, organization api.Organization) (projects []api.Project, err error) {

	// https://snyk.docs.apiary.io/#reference/projects/all-projects/list-all-projects
	getProjectsURL := fmt.Sprintf("%v/org/%v/projects", c.apiBaseURL, organization.ID)

	headers := map[string]string{
		"Content-Type": "application/json",
	}

	responseBody, err := c.getRequest(getProjectsURL, nil, headers)
	if err != nil {
		return
	}

	var response struct {
		Projects []api.Project `json:"projects"`
	}

	// unmarshal json body
	err = json.Unmarshal(responseBody, &response)
	if err != nil {
		log.Error().Err(err).Str("body", string(responseBody)).Str("url", getProjectsURL).Msgf("Failed unmarshalling snyk getProjectsURL response")
		return
	}

	return response.Projects, nil
}

func (c *client) GetStatus(ctx context.Context, repoSource, repoOwner, repoName string) (status string, err error) {

	getStatusURL := fmt.Sprintf("%v/...", c.apiBaseURL)

	headers := map[string]string{
		"Content-Type": "application/json",
	}

	responseBody, err := c.getRequest(getStatusURL, nil, headers)
	if err != nil {
		return
	}

	var response struct {
		Status string `json:"status"`
	}

	// unmarshal json body
	err = json.Unmarshal(responseBody, &response)
	if err != nil {
		log.Error().Err(err).Str("body", string(responseBody)).Str("url", getStatusURL).Msgf("Failed unmarshalling snyk status response")
		return
	}

	return response.Status, nil
}

func (c *client) getRequest(uri string, requestBody io.Reader, headers map[string]string, allowedStatusCodes ...int) (responseBody []byte, err error) {
	return c.makeRequest("GET", uri, requestBody, headers, allowedStatusCodes...)
}

func (c *client) postRequest(uri string, requestBody io.Reader, headers map[string]string, allowedStatusCodes ...int) (responseBody []byte, err error) {
	return c.makeRequest("POST", uri, requestBody, headers, allowedStatusCodes...)
}

func (c *client) putRequest(uri string, requestBody io.Reader, headers map[string]string, allowedStatusCodes ...int) (responseBody []byte, err error) {
	return c.makeRequest("PUT", uri, requestBody, headers, allowedStatusCodes...)
}

func (c *client) deleteRequest(uri string, requestBody io.Reader, headers map[string]string, allowedStatusCodes ...int) (responseBody []byte, err error) {
	return c.makeRequest("DELETE", uri, requestBody, headers, allowedStatusCodes...)
}

func (c *client) makeRequest(method, uri string, requestBody io.Reader, headers map[string]string, allowedStatusCodes ...int) (responseBody []byte, err error) {

	if len(allowedStatusCodes) == 0 {
		allowedStatusCodes = []int{http.StatusOK}
	}

	log.Debug().Interface("headers", headers).Interface("allowedStatusCodes", allowedStatusCodes).Msgf("%v %v | start", method, uri)

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

	log.Debug().Msgf("%v %v | new request created", method, uri)

	// perform actual request
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	log.Debug().Int("statusCode", response.StatusCode).Msgf("%v %v | request done", method, uri)

	if !foundation.IntArrayContains(allowedStatusCodes, response.StatusCode) {
		return nil, fmt.Errorf("%v %v responded with status code %v", method, uri, response.StatusCode)
	}

	log.Debug().Msgf("%v %v | status code checked", method, uri)

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return
	}

	log.Debug().Interface("headers", headers).Interface("allowedStatusCodes", allowedStatusCodes).Int("statusCode", response.StatusCode).Str("body", string(body)).Msgf("%v %v | finish", method, uri)

	return body, nil
}
