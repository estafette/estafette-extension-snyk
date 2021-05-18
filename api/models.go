package api

type APITokenCredentials struct {
	Name                 string                                  `json:"name,omitempty"`
	Type                 string                                  `json:"type,omitempty"`
	AdditionalProperties APITokenCredentialsAdditionalProperties `json:"additionalProperties,omitempty"`
}

type APITokenCredentialsAdditionalProperties struct {
	Token string `json:"token,omitempty"`
}

type Organization struct {
	Name  string `json:"name,omitempty"`
	ID    string `json:"id,omitempty"`
	Slug  string `json:"slug,omitempty"`
	URL   string `json:"url,omitempty"`
	Group *Group `json:"group,omitempty"`
}

type Group struct {
	Name string `json:"name,omitempty"`
	ID   string `json:"id,omitempty"`
}

type Project struct {
	Name                  string           `json:"name,omitempty"`
	ID                    string           `json:"id,omitempty"`
	Tags                  []Tag            `json:"tags,omitempty"`
	Branch                string           `json:"branch,omitempty"`
	RemoteRepoUrl         string           `json:"remoteRepoUrl,omitempty"`
	LastTestedDate        string           `json:"lastTestedDate,omitempty"`
	TotalDependencies     int              `json:"totalDependencies,omitempty"`
	IssueCountsBySeverity map[Severity]int `json:"issueCountsBySeverity,omitempty"`
	Origin                string           `json:"origin,omitempty"`
	Type                  string           `json:"type,omitempty"`

	// {
	// 	"created": "2018-10-29T09:50:54.014Z",
	// 	"readOnly": false,
	// 	"testFrequency": "daily",
	// 	"totalDependencies": 438,
	// 	"issueCountsBySeverity": {
	// 		"low": 8,
	// 		"high": 13,
	// 		"medium": 15
	// 	},
	// 	"remoteRepoUrl": "https://github.com/snyk/goof.git",
	// 	"lastTestedDate": "2019-02-05T06:21:00.000Z",
	// 	"isMonitored": true,
	// 	"owner": {
	// 		"id": "e713cf94-bb02-4ea0-89d9-613cce0caed2",
	// 		"name": "example-user@snyk.io",
	// 		"username": "exampleUser",
	// 		"email": "example-user@snyk.io"
	// 	},
	// },

}

type Severity string

const (
	SeverityLow    Severity = "low"
	SeverityHigh   Severity = "high"
	SeverityMedium Severity = "medium"
)

type Language int

const (
	LanguageUnknown Language = iota
	LanguageGolang
	LanguageNode
	LanguageMaven
	LanguageDotnet
	LanguagePython
)

type Tag struct {
	Key   string `json:"key,omitempty"`
	Value string `json:"value,omitempty"`
}

type SnykFlags struct {
	Language          Language
	FailOn            string
	File              string
	PackagesFolder    string
	SeverityThreshold string

	MavenMirrorUrl    string
	MavenUsername     string
	MavenPassword     string
	MavenUpdateParent bool

	BuildVersionMajor string
	BuildVersionMinor string
}
