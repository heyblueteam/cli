package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"blue-cli/common"

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
  - Company ID (your org slug from the URL, e.g. "acme" from blue.cc/org/acme)`,
	RunE: runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Welcome to Blue CLI setup!")
	fmt.Println()
	fmt.Println("You'll need your API credentials from Blue.")
	fmt.Println("Go to: Account Settings > API > Generate Token")
	fmt.Println()

	// API URL with default
	fmt.Printf("API URL [%s]: ", common.DefaultAPIUrl)
	apiUrl, _ := reader.ReadString('\n')
	apiUrl = strings.TrimSpace(apiUrl)
	if apiUrl == "" {
		apiUrl = common.DefaultAPIUrl
	}

	// Client ID
	fmt.Print("Client ID: ")
	clientID, _ := reader.ReadString('\n')
	clientID = strings.TrimSpace(clientID)
	if clientID == "" {
		return fmt.Errorf("client ID is required")
	}

	// Auth Token
	fmt.Print("Auth Token (Secret): ")
	authToken, _ := reader.ReadString('\n')
	authToken = strings.TrimSpace(authToken)
	if authToken == "" {
		return fmt.Errorf("auth token is required")
	}

	// Company ID
	fmt.Print("Company ID (org slug from URL, e.g. \"acme\" from blue.cc/org/acme): ")
	companyID, _ := reader.ReadString('\n')
	companyID = strings.TrimSpace(companyID)
	if companyID == "" {
		return fmt.Errorf("company ID is required")
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
