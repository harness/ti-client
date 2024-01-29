package types

type IntelligenceExecutionState string

const (
	FULL_RUN  IntelligenceExecutionState = "FULL_RUN"
	OPTIMIZED IntelligenceExecutionState = "OPTIMIZED"
	DISABLED  IntelligenceExecutionState = "DISABLED"
)
