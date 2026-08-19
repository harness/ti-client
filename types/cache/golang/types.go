package golang

// Metrics is uploaded via types.SavingsRequest for Go Build Intelligence.
type Metrics struct {
	Reports []Report `json:"reports"`
}

// Report mirrors the JSON file emitted by go-cache-proxy at close.
type Report struct {
	Version         int    `json:"version"`
	Mode            string `json:"mode"`
	Gets            int64  `json:"gets"`
	Hits            int64  `json:"hits"`
	Misses          int64  `json:"misses"`
	Puts            int64  `json:"puts"`
	BytesRestored   int64  `json:"bytes_restored"`
	BytesStored     int64  `json:"bytes_stored"`
	DurationMs      int64  `json:"duration_ms"`
	StartedAtUnixMs int64  `json:"started_at_unix_ms"`
	EndedAtUnixMs   int64  `json:"ended_at_unix_ms"`
}
