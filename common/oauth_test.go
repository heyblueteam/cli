package common

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"
	"time"
)

func TestPKCEChallengeRFC7636Vector(t *testing.T) {
	// RFC 7636 appendix B test vector.
	verifier := "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"
	want := "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM"
	if got := PKCEChallenge(verifier); got != want {
		t.Errorf("PKCEChallenge = %q, want %q", got, want)
	}
}

func TestCreatePKCEBounds(t *testing.T) {
	verifier, challenge, err := CreatePKCE()
	if err != nil {
		t.Fatalf("CreatePKCE: %v", err)
	}
	if len(verifier) < 43 || len(verifier) > 128 {
		t.Errorf("verifier length %d outside RFC 7636 bounds 43-128", len(verifier))
	}
	if challenge != PKCEChallenge(verifier) {
		t.Error("challenge is not the S256 digest of the verifier")
	}
	// Two runs must not repeat.
	verifier2, _, _ := CreatePKCE()
	if verifier == verifier2 {
		t.Error("verifier repeated across runs")
	}
}

func TestCreateStateUnpredictable(t *testing.T) {
	a, err := CreateState()
	if err != nil {
		t.Fatalf("CreateState: %v", err)
	}
	b, _ := CreateState()
	if a == "" || a == b {
		t.Error("state is empty or repeated")
	}
}

func TestOAuthBaseURL(t *testing.T) {
	cases := map[string]string{
		"https://api.blue.app/graphql":  "https://api.blue.app",
		"https://api.blue.app/graphql/": "https://api.blue.app",
		"https://api.blue.app":          "https://api.blue.app",
		"http://localhost:3000/graphql": "http://localhost:3000",
	}
	for in, want := range cases {
		if got := OAuthBaseURL(in); got != want {
			t.Errorf("OAuthBaseURL(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestRegisterOAuthClient(t *testing.T) {
	var gotName string
	var gotUris []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/oauth2/register" {
			t.Errorf("path = %q, want /oauth2/register", r.URL.Path)
		}
		var body struct {
			ClientName   string   `json:"client_name"`
			RedirectUris []string `json:"redirect_uris"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		gotName, gotUris = body.ClientName, body.RedirectUris
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]string{"client_id": "dcrid"})
	}))
	defer server.Close()

	clientID, err := RegisterOAuthClient(server.URL, "Blue CLI", "http://127.0.0.1:53712/callback")
	if err != nil {
		t.Fatalf("RegisterOAuthClient: %v", err)
	}
	if clientID != "dcrid" {
		t.Errorf("client_id = %q", clientID)
	}
	if gotName != "Blue CLI" {
		t.Errorf("client_name = %q", gotName)
	}
	// The registered URI must carry the exact bound port — the server
	// exact-matches redirect URIs.
	if len(gotUris) != 1 || gotUris[0] != "http://127.0.0.1:53712/callback" {
		t.Errorf("redirect_uris = %v", gotUris)
	}
}

func TestBuildAuthorizeURL(t *testing.T) {
	raw := BuildAuthorizeURL("https://api.blue.app", "dcrid", "http://127.0.0.1:53712/callback", "chal", "st")
	u, err := url.Parse(raw)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if u.Path != "/oauth2/authorize" {
		t.Errorf("path = %q, want /oauth2/authorize", u.Path)
	}
	q := u.Query()
	want := map[string]string{
		"response_type":         "code",
		"client_id":             "dcrid",
		"redirect_uri":          "http://127.0.0.1:53712/callback",
		"code_challenge":        "chal",
		"code_challenge_method": "S256",
		"state":                 "st",
	}
	for key, value := range want {
		if q.Get(key) != value {
			t.Errorf("%s = %q, want %q", key, q.Get(key), value)
		}
	}
}

func TestJWTExpiry(t *testing.T) {
	exp := time.Now().Add(time.Hour).Unix()
	payload, _ := json.Marshal(map[string]int64{"exp": exp})
	token := fmt.Sprintf("eyJhbGciOiJIUzI1NiJ9.%s.c2lnbmF0dXJl", base64.RawURLEncoding.EncodeToString(payload))
	got, ok := JWTExpiry(token)
	if !ok {
		t.Fatal("JWTExpiry not ok for a well-formed JWT")
	}
	if got.Unix() != exp {
		t.Errorf("exp = %d, want %d", got.Unix(), exp)
	}

	for _, bad := range []string{"", "a.b", "not-a-jwt", "a.b.c.d"} {
		if _, ok := JWTExpiry(bad); ok {
			t.Errorf("JWTExpiry(%q) unexpectedly ok", bad)
		}
	}
}

func TestExchangeAuthorizationCodeOauthError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid_grant"})
	}))
	defer server.Close()

	_, err := ExchangeAuthorizationCode(server.URL, "dcrid", "code", "verifier", "http://127.0.0.1:9/callback")
	if err == nil {
		t.Fatal("expected an error")
	}
	if got := err.Error(); got != "oauth error: invalid_grant ()" {
		t.Errorf("error = %q", got)
	}
}

func TestRefreshAccessTokenOmitsRefreshTokenWhenServerDoes(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Racing-client response: access token only, no refresh token.
		_ = json.NewEncoder(w).Encode(OAuthTokens{AccessToken: "fresh"})
	}))
	defer server.Close()

	tokens, err := RefreshAccessToken(server.URL, "dcrid", "rt")
	if err != nil {
		t.Fatalf("RefreshAccessToken: %v", err)
	}
	if tokens.AccessToken != "fresh" || tokens.RefreshToken != "" {
		t.Errorf("tokens = %+v", tokens)
	}
}

func TestRevokeToken(t *testing.T) {
	var received string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/oauth2/revoke" {
			t.Errorf("path = %q", r.URL.Path)
		}
		var body map[string]string
		_ = json.NewDecoder(r.Body).Decode(&body)
		received = body["token"]
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	if err := RevokeToken(server.URL, "tok"); err != nil {
		t.Fatalf("RevokeToken: %v", err)
	}
	if received != "tok" {
		t.Errorf("revoked %q, want tok", received)
	}

	failing := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer failing.Close()
	if err := RevokeToken(failing.URL, "tok"); err == nil {
		t.Error("expected an error on non-200")
	}
}

func TestOAuthTokenPersistence(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	// A PAT config that login must NOT touch.
	values := map[string]string{
		"API_URL":              "https://api.blue.app/graphql",
		"AUTH_TOKEN":           "pat-secret",
		"CLIENT_ID":            "pat-id",
		"COMPANY_ID":           "acme",
		"DEFAULT_WORKSPACE_ID": "ws_1",
	}
	if err := WriteConfigFile(values); err != nil {
		t.Fatalf("WriteConfigFile: %v", err)
	}

	if err := SetOAuthTokens("", "dcrid", "jwt", "refresh"); err != nil {
		t.Fatalf("SetOAuthTokens: %v", err)
	}
	got, err := ReadConfigFile()
	if err != nil {
		t.Fatalf("ReadConfigFile: %v", err)
	}
	for key, want := range map[string]string{
		"OAUTH_CLIENT_ID":      "dcrid",
		"OAUTH_ACCESS_TOKEN":   "jwt",
		"OAUTH_REFRESH_TOKEN":  "refresh",
		"AUTH_TOKEN":           "pat-secret",
		"CLIENT_ID":            "pat-id",
		"COMPANY_ID":           "acme",
		"DEFAULT_WORKSPACE_ID": "ws_1",
	} {
		if got[key] != want {
			t.Errorf("%s = %q, want %q", key, got[key], want)
		}
	}

	if err := ClearOAuthTokens(); err != nil {
		t.Fatalf("ClearOAuthTokens: %v", err)
	}
	got, _ = ReadConfigFile()
	for _, key := range []string{"OAUTH_CLIENT_ID", "OAUTH_ACCESS_TOKEN", "OAUTH_REFRESH_TOKEN"} {
		if got[key] != "" {
			t.Errorf("ClearOAuthTokens left %s", key)
		}
	}
	// PAT fields survive logout.
	if got["AUTH_TOKEN"] != "pat-secret" || got["CLIENT_ID"] != "pat-id" || got["COMPANY_ID"] != "acme" {
		t.Errorf("ClearOAuthTokens must keep PAT/company fields, got %v", got)
	}
}

// Concurrent refreshes must not corrupt the config file: the write lock
// serializes them, and the final state is one consistent generation.
func TestSetOAuthTokensConcurrent(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	if err := WriteConfigFile(map[string]string{"COMPANY_ID": "acme"}); err != nil {
		t.Fatalf("WriteConfigFile: %v", err)
	}

	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			_ = SetOAuthTokens("", "dcrid", fmt.Sprintf("token-%d", n), fmt.Sprintf("refresh-%d", n))
		}(i)
	}
	wg.Wait()

	got, err := ReadConfigFile()
	if err != nil {
		t.Fatalf("ReadConfigFile: %v", err)
	}
	if got["COMPANY_ID"] != "acme" {
		t.Errorf("COMPANY_ID lost: %v", got)
	}
	if got["OAUTH_ACCESS_TOKEN"] == "" || got["OAUTH_REFRESH_TOKEN"] == "" {
		t.Errorf("OAuth fields missing after concurrent writes: %v", got)
	}
}

func TestLoadConfigAcceptsOAuthSessionWithoutCompany(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	t.Setenv("API_URL", "https://api.blue.app/graphql")
	t.Setenv("OAUTH_ACCESS_TOKEN", "jwt")
	t.Setenv("OAUTH_REFRESH_TOKEN", "refresh")
	t.Setenv("OAUTH_CLIENT_ID", "dcrid")
	for _, key := range []string{"AUTH_TOKEN", "CLIENT_ID", "COMPANY_ID", "DEFAULT_WORKSPACE_ID"} {
		t.Setenv(key, "")
	}

	config, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	if config.OAuthAccessToken != "jwt" || config.OAuthRefreshToken != "refresh" {
		t.Errorf("OAuth fields = %q / %q", config.OAuthAccessToken, config.OAuthRefreshToken)
	}
	if config.CompanyID != "" {
		t.Errorf("CompanyID = %q, want empty (company is command context, not a credential)", config.CompanyID)
	}
}

func TestLoadConfigKeepsPATCompanyRequirement(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("API_URL", "https://api.blue.app/graphql")
	t.Setenv("AUTH_TOKEN", "pat-secret")
	t.Setenv("CLIENT_ID", "pat-id")
	t.Setenv("COMPANY_ID", "")
	for _, key := range []string{"OAUTH_ACCESS_TOKEN", "OAUTH_REFRESH_TOKEN", "OAUTH_CLIENT_ID"} {
		t.Setenv(key, "")
	}
	if _, err := LoadConfig(); err == nil {
		t.Error("PAT config without COMPANY_ID must keep failing as it always has")
	}
}

func TestLoadConfigRejectsEmpty(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	for _, key := range []string{"API_URL", "AUTH_TOKEN", "CLIENT_ID", "COMPANY_ID", "DEFAULT_WORKSPACE_ID", "OAUTH_ACCESS_TOKEN", "OAUTH_REFRESH_TOKEN", "OAUTH_CLIENT_ID"} {
		t.Setenv(key, "")
	}
	if _, err := LoadConfig(); err == nil {
		t.Fatal("expected 'not logged in' error for an empty config")
	}
}

func TestClientSendsBearerForOAuth(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	var gotAuth, gotCompany, gotPatID string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotCompany = r.Header.Get("X-Bloo-Company-ID")
		gotPatID = r.Header.Get("X-Bloo-Token-ID")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"data": map[string]interface{}{"ok": true}})
	}))
	defer server.Close()

	config := &Config{
		APIUrl:            server.URL + "/graphql",
		OAuthAccessToken:  makeJWT(time.Now().Add(time.Hour)),
		OAuthRefreshToken: "refresh",
		OAuthClientID:     "dcrid",
		CompanyID:         "acme",
	}
	if _, err := NewClient(config).ExecuteQuery("{ ok }", nil); err != nil {
		t.Fatalf("ExecuteQuery: %v", err)
	}
	if gotAuth != "Bearer "+config.OAuthAccessToken {
		t.Errorf("Authorization = %q", gotAuth)
	}
	if gotPatID != "" {
		t.Error("OAuth mode must not send PAT headers")
	}
	if gotCompany != "acme" {
		t.Errorf("X-Bloo-Company-ID = %q", gotCompany)
	}
}

func TestClientOmitsCompanyHeaderWhenUnset(t *testing.T) {
	var gotCompany, gotAuth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotCompany = r.Header.Get("X-Bloo-Company-ID")
		gotAuth = r.Header.Get("Authorization")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"data": map[string]interface{}{"ok": true}})
	}))
	defer server.Close()

	config := &Config{
		APIUrl:           server.URL + "/graphql",
		OAuthAccessToken: makeJWT(time.Now().Add(time.Hour)),
	}
	if _, err := NewClient(config).ExecuteQuery("{ ok }", nil); err != nil {
		t.Fatalf("ExecuteQuery: %v", err)
	}
	if gotCompany != "" {
		t.Errorf("X-Bloo-Company-ID = %q, want unset", gotCompany)
	}
	if gotAuth == "" {
		t.Error("Authorization missing")
	}
}

func TestClientSilentlyRefreshesExpiredToken(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	var refreshCalls int
	var sawToken string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oauth2/token":
			refreshCalls++
			_ = json.NewEncoder(w).Encode(OAuthTokens{AccessToken: "fresh-jwt", RefreshToken: "next-refresh"})
		case "/graphql":
			sawToken = r.Header.Get("Authorization")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"data": map[string]interface{}{"ok": true}})
		default:
			t.Errorf("unexpected path %q", r.URL.Path)
		}
	}))
	defer server.Close()

	config := &Config{
		APIUrl:            server.URL + "/graphql",
		OAuthAccessToken:  makeJWT(time.Now().Add(-time.Minute)), // already expired
		OAuthRefreshToken: "stale-refresh",
		OAuthClientID:     "dcrid",
		CompanyID:         "acme",
	}
	if _, err := NewClient(config).ExecuteQuery("{ ok }", nil); err != nil {
		t.Fatalf("ExecuteQuery: %v", err)
	}
	if refreshCalls != 1 {
		t.Fatalf("refresh calls = %d, want 1", refreshCalls)
	}
	if sawToken != "Bearer fresh-jwt" {
		t.Errorf("Authorization = %q, want the refreshed token", sawToken)
	}
	// The rotated session is persisted for the next command.
	saved, err := ReadConfigFile()
	if err != nil {
		t.Fatalf("ReadConfigFile: %v", err)
	}
	if saved["OAUTH_ACCESS_TOKEN"] != "fresh-jwt" || saved["OAUTH_REFRESH_TOKEN"] != "next-refresh" {
		t.Errorf("persisted session = %v", saved)
	}
}

func TestClientRetriesOnceOn401(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	var graphQLCalls int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oauth2/token":
			_ = json.NewEncoder(w).Encode(OAuthTokens{AccessToken: "fresh-jwt", RefreshToken: "next-refresh"})
		case "/graphql":
			graphQLCalls++
			if graphQLCalls == 1 {
				// First call 401s (e.g. the token was rotated elsewhere).
				w.WriteHeader(http.StatusUnauthorized)
				_ = json.NewEncoder(w).Encode(map[string]interface{}{"errors": []map[string]string{{"message": "Unauthorized"}}})
				return
			}
			if r.Header.Get("Authorization") != "Bearer fresh-jwt" {
				t.Errorf("retry did not use the refreshed token: %q", r.Header.Get("Authorization"))
			}
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"data": map[string]interface{}{"ok": true}})
		default:
			t.Errorf("unexpected path %q", r.URL.Path)
		}
	}))
	defer server.Close()

	// Unexpired token, so no pre-flight refresh: the 401 path is what retries.
	config := &Config{
		APIUrl:            server.URL + "/graphql",
		OAuthAccessToken:  makeJWT(time.Now().Add(time.Hour)),
		OAuthRefreshToken: "refresh",
		OAuthClientID:     "dcrid",
		CompanyID:         "acme",
	}
	if _, err := NewClient(config).ExecuteQuery("{ ok }", nil); err != nil {
		t.Fatalf("ExecuteQuery: %v", err)
	}
	if graphQLCalls != 2 {
		t.Fatalf("graphql calls = %d, want 2 (one failed, one retried)", graphQLCalls)
	}
}

func TestClientSendsPATHeadersWhenClientIDSet(t *testing.T) {
	var gotID, gotSecret, gotAuth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotID = r.Header.Get("X-Bloo-Token-ID")
		gotSecret = r.Header.Get("X-Bloo-Token-Secret")
		gotAuth = r.Header.Get("Authorization")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"data": map[string]interface{}{"ok": true}})
	}))
	defer server.Close()

	config := &Config{APIUrl: server.URL + "/graphql", AuthToken: "pat-secret", ClientID: "pat-id", CompanyID: "acme"}
	if _, err := NewClient(config).ExecuteQuery("{ ok }", nil); err != nil {
		t.Fatalf("ExecuteQuery: %v", err)
	}
	if gotID != "pat-id" || gotSecret != "pat-secret" {
		t.Errorf("PAT headers = %q / %q", gotID, gotSecret)
	}
	if gotAuth != "" {
		t.Errorf("Authorization unexpectedly set for a PAT config: %q", gotAuth)
	}
}

// makeJWT builds an unsigned JWT with only an exp claim — JWTExpiry does not
// verify signatures, and the server-side tests cover real tokens.
func makeJWT(expiry time.Time) string {
	payload, _ := json.Marshal(map[string]int64{"exp": expiry.Unix()})
	return "h." + base64.RawURLEncoding.EncodeToString(payload) + ".s"
}
