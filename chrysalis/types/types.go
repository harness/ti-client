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

type SelectAndSplitRequest struct {
	Files            map[string]uint64 `json:"files"`
	ExecutionContext map[string]string `json:"executionContext"`
	AllTests         []string          `json:"allTests"`
	SplitConfig      SplitConfig       `json:"splitConfig"`
}

type SplitConfig struct {
	MaxStages        int    `json:"maxStages"`
	TimeDataKey      string `json:"timeDataKey"`      // "class_name", "name", "file_name", "suite_name"
	MaxTestsPerStage int    `json:"maxTestsPerStage"` // 0 = no limit; >0 caps tests per stage
}

type SelectAndSplitResponse struct {
	SkipTests           []string      `json:"skipTests"`
	FailedTests         []string      `json:"failedTests"`
	Parallelism         int           `json:"parallelism"`
	TotalSelectedTests  int           `json:"totalSelectedTests"`
	EstimatedDurationMs map[int]int64 `json:"estimatedDurationMs"`
}

type StageBatchRequest struct {
	StageIndex int `json:"stageIndex"`
}

type StageBatchResponse struct {
	TestIDs             []string `json:"testIds"`
	StageIndex          int      `json:"stageIndex"`
	TotalStages         int      `json:"totalStages"`
	EstimatedDurationMs int64    `json:"estimatedDurationMs"`
}
