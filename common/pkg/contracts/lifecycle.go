// Package contracts holds the cross-service model lifecycle contracts (plain structs, no database types).
package contracts

import "time"

// LifecycleAPIVersion is bumped on any breaking change to these contracts.
const LifecycleAPIVersion = "v1"

// Model lifecycle statuses shared between services.
const (
	StatusDraft      = "draft"
	StatusQueue      = "queue"
	StatusRunning    = "running"
	StatusProcessing = "processing"
	StatusCompleted  = "completed"
	StatusPublished  = "published"
	StatusFailed     = "failed"
	StatusCancelled  = "cancelled"
)

// allowedTransitions is the model lifecycle state machine.
var allowedTransitions = map[string]map[string]bool{
	StatusDraft:      {StatusQueue: true, StatusCancelled: true},
	StatusQueue:      {StatusRunning: true, StatusFailed: true, StatusCancelled: true},
	StatusRunning:    {StatusProcessing: true, StatusCompleted: true, StatusFailed: true, StatusCancelled: true},
	StatusProcessing: {StatusCompleted: true, StatusFailed: true},
	StatusCompleted:  {StatusPublished: true, StatusQueue: true}, // re-run allowed
	StatusPublished:  {StatusQueue: true},
	StatusFailed:     {StatusQueue: true}, // retry allowed
	StatusCancelled:  {StatusQueue: true},
}

// CanTransition reports whether from -> to is a legal move (from == to is always allowed).
func CanTransition(from, to string) bool {
	if from == to {
		return true
	}
	succ, ok := allowedTransitions[from]
	if !ok {
		return false
	}
	return succ[to]
}

// --- Internal lifecycle API DTOs (backend owns Model.status) ---

// MarkRunningRequest claims a queued model for execution on a compute instance.
// The backend transitions queue -> running atomically and returns Claimed=false
// (HTTP 409) if the model was no longer queued.
type MarkRunningRequest struct {
	WebserviceID uint `json:"webservice_id"`
}

// MarkFailedRequest fails an in-flight model with a human-readable reason.
type MarkFailedRequest struct {
	Reason string `json:"reason"`
}

// RunSessionRequest carries post-dispatch session metadata returned by the
// compute instance. It updates fields only; it does not change status.
type RunSessionRequest struct {
	SessionID   *int64  `json:"session_id,omitempty"`
	CallbackURL *string `json:"callback_url,omitempty"`
}

// TransitionResponse is the standard response for a lifecycle transition.
type TransitionResponse struct {
	ModelID uint   `json:"model_id"`
	Status  string `json:"status"`
	Claimed bool   `json:"claimed"`
}

// ActiveModel is a read-only projection of an in-flight model (queued or
// running) exposed by the backend so the webservice can run its instance
// reconciliation sweeps WITHOUT reading the models table. It carries only the
// facts those sweeps need, never the full DB row.
type ActiveModel struct {
	ModelID              uint       `json:"model_id"`
	WebserviceID         *uint      `json:"webservice_id,omitempty"`
	Status               string     `json:"status"`
	CalculationStartedAt *time.Time `json:"calculation_started_at,omitempty"`
}

// ActiveModelsResponse wraps the active-model projection list.
type ActiveModelsResponse struct {
	Models []ActiveModel `json:"models"`
}

// --- Versioned compute-dispatch queue payload ---

// TaskDispatchModelCalculation is the asynq task type for the compute pipeline.
const TaskDispatchModelCalculation = "dispatch_model_calculation"

// DispatchModelCalculation is the versioned payload enqueued by the backend and
// consumed by the webservice dispatcher. Version is set to LifecycleAPIVersion;
// consumers should reject or adapt unknown versions.
type DispatchModelCalculation struct {
	Version string                 `json:"version"`
	ModelID uint                   `json:"model_id"`
	UserID  string                 `json:"user_id"`
	Payload map[string]interface{} `json:"payload"`
}
