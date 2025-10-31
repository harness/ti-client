package types

// UploadCgRequest represents a request to upload code graph data including tests and chains
type UploadCgRequest struct {
	Identifier       Identifier     `json:"identifier"`
	Tests            []Test         `json:"tests"`
	Chains           []Chain        `json:"chains"`
	PathToTestNumMap map[string]int `json:"pathToTestNumMap"`
	TotalTests       int            `json:"totalTests"`
	PreviousFailures []string       `json:"previousFailures"`
}

type SkipTestsRequest struct {
	Files            map[string]uint64 `json:"files"`
	ExecutionContext map[string]string `json:"executionContext"`
}
