package types

// UploadCgRequest represents a request to upload code graph data including tests and chains
type UploadCgRequest struct {
	Identifier       Identifier     `json:"identifier"`
	Tests            []Test         `json:"tests"`
	Chains           []Chain        `json:"chains"`
	PathToTestNumMap map[string]int `json:"pathToTestNumMap"`
	TotalTests       int            `json:"totalTests"`
	FailedTests      []string       `json:"failedTests"`
}
