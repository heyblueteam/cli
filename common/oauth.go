package common

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// oauthHTTPClient is shared by every OAuth endpoint call (registration, token
// exchange, refresh, revoke). Short timeout: these are small JSON round trips.
var oauthHTTPClient = &http.Client{Timeout: 30 * time.Second}

// OAuthBaseURL derives the API origin from a GraphQL API URL
// (e.g. "https://api.blue.app/graphql" → "https://api.blue.app").
func OAuthBaseURL(apiURL string) string {
	return strings.TrimSuffix(strings.TrimRight(apiURL, "/"), "/graphql")
}

// OAuthTokens is the token-endpoint response shared by the authorization-code
// and refresh-token grants. RefreshToken is omitted by the server when a
// racing refresh won — keep the one already held in that case.
type OAuthTokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

// CreatePKCE generates a PKCE S256 pair (RFC 7636): a random 86-char
// base64url verifier and its base64url(sha256) challenge.
func CreatePKCE() (verifier string, challenge string, err error) {
	raw := make([]byte, 64)
	if _, err = rand.Read(raw); err != nil {
		return "", "", fmt.Errorf("failed to generate PKCE verifier: %w", err)
	}
	verifier = base64.RawURLEncoding.EncodeToString(raw)
	challenge = PKCEChallenge(verifier)
	return verifier, challenge, nil
}

// PKCEChallenge returns the S256 challenge for a verifier.
func PKCEChallenge(verifier string) string {
	digest := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(digest[:])
}

// CreateState generates a random OAuth state value.
func CreateState() (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", fmt.Errorf("failed to generate state: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}

// RegisterOAuthClient dynamically registers a public OAuth client (RFC 7591)
// for this login, exactly like MCP connector clients do. The redirect URI is
// the loopback callback with the ACTUAL bound port — the server exact-matches
// it, so registration must happen after the listener is up.
func RegisterOAuthClient(baseURL, clientName, redirectURI string) (string, error) {
	body, err := json.Marshal(map[string]interface{}{
		"client_name": clientName,
		"redirect_uris": []string{
			redirectURI,
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to build request: %w", err)
	}
	req, err := http.NewRequest("POST", baseURL+"/oauth2/register", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := oauthHTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("client registration failed: %w", err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}
	if resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("client registration rejected (status %d): %s", resp.StatusCode, strings.TrimSpace(string(raw)))
	}
	var registered struct {
		ClientID string `json:"client_id"`
	}
	if err := json.Unmarshal(raw, &registered); err != nil || registered.ClientID == "" {
		return "", fmt.Errorf("client registration response carried no client_id")
	}
	return registered.ClientID, nil
}

// BuildAuthorizeURL builds the /oauth2/authorize URL the user's browser opens.
func BuildAuthorizeURL(baseURL, clientID, redirectURI, codeChallenge, state string) string {
	q := url.Values{}
	q.Set("response_type", "code")
	q.Set("client_id", clientID)
	q.Set("redirect_uri", redirectURI)
	q.Set("code_challenge", codeChallenge)
	q.Set("code_challenge_method", "S256")
	q.Set("state", state)
	return baseURL + "/oauth2/authorize?" + q.Encode()
}

func postOAuth(baseURL, path string, payload map[string]string) (*OAuthTokens, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to build request: %w", err)
	}
	req, err := http.NewRequest("POST", baseURL+path, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := oauthHTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		var oauthErr struct {
			Error       string `json:"error"`
			Description string `json:"error_description"`
		}
		_ = json.Unmarshal(raw, &oauthErr)
		if oauthErr.Error != "" {
			return nil, fmt.Errorf("oauth error: %s (%s)", oauthErr.Error, oauthErr.Description)
		}
		return nil, fmt.Errorf("unexpected status %d from %s", resp.StatusCode, path)
	}

	var tokens OAuthTokens
	if err := json.Unmarshal(raw, &tokens); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	if tokens.AccessToken == "" {
		return nil, fmt.Errorf("empty access token from %s", path)
	}
	return &tokens, nil
}

// ExchangeAuthorizationCode redeems the consent code for tokens (PKCE S256).
func ExchangeAuthorizationCode(baseURL, clientID, code, codeVerifier, redirectURI string) (*OAuthTokens, error) {
	return postOAuth(baseURL, "/oauth2/token", map[string]string{
		"grant_type":    "authorization_code",
		"code":          code,
		"code_verifier": codeVerifier,
		"redirect_uri":  redirectURI,
		"client_id":     clientID,
	})
}

// RefreshAccessToken rotates the refresh token for a fresh access token.
func RefreshAccessToken(baseURL, clientID, refreshToken string) (*OAuthTokens, error) {
	return postOAuth(baseURL, "/oauth2/token", map[string]string{
		"grant_type":    "refresh_token",
		"refresh_token": refreshToken,
		"client_id":     clientID,
	})
}

// RevokeToken revokes the grant server-side (RFC 7009). Always returns nil on
// HTTP 200; a network or server failure is returned as an error.
func RevokeToken(baseURL, token string) error {
	body, err := json.Marshal(map[string]string{"token": token})
	if err != nil {
		return fmt.Errorf("failed to build request: %w", err)
	}
	req, err := http.NewRequest("POST", baseURL+"/oauth2/revoke", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := oauthHTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status %d from /oauth2/revoke", resp.StatusCode)
	}
	return nil
}

// JWTExpiry reads the `exp` claim of a JWT WITHOUT verifying the signature.
// The signature is checked by the server — this only decides locally whether
// to refresh before sending, so tampering can only force an unnecessary
// refresh, never gain access.
func JWTExpiry(token string) (time.Time, bool) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return time.Time{}, false
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return time.Time{}, false
	}
	var claims struct {
		Exp int64 `json:"exp"`
	}
	if err := json.Unmarshal(payload, &claims); err != nil || claims.Exp == 0 {
		return time.Time{}, false
	}
	return time.Unix(claims.Exp, 0), true
}

// oauthConfigLock serializes config-file writes: two concurrent commands
// refreshing at once must not interleave writes and lose the newest token.
var oauthConfigLock sync.Mutex

// SetOAuthTokens persists the browser-login session to the global config
// file under dedicated OAUTH_* keys. PAT fields (AUTH_TOKEN / CLIENT_ID),
// company, and workspace settings are left untouched, so the two credential
// kinds never overload one another and logout can clear just this session.
// apiURL is persisted only when non-empty (a --api-url login), so later
// refreshes and logout talk to the same API the login used.
func SetOAuthTokens(apiURL, clientID, accessToken, refreshToken string) error {
	oauthConfigLock.Lock()
	defer oauthConfigLock.Unlock()
	values, err := ReadConfigFile()
	if err != nil {
		values = map[string]string{}
	}
	if apiURL != "" {
		values["API_URL"] = apiURL
	}
	values["OAUTH_CLIENT_ID"] = clientID
	values["OAUTH_ACCESS_TOKEN"] = accessToken
	if refreshToken != "" {
		values["OAUTH_REFRESH_TOKEN"] = refreshToken
	} else {
		delete(values, "OAUTH_REFRESH_TOKEN")
	}
	return WriteConfigFile(values)
}

// ClearOAuthTokens removes only the browser-login session from the global
// config file; PAT credentials and every other key stay as they are.
func ClearOAuthTokens() error {
	oauthConfigLock.Lock()
	defer oauthConfigLock.Unlock()
	values, err := ReadConfigFile()
	if err != nil {
		return err
	}
	delete(values, "OAUTH_CLIENT_ID")
	delete(values, "OAUTH_ACCESS_TOKEN")
	delete(values, "OAUTH_REFRESH_TOKEN")
	return WriteConfigFile(values)
}
