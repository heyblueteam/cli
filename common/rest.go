package common

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// RESTError is a non-2xx response from the REST /v1 API. The API returns
// {"error": "..."} (and {"issues": [...]} for input validation failures).
type RESTError struct {
	Status int
	Method string
	Path   string
}

func (e *RESTError) Error() string {
	return fmt.Sprintf("API error (HTTP %d): %s %s", e.Status, e.Method, e.Path)
}

// restBaseURL derives the REST base (scheme://host/v1) from the configured
// GraphQL API URL. The REST router is mounted at /v1 on the same server that
// serves /graphql, so any deployment that overrides API_URL keeps working.
func (c *Client) restBaseURL() string {
	base := strings.TrimSuffix(c.config.APIUrl, "/graphql")
	base = strings.TrimSuffix(base, "/")
	return base + "/v1"
}

// REST performs one REST /v1 round trip: method + path on the API host that
// serves GraphQL, optional query params, optional JSON body, JSON response
// decoded into out (when out is non-nil). OAuth sessions retry exactly once
// on a 401, mirroring ExecuteQuery.
func (c *Client) REST(method, path string, query url.Values, body interface{}, out interface{}) error {
	data, status, err := c.rest(method, path, query, body)
	if status == http.StatusUnauthorized && c.config.OAuthAccessToken != "" {
		if refreshErr := c.refreshOAuth(); refreshErr != nil {
			return refreshErr
		}
		data, status, err = c.rest(method, path, query, body)
	}
	if err != nil {
		return err
	}
	if status >= 400 {
		return decodeRESTError(status, method, path, data)
	}
	if out == nil {
		return nil
	}
	if len(data) == 0 {
		return fmt.Errorf("empty response from %s %s", method, path)
	}
	if err := json.Unmarshal(data, out); err != nil {
		return fmt.Errorf("error parsing response from %s %s: %w", method, path, err)
	}
	return nil
}

func (c *Client) rest(method, path string, query url.Values, body interface{}) ([]byte, int, error) {
	fullURL := c.restBaseURL() + path
	if len(query) > 0 {
		fullURL += "?" + query.Encode()
	}

	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, 0, fmt.Errorf("error marshaling request: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequest(method, fullURL, reqBody)
	if err != nil {
		return nil, 0, fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if err := c.applyAuthHeaders(req); err != nil {
		return nil, 0, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("error executing request: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("error reading response: %w", err)
	}
	return data, resp.StatusCode, nil
}

func decodeRESTError(status int, method, path string, data []byte) error {
	restErr := &RESTError{Status: status, Method: method, Path: path}
	var parsed struct {
		Error  string `json:"error"`
		Issues []struct {
			Path    []interface{} `json:"path"`
			Message string        `json:"message"`
		} `json:"issues"`
	}
	if err := json.Unmarshal(data, &parsed); err == nil && parsed.Error != "" {
		msg := parsed.Error
		if len(parsed.Issues) > 0 {
			details := make([]string, 0, len(parsed.Issues))
			for _, issue := range parsed.Issues {
				path := make([]string, 0, len(issue.Path))
				for _, p := range issue.Path {
					if s, ok := p.(string); ok {
						path = append(path, s)
					}
				}
				details = append(details, strings.Join(path, ".")+": "+issue.Message)
			}
			msg += " (" + strings.Join(details, "; ") + ")"
		}
		return fmt.Errorf("%w: %s", restErr, msg)
	}
	return restErr
}

// NowMillis is the default entitySequence source for record writes: the
// current Unix time in milliseconds, matching what Blue's own web client
// sends.
func NowMillis() int64 {
	return time.Now().UnixMilli()
}
