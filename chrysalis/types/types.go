package types

// UploadCgRequest represents a request to upload code graph data including tests and chains
type UploadCgRequest struct {
	Identifier Identifier `json:"identifier"`
	Tests      []Test     `json:"tests"`
	Chains     []Chain    `json:"chains"`
}
