package protocol

import "encoding/json"

// Thread represents a minimal thread descriptor used in thread responses.
type Thread struct {
	ID string `json:"id,omitempty"`
}

// ThreadResponse is the shared shape for thread/start and thread/resume responses.
type ThreadResponse struct {
	ThreadID string  `json:"threadId,omitempty"`
	Thread   *Thread `json:"thread,omitempty"`
}

// ThreadStartResponse is the response payload for thread/start.
type ThreadStartResponse = ThreadResponse

// ThreadResumeResponse is the response payload for thread/resume.
type ThreadResumeResponse = ThreadResponse

// TurnNotification describes turn/started and turn/completed notifications.
type TurnNotification struct {
	ThreadID string                `json:"threadId,omitempty"`
	Turn     *TurnNotificationTurn `json:"turn,omitempty"`
}

// TurnStartedNotification is the payload for turn/started.
type TurnStartedNotification = TurnNotification

// TurnCompletedNotification is the payload for turn/completed.
type TurnCompletedNotification = TurnNotification

// TurnNotificationTurn describes a turn summary in notifications.
type TurnNotificationTurn struct {
	ID     string                 `json:"id,omitempty"`
	Status string                 `json:"status,omitempty"`
	Error  *TurnNotificationError `json:"error,omitempty"`
}

// TurnNotificationError describes a turn error payload.
type TurnNotificationError struct {
	Message string `json:"message,omitempty"`
}

// ItemCompletedNotification is the payload for item/completed.
type ItemCompletedNotification struct {
	ThreadID string          `json:"threadId,omitempty"`
	Item     json.RawMessage `json:"item,omitempty"`
}

// ErrorNotification is the payload for error notifications.
type ErrorNotification struct {
	ThreadID  string                 `json:"threadId,omitempty"`
	WillRetry *bool                  `json:"willRetry,omitempty"`
	Error     *TurnNotificationError `json:"error,omitempty"`
}
