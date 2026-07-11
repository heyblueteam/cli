package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/heyblueteam/cli/common"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Set up Blue CLI credentials",
	Long: `Configure the Blue CLI with your API credentials.

Credentials are saved to ~/.config/blue/config.env and used
for all future commands. You can re-run this command to update
your credentials at any time.

You'll need:
  - Client ID and Auth Token from Account Settings > API > Generate Token
  - Company ID (your org slug from the URL, e.g. "acme" from blue.app/org/acme)

Non-interactive use (scripts, agents): pass --client-id, --auth-token,
and --company-id together to skip all prompts.`,
	Example: `  blue init
  blue init --client-id <ID> --auth-token <SECRET> --company-id acme`,
	RunE: runInit,
}

var (
	initAPIURL    string
	initClientID  string
	initAuthToken string
	initCompanyID string
)

func init() {
	rootCmd.AddCommand(initCmd)

	initCmd.Flags().StringVar(&initAPIURL, "api-url", "", "API URL (default: "+common.DefaultAPIUrl+")")
	initCmd.Flags().StringVar(&initClientID, "client-id", "", "Client ID (skips prompt)")
	initCmd.Flags().StringVar(&initAuthToken, "auth-token", "", "Auth token / secret (skips prompt)")
	initCmd.Flags().StringVar(&initCompanyID, "company-id", "", "Company ID / org slug (skips prompt)")
}

func runInit(cmd *cobra.Command, args []string) error {
	apiUrl, clientID, authToken, companyID := initAPIURL, initClientID, initAuthToken, initCompanyID

	flagsProvided := clientID != "" || authToken != "" || companyID != ""
	flagsComplete := clientID != "" && authToken != "" && companyID != ""

	if flagsProvided && !flagsComplete {
		return fmt.Errorf("--client-id, --auth-token, and --company-id must all be provided together for non-interactive setup")
	}

	if apiUrl == "" {
		apiUrl = common.DefaultAPIUrl
	}

	if !flagsComplete {
		reader := bufio.NewReader(os.Stdin)

		fmt.Println("Welcome to Blue CLI setup!")
		fmt.Println()
		fmt.Println("You'll need your API credentials from Blue.")
		fmt.Println("Go to: Account Settings > API > Generate Token")
		fmt.Println()

		// API URL with default
		fmt.Printf("API URL [%s]: ", apiUrl)
		input, _ := reader.ReadString('\n')
		if input = strings.TrimSpace(input); input != "" {
			apiUrl = input
		}

		// Client ID
		fmt.Print("Client ID: ")
		clientID, _ = reader.ReadString('\n')
		clientID = strings.TrimSpace(clientID)
		if clientID == "" {
			return fmt.Errorf("client ID is required")
		}

		// Auth Token
		fmt.Print("Auth Token (Secret): ")
		authToken, _ = reader.ReadString('\n')
		authToken = strings.TrimSpace(authToken)
		if authToken == "" {
			return fmt.Errorf("auth token is required")
		}

		// Company ID
		fmt.Print("Company ID (org slug from URL, e.g. \"acme\" from blue.app/org/acme): ")
		companyID, _ = reader.ReadString('\n')
		companyID = strings.TrimSpace(companyID)
		if companyID == "" {
			return fmt.Errorf("company ID is required")
		}
	}

	// Create config directory
	configDir := common.ConfigDir()
	if configDir == "" {
		return fmt.Errorf("could not determine config directory")
	}

	if err := os.MkdirAll(configDir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Write config file
	configPath := common.ConfigPath()
	content := fmt.Sprintf("API_URL=%s\nAUTH_TOKEN=%s\nCLIENT_ID=%s\nCOMPANY_ID=%s\n",
		apiUrl, authToken, clientID, companyID)

	if err := os.WriteFile(configPath, []byte(content), 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	// Add to known companies list
	if err := common.AddCompany(companyID); err != nil {
		fmt.Printf("Warning: could not update companies list: %v\n", err)
	}

	fmt.Println()
	fmt.Printf("Config saved to %s\n", configPath)
	fmt.Println()

	// Test the connection
	fmt.Println("Testing connection...")
	config, err := common.LoadConfig()
	if err != nil {
		fmt.Printf("Warning: could not verify credentials: %v\n", err)
		fmt.Println("Config was saved — you can update it by running 'blue init' again.")
		return nil
	}

	client := common.NewClient(config)

	// Quick test query
	query := fmt.Sprintf(`query { projectList(filter: { companyIds: ["%s"] }, take: 1) { totalCount } }`, companyID)
	var response struct {
		ProjectList struct {
			TotalCount int `json:"totalCount"`
		} `json:"projectList"`
	}

	if err := client.ExecuteQueryWithResult(query, nil, &response); err != nil {
		fmt.Printf("Warning: connection test failed: %v\n", err)
		fmt.Println("Config was saved — check your credentials and try again.")
		return nil
	}

	fmt.Printf("Connected! Found %d workspaces in your account.\n", response.ProjectList.TotalCount)
	fmt.Println()
	fmt.Println("You're all set. Try: blue workspaces list --simple")

	return nil
}
