package main

import (
	"time"
)

// Project is an object which is a package that is actively
// tracked by Snyk.
type Project struct {
	Name                  string    `json:"name"`
	ID                    string    `json:"id"`
	Created               time.Time `json:"created"`
	Origin                string    `json:"origin"`
	Type                  string    `json:"type"`
	ReadOnly              bool      `json:"readOnly"`
	TestFrequency         string    `json:"testFrequency"`
	TotalDependencies     int       `json:"totalDependencies"`
	IssueCountsBySeverity struct {
		Low    int `json:"low"`
		High   int `json:"high"`
		Medium int `json:"medium"`
	} `json:"issueCountsBySeverity"`
	LastTestedDate time.Time `json:"lastTestedDate"`
	ImportingUser  struct {
		ID       string `json:"id"`
		Name     string `json:"name"`
		Username string `json:"username"`
		Email    string `json:"email"`
	} `json:"importingUser"`
	ImageID  string `json:"imageId,omitempty"`
	ImageTag string `json:"imageTag,omitempty"`
}

// Projects is a collection of idividual Project objects which is a package
// that is actively tracked by Snyk.
type Projects = []Project

// Vulnerabilities are  from a given project's Issues.
type Vulnerabilities []struct {
	ID             string   `json:"id"`
	URL            string   `json:"url"`
	Title          string   `json:"title"`
	Type           string   `json:"type"`
	Description    string   `json:"description"`
	From           []string `json:"from"`
	Package        string   `json:"package"`
	Version        string   `json:"version"`
	Severity       string   `json:"severity"`
	Language       string   `json:"language"`
	PackageManager string   `json:"packageManager"`
	Semver         struct {
		Unaffected string `json:"unaffected"`
		//Vulnerable string `json:"vulnerable"`
	} `json:"semver"`
	PublicationTime time.Time `json:"publicationTime"`
	DisclosureTime  time.Time `json:"disclosureTime"`
	IsUpgradable    bool      `json:"isUpgradable"`
	IsPinnable      bool      `json:"isPinnable"`
	IsPatchable     bool      `json:"isPatchable"`
	Identifiers     struct {
		CVE         []interface{} `json:"CVE"`
		CWE         []interface{} `json:"CWE"`
		OSVDB       []interface{} `json:"OSVDB"`
		ALTERNATIVE []interface{} `json:"ALTERNATIVE"`
	} `json:"identifiers"`
	Credit    []string `json:"credit"`
	CVSSv3    string   `json:"CVSSv3"`
	CvssScore float64  `json:"cvssScore"`
	Patches   []struct {
		ID               string        `json:"id"`
		Urls             []string      `json:"urls"`
		Version          string        `json:"version"`
		Comments         []interface{} `json:"comments"`
		ModificationTime time.Time     `json:"modificationTime"`
	} `json:"patches"`
	IsIgnored   bool          `json:"isIgnored"`
	IsPatched   bool          `json:"isPatched"`
	UpgradePath []interface{} `json:"upgradePath"`
	Ignored     []struct {
		Reason  string    `json:"reason"`
		Expires time.Time `json:"expires"`
		Source  string    `json:"source"`
	} `json:"ignored,omitempty"`
	Patched []struct {
		Patched time.Time `json:"patched"`
	} `json:"patched,omitempty"`
}

type projectIssuesWrapper struct {
	Ok     bool `json:"ok"`
	Issues struct {
		Vulnerabilities Vulnerabilities `json:"vulnerabilities"`
		Licenses        []interface{}   `json:"licenses"`
	} `json:"issues"`
	DependencyCount int    `json:"dependencyCount"`
	PackageManager  string `json:"packageManager"`
}

type projectFilters struct {
	Severities []string `json:"severities"`
	Types      []string `json:"types"`
	Ignored    bool     `json:"ignored"`
	Patched    bool     `json:"patched"`
}

func defaultFilters() *projectFilters {
	return &projectFilters{
		Severities: []string{
			"high", "medium", "low",
		},
		Types: []string{
			"vuln", "license",
		},
		Ignored: false,
		Patched: false,
	}
}
