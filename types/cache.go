package types

type IntelligenceExecutionState string

const (
	FULL_RUN IntelligenceExecutionState = "FULL_RUN"
	CACHED   IntelligenceExecutionState = "CACHED"
	NO_OP    IntelligenceExecutionState = "NO_OP"
)
