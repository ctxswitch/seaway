package registry

import (
	"io"
	"net/http"
)

type ClientError struct {
	URL        string
	StatusCode int
	Message    string
}

func NewClientError(resp *http.Response) *ClientError {
	msg, error := io.ReadAll(resp.Body)
	if error != nil {
		msg = []byte("unable to read response body")
	}
	return &ClientError{
		URL:        resp.Request.URL.String(),
		StatusCode: resp.StatusCode,
		Message:    string(msg),
	}
}

func (e ClientError) Error() string {
	return e.Message
}

func IsNotFound(err error) bool {
	switch e := err.(type) {
	case *ClientError:
		return e.StatusCode == http.StatusNotFound
	}

	return false
}
