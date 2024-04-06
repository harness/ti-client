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
)

type SavingsRequest struct {
	GradleProfile gradle.Profile `json:"gradle_profile"`
}
