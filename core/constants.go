package core

// ConditionType is a type for conditions that can be asserted
type ConditionType string

const (
	// Completed implies experiment is complete
	Completed ConditionType = "completed"
	// Successful     ConditionType = "successful"
	// Failure        ConditionType = "failure"
	// HandlerFailure ConditionType = "handlerFailure"

	// WinnerFound implies experiment has found a winner
	WinnerFound ConditionType = "winnerFound"
	// CandidateWon   ConditionType = "candidateWon"
	// BaselineWon    ConditionType = "baselineWon"
	// NoWinner       ConditionType = "noWinner"
)
