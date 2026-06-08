package highrise

import "fmt"

// ResponseError represents an error returned by the Highrise server
// in response to a bot action request.
type ResponseError struct {
	Message string
}

func (e *ResponseError) Error() string {
	return fmt.Sprintf("highrise: %s", e.Message)
}

// NewResponseError creates a new ResponseError with the given message.
func NewResponseError(msg string) *ResponseError {
	return &ResponseError{Message: msg}
}

// ConnectionError wraps errors from the WebSocket connection layer.
// It includes the request type and RID for easier debugging.
type ConnectionError struct {
	ReqType string
	RID     string
	Err     error
}

func (e *ConnectionError) Error() string {
	if e.ReqType != "" {
		return fmt.Sprintf("connection error [%s/%s]: %v", e.ReqType, e.RID, e.Err)
	}
	return fmt.Sprintf("connection error: %v", e.Err)
}

func (e *ConnectionError) Unwrap() error {
	return e.Err
}
