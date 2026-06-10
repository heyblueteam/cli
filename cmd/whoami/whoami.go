package whoami

import (
	"encoding/json"
	"fmt"

	"github.com/heyblueteam/cli/common"
	"github.com/spf13/cobra"
)

// Cmd shows the authenticated user and active CLI context.
var Cmd = &cobra.Command{
	Use:   "whoami",
	Short: "Show authenticated identity and context",
	Long:  "Show the current Blue user, configured company, API endpoint, and config path.",
	Example: `  blue whoami
  blue whoami --format json`,
	RunE: runWhoami,
}

var whoamiFormat string

type whoamiResult struct {
	User       whoamiUser    `json:"user"`
	Company    whoamiCompany `json:"company"`
	APIURL     string        `json:"apiUrl"`
	ConfigPath string        `json:"configPath"`
}

type whoamiUser struct {
	ID        string `json:"id"`
	UID       string `json:"uid"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	FullName  string `json:"fullName"`
	Locale    string `json:"locale"`
	Timezone  string `json:"timezone"`
	CreatedAt string `json:"createdAt"`
}

type whoamiCompany struct {
	ID          string `json:"id"`
	UID         string `json:"uid"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	AccessLevel string `json:"accessLevel"`
	Plan        string `json:"plan"`
	Tier        string `json:"tier"`
	CreatedAt   string `json:"createdAt"`
}

func init() {
	Cmd.Flags().StringVar(&whoamiFormat, "format", "text", "Output format: text, json")
}

func runWhoami(cmd *cobra.Command, args []string) error {
	config, err := common.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}
	client := common.NewClient(config)

	query := `query Whoami($companyId: String) {
		currentUser {
			id uid username email fullName locale timezone createdAt
		}
		company(id: $companyId) {
			id uid name slug accessLevel plan tier createdAt
		}
	}`
	variables := map[string]interface{}{"companyId": config.CompanyID}

	var response struct {
		CurrentUser whoamiUser    `json:"currentUser"`
		Company     whoamiCompany `json:"company"`
	}
	if err := client.ExecuteQueryWithResult(query, variables, &response); err != nil {
		return fmt.Errorf("failed to read identity: %w", err)
	}

	result := whoamiResult{
		User:       response.CurrentUser,
		Company:    response.Company,
		APIURL:     config.APIUrl,
		ConfigPath: common.ConfigPath(),
	}

	switch whoamiFormat {
	case "text":
		printText(result)
		return nil
	case "json":
		out, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(out))
		return nil
	default:
		return fmt.Errorf("invalid format %q. Use text or json", whoamiFormat)
	}
}

func printText(result whoamiResult) {
	fmt.Println("User")
	fmt.Printf("  Name:     %s\n", result.User.FullName)
	fmt.Printf("  Email:    %s\n", result.User.Email)
	fmt.Printf("  Username: %s\n", result.User.Username)
	fmt.Printf("  ID:       %s\n", result.User.ID)
	if result.User.Timezone != "" {
		fmt.Printf("  Timezone: %s\n", result.User.Timezone)
	}

	fmt.Println("\nCompany")
	fmt.Printf("  Name:        %s\n", result.Company.Name)
	fmt.Printf("  Slug:        %s\n", result.Company.Slug)
	fmt.Printf("  ID:          %s\n", result.Company.ID)
	fmt.Printf("  Access:      %s\n", result.Company.AccessLevel)
	if result.Company.Plan != "" {
		fmt.Printf("  Plan:        %s\n", result.Company.Plan)
	}
	if result.Company.Tier != "" {
		fmt.Printf("  Tier:        %s\n", result.Company.Tier)
	}

	fmt.Println("\nConfig")
	fmt.Printf("  API URL:     %s\n", result.APIURL)
	fmt.Printf("  Config path: %s\n", result.ConfigPath)
}
