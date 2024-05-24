package types

import (
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
}
