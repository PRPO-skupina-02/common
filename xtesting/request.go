package xtesting

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func NewTestingRequest(t *testing.T, targetURL string, method string, body any) *http.Request {
	var req *http.Request
	var err error

	if body == nil {
		req, err = http.NewRequest(method, targetURL, nil)
		require.NoError(t, err)
	} else {
		jsonBody, err := json.Marshal(body)
		require.NoError(t, err)

		bodyReader := bytes.NewReader(jsonBody)

		req, err = http.NewRequest(method, targetURL, bodyReader)
		require.NoError(t, err)
	}

	return req
}
