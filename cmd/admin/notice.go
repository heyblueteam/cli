package admin

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/heyblueteam/cli/common"
	"github.com/spf13/cobra"
)

var noticeCmd = &cobra.Command{
	Use:   "notice",
	Short: "Manage the authenticated app's top-bar notice",
}

var noticeShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show the configured top-bar notice",
	RunE:  runNoticeShow,
}

var noticeSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set the authenticated app's top-bar notice",
	Example: `  blue admin notice set --id maintenance-2026-09-12 \
    --message "Scheduled maintenance" \
    --variant maintenance \
    --link https://status.blue.app \
    --link-label "Learn more" \
    --starts-at 2026-09-12T02:00:00Z \
    --ends-at 2026-09-12T04:00:00Z

  blue admin notice set --id release-2026-09 --message "New features are live"`,
	RunE: runNoticeSet,
}

var noticeClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Remove the top-bar notice",
	RunE:  runNoticeClear,
}

var (
	noticeID          string
	noticeVariant     string
	noticeMessage     string
	noticeLink        string
	noticeLinkLabel   string
	noticeDetails     string
	noticeStartsAt    string
	noticeEndsAt      string
	noticeDismissible bool
	noticeFormat      string
)

type appNotice struct {
	Enabled     bool    `json:"enabled"`
	ID          *string `json:"id"`
	Variant     *string `json:"variant"`
	Message     *string `json:"message"`
	Href        *string `json:"href"`
	LinkLabel   *string `json:"linkLabel"`
	Details     *string `json:"details"`
	StartsAt    *string `json:"startsAt"`
	EndsAt      *string `json:"endsAt"`
	Dismissible *bool   `json:"dismissible"`
}

func init() {
	noticeCmd.AddCommand(noticeShowCmd, noticeSetCmd, noticeClearCmd)

	noticeShowCmd.Flags().StringVar(&noticeFormat, "format", "text", "Output format: text, json")

	noticeSetCmd.Flags().StringVar(&noticeID, "id", "", "Stable notice identifier")
	noticeSetCmd.Flags().StringVar(&noticeVariant, "variant", "info", "Notice style: info, warning, maintenance")
	noticeSetCmd.Flags().StringVar(&noticeMessage, "message", "", "Top-bar message")
	noticeSetCmd.Flags().StringVar(&noticeLink, "link", "", "Optional relative or HTTPS Blue URL")
	noticeSetCmd.Flags().StringVar(&noticeLinkLabel, "link-label", "", "Optional link label")
	noticeSetCmd.Flags().StringVar(&noticeDetails, "details", "", "Optional details shown in a modal")
	noticeSetCmd.Flags().StringVar(&noticeStartsAt, "starts-at", "", "Optional UTC start time in RFC3339 format")
	noticeSetCmd.Flags().StringVar(&noticeEndsAt, "ends-at", "", "Optional UTC end time in RFC3339 format")
	noticeSetCmd.Flags().BoolVar(&noticeDismissible, "dismissible", true, "Allow users to dismiss the notice")

	_ = noticeSetCmd.MarkFlagRequired("id")
	_ = noticeSetCmd.MarkFlagRequired("message")
}

func loadClient() (*common.Client, error) {
	config, err := common.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}
	return common.NewClient(config), nil
}

func runNoticeShow(cmd *cobra.Command, args []string) error {
	client, err := loadClient()
	if err != nil {
		return err
	}

	query := `query AdminAppNotice {
		adminQueries {
			appNotice {
				enabled id variant message href linkLabel details startsAt endsAt dismissible
			}
		}
	}`
	var response struct {
		AdminQueries struct {
			AppNotice appNotice `json:"appNotice"`
		} `json:"adminQueries"`
	}
	if err := client.ExecuteQueryWithResult(query, nil, &response); err != nil {
		return fmt.Errorf("failed to read app notice: %w", err)
	}

	switch noticeFormat {
	case "text":
		printNotice(response.AdminQueries.AppNotice)
		return nil
	case "json":
		output, err := json.MarshalIndent(response.AdminQueries.AppNotice, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(output))
		return nil
	default:
		return fmt.Errorf("invalid format %q. Use text or json", noticeFormat)
	}
}

func runNoticeSet(cmd *cobra.Command, args []string) error {
	variant, err := normalizeVariant(noticeVariant)
	if err != nil {
		return err
	}

	client, err := loadClient()
	if err != nil {
		return err
	}

	mutation := `mutation SetAppNotice($input: AppNoticeInput!) {
		adminMutations { setAppNotice(input: $input) }
	}`
	input := map[string]interface{}{
		"id":          noticeID,
		"variant":     variant,
		"message":     noticeMessage,
		"dismissible": noticeDismissible,
	}
	if noticeLink != "" {
		input["href"] = noticeLink
	}
	if noticeLinkLabel != "" {
		input["linkLabel"] = noticeLinkLabel
	}
	if noticeDetails != "" {
		input["details"] = noticeDetails
	}
	if noticeStartsAt != "" {
		input["startsAt"] = noticeStartsAt
	}
	if noticeEndsAt != "" {
		input["endsAt"] = noticeEndsAt
	}

	var response struct {
		AdminMutations struct {
			SetAppNotice bool `json:"setAppNotice"`
		} `json:"adminMutations"`
	}
	if err := client.ExecuteQueryWithResult(mutation, map[string]interface{}{"input": input}, &response); err != nil {
		return fmt.Errorf("failed to set app notice: %w", err)
	}
	if !response.AdminMutations.SetAppNotice {
		return fmt.Errorf("API did not confirm the app notice update")
	}

	fmt.Printf("App notice %q is now active.\n", noticeID)
	return nil
}

func runNoticeClear(cmd *cobra.Command, args []string) error {
	client, err := loadClient()
	if err != nil {
		return err
	}

	mutation := `mutation ClearAppNotice { adminMutations { clearAppNotice } }`
	var response struct {
		AdminMutations struct {
			ClearAppNotice bool `json:"clearAppNotice"`
		} `json:"adminMutations"`
	}
	if err := client.ExecuteQueryWithResult(mutation, nil, &response); err != nil {
		return fmt.Errorf("failed to clear app notice: %w", err)
	}
	if !response.AdminMutations.ClearAppNotice {
		return fmt.Errorf("API did not confirm clearing the app notice")
	}

	fmt.Println("App notice cleared.")
	return nil
}

func normalizeVariant(value string) (string, error) {
	switch strings.ToUpper(strings.TrimSpace(value)) {
	case "INFO":
		return "INFO", nil
	case "WARNING":
		return "WARNING", nil
	case "MAINTENANCE":
		return "MAINTENANCE", nil
	default:
		return "", fmt.Errorf("invalid variant %q. Use info, warning, or maintenance", value)
	}
}

func printNotice(notice appNotice) {
	if !notice.Enabled {
		fmt.Println("No app notice is active.")
		return
	}
	fmt.Println("App notice")
	printNoticeField("ID", notice.ID)
	printNoticeField("Variant", notice.Variant)
	printNoticeField("Message", notice.Message)
	printNoticeField("Link", notice.Href)
	printNoticeField("Link label", notice.LinkLabel)
	printNoticeField("Details", notice.Details)
	printNoticeField("Starts at", notice.StartsAt)
	printNoticeField("Ends at", notice.EndsAt)
	if notice.Dismissible != nil {
		fmt.Printf("Dismissible: %t\n", *notice.Dismissible)
	}
}

func printNoticeField(label string, value *string) {
	if value != nil && *value != "" {
		fmt.Printf("%s: %s\n", label, *value)
	}
}
