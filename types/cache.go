package types

import (
	"github.com/harness/ti-client/types/cache/dlc"
	"github.com/harness/ti-client/types/cache/gradle"
)

type IntelligenceExecutionState string

const (
	FULL_RUN  IntelligenceExecutionState = "FULL_RUN"
	OPTIMIZED IntelligenceExecutionState = "OPTIMIZED"
	DISABLED  IntelligenceExecutionState = "DISABLED"
)

type SavingsFeature string

const (
	BUILD_CACHE SavingsFeature = "build_cache"
	TI          SavingsFeature = "test_intelligence"
	DLC         SavingsFeature = "docker_layer_caching"
)

type SavingsRequest struct {
	GradleMetrics gradle.Metrics `json:"gradle_metrics"`
	DlcMetrics    dlc.Metrics    `json:"dlc_metrics"`
}

type SavingsOverview struct {
	FeatureName  SavingsFeature             `json:"feature_name"`
	TimeTakenMs  int64                      `json:"time_taken_ms"`
	TimeSavedMs  int64                      `json:"time_saved_ms"`
	BaselineMs   int64                      `json:"baseline_ms"`
	FeatureState IntelligenceExecutionState `json:"feature_state"`
}

type SavingsResponse struct {
	Overview   []SavingsOverview `json:"overview"`
	DlcMetrics dlc.Metrics       `json:"dlc_metrics"`
}
