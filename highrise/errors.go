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
type ConnectionError struct {
	Err error
}

func (e *ConnectionError) Error() string {
	return fmt.Sprintf("connection error: %v", e.Err)
}

func (e *ConnectionError) Unwrap() error {
	return e.Err
}
