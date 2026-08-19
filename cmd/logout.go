package cmd

import (
	"fmt"

	"github.com/heyblueteam/cli/common"

	"github.com/spf13/cobra"
)

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Sign out of the Blue CLI browser session",
	Long: `Revoke the 'blue login' session server-side and remove it locally.

Only touches the browser-login credentials. A personal access token set up
with 'blue init' is untouched (revoke that from Account Settings > API in
Blue).`,
	Example: `  blue logout`,
	RunE:    runLogout,
}

func init() {
	rootCmd.AddCommand(logoutCmd)
}

func runLogout(cmd *cobra.Command, args []string) error {
	values, err := common.ReadConfigFile()
	if err != nil || values["OAUTH_ACCESS_TOKEN"] == "" {
		fmt.Println("No browser login session found — nothing to log out.")
		return nil
	}

	// Revoke server-side first: both token kinds work (RFC 7009). A failed
	// revoke still clears locally — the user asked to be signed out, and the
	// grant can also be revoked from the Connected Apps settings page.
	baseURL := common.OAuthBaseURL(values["API_URL"])
	if baseURL == "" {
		baseURL = common.OAuthBaseURL(common.DefaultAPIUrl)
	}
	token := values["OAUTH_REFRESH_TOKEN"]
	if token == "" {
		token = values["OAUTH_ACCESS_TOKEN"]
	}
	if err := common.RevokeToken(baseURL, token); err != nil {
		fmt.Printf("Warning: could not revoke the session on the server (%v).\n", err)
		fmt.Println("You can also revoke \"Blue CLI\" under Account -> Security -> Connected apps.")
	}

	if err := common.ClearOAuthTokens(); err != nil {
		return fmt.Errorf("signed out server-side, but could not clear the local session: %w", err)
	}

	fmt.Println("Logged out. Run 'blue login' to sign in again.")
	return nil
}
