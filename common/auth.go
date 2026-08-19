package common

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/joho/godotenv"
)

// Default API URL
const DefaultAPIUrl = "https://api.blue.app/graphql"

// Config holds API configuration
type Config struct {
	APIUrl string
	// PAT credentials (`blue init`): sent as X-Bloo-Token-* headers.
	AuthToken string
	ClientID  string
	// OAuth credentials (`blue login`): the access token is sent as
	// Authorization: Bearer and refreshed transparently. Stored separately
	// from PAT fields so the two modes never overload one another.
	OAuthAccessToken  string
	OAuthRefreshToken string
	OAuthClientID     string
	CompanyID         string
	DefaultWorkspace  string
}

// GraphQLRequest represents a GraphQL request
type GraphQLRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables,omitempty"`
}

// GraphQLError represents a GraphQL error
type GraphQLError struct {
	Message    string                 `json:"message"`
	Extensions map[string]interface{} `json:"extensions,omitempty"`
}

// Client handles Blue API communication
type Client struct {
	config      *Config
	httpClient  *http.Client
	projectID   string
	projectSlug string
}

// ConfigDir returns the path to the Blue CLI config directory
func ConfigDir() string {
	// Check XDG_CONFIG_HOME first
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "blue")
	}
	// Fall back to ~/.config/blue
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "blue")
}

// ConfigPath returns the path to the Blue CLI config file
func ConfigPath() string {
	dir := ConfigDir()
	if dir == "" {
		return ""
	}
	return filepath.Join(dir, "config.env")
}

// LoadConfig loads configuration with the following priority:
// 1. Environment variables (always win)
// 2. .env file in current directory
// 3. ~/.config/blue/config.env (global config from "blue init")
func LoadConfig() (*Config, error) {
	// Try loading .env from current directory (silently ignore if not found)
	_ = godotenv.Load()

	// Try loading global config (silently ignore if not found)
	globalConfig := ConfigPath()
	if globalConfig != "" {
		_ = godotenv.Load(globalConfig)
	}

	config := &Config{
		APIUrl:            os.Getenv("API_URL"),
		AuthToken:         os.Getenv("AUTH_TOKEN"),
		ClientID:          os.Getenv("CLIENT_ID"),
		OAuthAccessToken:  os.Getenv("OAUTH_ACCESS_TOKEN"),
		OAuthRefreshToken: os.Getenv("OAUTH_REFRESH_TOKEN"),
		OAuthClientID:     os.Getenv("OAUTH_CLIENT_ID"),
		CompanyID:         os.Getenv("COMPANY_ID"),
		DefaultWorkspace:  os.Getenv("DEFAULT_WORKSPACE_ID"),
	}

	// Default API URL
	if config.APIUrl == "" {
		config.APIUrl = DefaultAPIUrl
	}

	// Two credential shapes:
	//   PAT   (`blue init`):   CLIENT_ID + AUTH_TOKEN
	//   OAuth (`blue login`):  OAUTH_ACCESS_TOKEN (+ OAUTH_REFRESH_TOKEN)
	hasPAT := config.ClientID != "" && config.AuthToken != ""
	hasOAuth := config.OAuthAccessToken != ""
	if !hasPAT && !hasOAuth {
		return nil, fmt.Errorf("not logged in. Run 'blue login' to sign in with your browser, or 'blue init' to use an API token")
	}

	// Company is command context, not a credential: OAuth logins may load
	// without one (--company selects per command); PAT configs keep today's
	// requirement so existing behavior is unchanged.
	if hasPAT && !hasOAuth && config.CompanyID == "" {
		return nil, fmt.Errorf("no company configured. Run 'blue init' again, or set COMPANY_ID / pass --company")
	}

	return config, nil
}

// NewClient creates a new Blue API client
func NewClient(config *Config) *Client {
	return &Client{
		config: config,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// oauthAccessToken returns a live OAuth access token, silently refreshing
// first when the stored JWT is expired (or unparsable). PAT configs return
// their secret unchanged — a PAT has no refresh token.
func (c *Client) oauthAccessToken() (string, error) {
	if c.config.OAuthAccessToken == "" {
		return "", nil
	}
	// Refresh 30s before actual expiry, so a request built now does not land
	// expired.
	if exp, ok := JWTExpiry(c.config.OAuthAccessToken); ok && time.Now().Add(30*time.Second).Before(exp) {
		return c.config.OAuthAccessToken, nil
	}
	if err := c.refreshOAuth(); err != nil {
		return "", err
	}
	return c.config.OAuthAccessToken, nil
}

// refreshOAuth redeems the stored refresh token and persists the rotated
// session (best-effort persist: a failed write only costs a refresh on the
// next command, never a failure of this one).
func (c *Client) refreshOAuth() error {
	if c.config.OAuthRefreshToken == "" {
		return fmt.Errorf("login session expired — run 'blue login' again")
	}
	tokens, err := RefreshAccessToken(OAuthBaseURL(c.config.APIUrl), c.config.OAuthClientID, c.config.OAuthRefreshToken)
	if err != nil {
		return fmt.Errorf("login session expired — run 'blue login' again: %w", err)
	}
	c.config.OAuthAccessToken = tokens.AccessToken
	if tokens.RefreshToken != "" {
		c.config.OAuthRefreshToken = tokens.RefreshToken
	}
	_ = SetOAuthTokens("", c.config.OAuthClientID, c.config.OAuthAccessToken, c.config.OAuthRefreshToken)
	return nil
}

// applyAuthHeaders sets the auth headers shared by every request. PAT configs
// send the X-Bloo-Token-* pair; OAuth configs send the access token as a
// Bearer token (refreshing it first if expired). The company header is only
// set when a company is configured — org-scoped commands resolve it via
// --company / COMPANY_ID, it is not a credential.
func (c *Client) applyAuthHeaders(req *http.Request) error {
	if c.config.OAuthAccessToken != "" {
		token, err := c.oauthAccessToken()
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+token)
	} else {
		req.Header.Set("X-Bloo-Token-ID", c.config.ClientID)
		req.Header.Set("X-Bloo-Token-Secret", c.config.AuthToken)
	}
	if c.config.CompanyID != "" {
		req.Header.Set("X-Bloo-Company-ID", c.config.CompanyID)
	}
	if c.projectID != "" {
		req.Header.Set("X-Bloo-Project-Id", c.projectID)
	} else if c.projectSlug != "" {
		req.Header.Set("X-Bloo-Project-Id", c.projectSlug)
	}
	return nil
}

// ExecuteQuery executes a GraphQL query and returns the raw response. OAuth
// sessions retry exactly once on a 401: the token is refreshed and the
// request resent, covering rotation races and clock skew without looping.
func (c *Client) ExecuteQuery(query string, variables map[string]interface{}) (map[string]interface{}, error) {
	data, status, err := c.executeQuery(query, variables)
	if status != http.StatusUnauthorized || c.config.OAuthAccessToken == "" {
		return data, err
	}
	if refreshErr := c.refreshOAuth(); refreshErr != nil {
		return nil, refreshErr
	}
	data, _, err = c.executeQuery(query, variables)
	return data, err
}

// executeQuery performs one GraphQL round trip and also returns the HTTP
// status (0 for failures before a response existed).
func (c *Client) executeQuery(query string, variables map[string]interface{}) (map[string]interface{}, int, error) {
	reqBody := GraphQLRequest{
		Query:     query,
		Variables: variables,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, 0, fmt.Errorf("error marshaling request: %w", err)
	}

	req, err := http.NewRequest("POST", c.config.APIUrl, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, 0, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if err := c.applyAuthHeaders(req); err != nil {
		return nil, 0, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("error executing request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("error reading response: %w", err)
	}

	var response struct {
		Data   map[string]interface{} `json:"data"`
		Errors []GraphQLError         `json:"errors"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, resp.StatusCode, fmt.Errorf("error parsing response: %w", err)
	}

	if len(response.Errors) > 0 {
		return nil, resp.StatusCode, fmt.Errorf("GraphQL error: %s", response.Errors[0].Message)
	}

	if response.Data == nil {
		return nil, resp.StatusCode, fmt.Errorf("no data in response")
	}

	return response.Data, resp.StatusCode, nil
}

// ExecuteQueryWithResult executes a GraphQL query and unmarshals the result
func (c *Client) ExecuteQueryWithResult(query string, variables map[string]interface{}, result interface{}) error {
	data, err := c.ExecuteQuery(query, variables)
	if err != nil {
		return err
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("error marshaling data: %w", err)
	}

	if err := json.Unmarshal(jsonData, result); err != nil {
		return fmt.Errorf("error unmarshaling result: %w", err)
	}

	return nil
}

// SetProjectID sets the project ID for requests that require project context
func (c *Client) SetProjectID(projectID string) {
	c.projectID = projectID
	c.projectSlug = ""
}

// SetProjectSlug sets the project slug for requests that require project context
func (c *Client) SetProjectSlug(projectSlug string) {
	c.projectSlug = projectSlug
	c.projectID = ""
}

// SetProject sets the project ID or slug for requests that require project context
func (c *Client) SetProject(project string) {
	if len(project) > 20 && (project[8] == '-' || project[13] == '-' || project[18] == '-') {
		c.SetProjectID(project)
	} else {
		c.SetProjectSlug(project)
	}
}

// GetProjectID returns the current project ID
func (c *Client) GetProjectID() string {
	return c.projectID
}

// GetProjectSlug returns the current project slug
func (c *Client) GetProjectSlug() string {
	return c.projectSlug
}

// GetProjectContext returns the current project context (ID or slug)
func (c *Client) GetProjectContext() string {
	if c.projectID != "" {
		return c.projectID
	}
	return c.projectSlug
}

// GetCompanyID returns the configured company ID (slug)
func (c *Client) GetCompanyID() string {
	return c.config.CompanyID
}

// ResolveCompanyID resolves the company slug to the actual company ID (cuid).
// Some GraphQL queries (like dashboards) require the real ID, not the slug.
func (c *Client) ResolveCompanyID() (string, error) {
	query := fmt.Sprintf(`query { company(id: "%s") { id } }`, c.config.CompanyID)
	data, err := c.ExecuteQuery(query, nil)
	if err != nil {
		return "", fmt.Errorf("failed to resolve company ID: %w", err)
	}
	if company, ok := data["company"].(map[string]interface{}); ok {
		if id, ok := company["id"].(string); ok {
			return id, nil
		}
	}
	return "", fmt.Errorf("could not resolve company ID from slug '%s'", c.config.CompanyID)
}

// ResolveProjectID resolves a workspace ID or slug to the actual project ID.
func (c *Client) ResolveProjectID(project string) (string, error) {
	query := `query ResolveProject($project: String!) {
		project(id: $project) {
			id
		}
	}`

	variables := map[string]interface{}{
		"project": project,
	}

	var response struct {
		Project struct {
			ID string `json:"id"`
		} `json:"project"`
	}

	if err := c.ExecuteQueryWithResult(query, variables, &response); err != nil {
		return "", err
	}

	if response.Project.ID == "" {
		return "", fmt.Errorf("could not resolve workspace ID from %q", project)
	}

	return response.Project.ID, nil
}

// DownloadFile downloads a file from the given URL using the authenticated client
func (c *Client) DownloadFile(url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	if err := c.applyAuthHeaders(req); err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %w", err)
	}

	return data, nil
}
