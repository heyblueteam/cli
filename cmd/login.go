package cmd

import (
	"context"
	"fmt"
	"html"
	"net"
	"net/http"
	"os/exec"
	"runtime"
	"time"

	"github.com/heyblueteam/cli/common"

	"github.com/spf13/cobra"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Sign in to Blue with your browser",
	Long: `Sign in to the Blue CLI with your browser.

Opens a browser to the Blue consent screen; after you approve, the CLI
receives its token automatically — no Client ID / Auth Token copy-paste.
The connection appears under Account -> Security -> Connected apps as
"Blue CLI" and can be revoked there or with 'blue logout'.

For scripts, CI, and agents without a browser, keep using 'blue init'.`,
	Example: `  blue login`,
	RunE:    runLogin,
}

var loginAPIURL string

// loginTimeout bounds the whole wait-for-browser round trip.
const loginTimeout = 2 * time.Minute

func init() {
	rootCmd.AddCommand(loginCmd)
	loginCmd.Flags().StringVar(&loginAPIURL, "api-url", "", "API URL (default: "+common.DefaultAPIUrl+")")
}

type callbackResult struct {
	code  string
	state string
	err   string
}

func runLogin(cmd *cobra.Command, args []string) error {
	apiURL := loginAPIURL
	if apiURL == "" {
		if existing, err := common.ReadConfigFile(); err == nil && existing["API_URL"] != "" {
			apiURL = existing["API_URL"]
		}
	}
	if apiURL == "" {
		apiURL = common.DefaultAPIUrl
	}
	baseURL := common.OAuthBaseURL(apiURL)

	verifier, challenge, err := common.CreatePKCE()
	if err != nil {
		return err
	}
	state, err := common.CreateState()
	if err != nil {
		return err
	}

	// Bind the loopback callback FIRST (RFC 8252) — the OS assigns the port,
	// and the client registered below must carry the exact URI, port included,
	// because the authorization server exact-matches redirect URIs.
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return fmt.Errorf("could not bind a local callback port: %w", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	redirectURI := fmt.Sprintf("http://127.0.0.1:%d/callback", port)

	// Dynamic client registration (RFC 7591) — the same path MCP connector
	// clients use. Public client, no secret; PKCE is the proof of possession.
	clientID, err := common.RegisterOAuthClient(baseURL, "Blue CLI", redirectURI)
	if err != nil {
		return err
	}

	authorizeURL := common.BuildAuthorizeURL(baseURL, clientID, redirectURI, challenge, state)

	results := make(chan callbackResult, 1)
	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		result := callbackResult{code: q.Get("code"), state: q.Get("state"), err: q.Get("error")}
		results <- result

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if result.err != "" || result.state != state || result.code == "" {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "<h1>Sign-in failed</h1><p>%s</p><p>Return to your terminal and run <code>blue login</code> again.</p>", html.EscapeString(result.err))
			return
		}
		fmt.Fprint(w, "<h1>Signed in to Blue</h1><p>You can close this tab and return to your terminal.</p>")
	})
	server := &http.Server{Handler: mux}
	go func() { _ = server.Serve(listener) }()
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = server.Shutdown(ctx)
	}()

	fmt.Println("Opening Blue sign-in in your browser…")
	fmt.Println()
	fmt.Println("  " + authorizeURL)
	fmt.Println()
	if err := openBrowser(authorizeURL); err != nil {
		fmt.Println("Could not open a browser automatically.")
		fmt.Println("Open the URL above manually — waiting for the sign-in to complete…")
	}

	var callback callbackResult
	select {
	case callback = <-results:
	case <-time.After(loginTimeout):
		return fmt.Errorf("no response from the browser within %s — check the browser tab, then run 'blue login' again", loginTimeout)
	}
	if callback.err != "" {
		return fmt.Errorf("sign-in was rejected: %s", callback.err)
	}
	if callback.state != state {
		return fmt.Errorf("sign-in state mismatch — this can happen with two 'blue login' runs at once; run 'blue login' again")
	}
	if callback.code == "" {
		return fmt.Errorf("the sign-in response carried no authorization code — run 'blue login' again")
	}

	tokens, err := common.ExchangeAuthorizationCode(baseURL, clientID, callback.code, verifier, redirectURI)
	if err != nil {
		return fmt.Errorf("could not complete sign-in: %w", err)
	}

	// Persist the session under its own OAUTH_* keys. PAT credentials,
	// company, and workspace settings are untouched — a browser login replaces
	// nothing, it is added; 'blue logout' removes exactly these keys.
	if err := common.SetOAuthTokens(apiURL, clientID, tokens.AccessToken, tokens.RefreshToken); err != nil {
		return fmt.Errorf("signed in, but could not save the session: %w", err)
	}

	// Confirm what the token reaches. Best-effort: a failure here does not
	// undo a successful login.
	config, err := common.LoadConfig()
	fmt.Println()
	if err != nil {
		fmt.Println("Logged in. Try: blue whoami")
		return nil
	}
	client := common.NewClient(config)
	query := `query LoginWhoami { currentUser { email } companies(take: 20) { items { name slug } } }`
	var response struct {
		CurrentUser struct {
			Email string `json:"email"`
		} `json:"currentUser"`
		Companies struct {
			Items []struct {
				Name string `json:"name"`
				Slug string `json:"slug"`
			} `json:"items"`
		} `json:"companies"`
	}
	if err := client.ExecuteQueryWithResult(query, nil, &response); err == nil && response.CurrentUser.Email != "" {
		fmt.Printf("Logged in as %s\n", response.CurrentUser.Email)
		if len(response.Companies.Items) > 0 {
			fmt.Print("Your companies: ")
			for i, c := range response.Companies.Items {
				if i > 0 {
					fmt.Print(", ")
				}
				fmt.Printf("%s (%s)", c.Slug, c.Name)
				_ = common.AddCompany(c.Slug)
			}
			fmt.Println()
			if config.CompanyID == "" {
				fmt.Println("Pick a default company with 'blue company use <slug>' or pass --company per command.")
			}
		}
	} else {
		fmt.Println("Logged in (could not verify the connection — check with: blue whoami)")
	}
	fmt.Println("The connection shows under Account -> Security -> Connected apps as \"Blue CLI\".")
	fmt.Println("Try: blue whoami")
	return nil
}

// openBrowser opens the system browser at url. Returns an error when no
// browser opener exists (headless boxes, SSH sessions) — the caller prints
// the URL for manual opening instead of failing.
func openBrowser(url string) error {
	var command *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		command = exec.Command("open", url)
	case "windows":
		command = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		command = exec.Command("xdg-open", url)
	}
	return command.Start()
}
