package errors

import (
	"fmt"
	"io"
	"net/http"
)

type ProxmoxError struct {
	StatusCode int
	Err        string
	Message    string
}

type ProxmoxErrorBody struct {
	Data   string            `json:"data"`
	Errors map[string]string `json:"errors"`
}

func (e *ProxmoxError) Error() string {
	return fmt.Sprintf("%s - %s", e.Err, e.Message)
}

func ApiError(h *http.Response, e error) error {
	var msg string
	b, err := io.ReadAll(h.Body)
	if err != nil {
		msg = "Unknown"
	} else {
		msg = string(b)
	}
	return &ProxmoxError{
		StatusCode: h.StatusCode,
		Err:        e.Error(),
		Message:    msg,
	}
}
