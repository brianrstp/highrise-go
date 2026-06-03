package highrise

import "fmt"

type ResponseError struct {
	Message string
}

func (e *ResponseError) Error() string {
	return fmt.Sprintf("highrise: %s", e.Message)
}

func NewResponseError(msg string) *ResponseError {
	return &ResponseError{Message: msg}
}

type ConnectionError struct {
	Err error
}

func (e *ConnectionError) Error() string {
	return fmt.Sprintf("connection error: %v", e.Err)
}

func (e *ConnectionError) Unwrap() error {
	return e.Err
}
