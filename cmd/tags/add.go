package tags

import (
	"fmt"
	"strings"

	"blue-cli/common"

	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add tags to a record",
	Long:  "Add existing tags by ID or new tags by title to a record.",
	Example: `  blue tags add --record <id> --tag-ids "tag1,tag2"
  blue tags add --record <id> --tag-titles "Bug,Priority" --workspace <id>`,
	RunE: runAdd,
}

var (
	addRecord    string
	addTagIDs    string
	addTagTitles string
	addWorkspace string
	addSimple    bool
)

func init() {
	addCmd.Flags().StringVarP(&addRecord, "record", "r", "", "Record ID to add tags to (required)")
	addCmd.Flags().StringVar(&addTagIDs, "tag-ids", "", "Comma-separated tag IDs to add")
	addCmd.Flags().StringVar(&addTagTitles, "tag-titles", "", "Comma-separated tag titles to add")
	addCmd.Flags().StringVarP(&addWorkspace, "workspace", "w", "", "Workspace ID (required when using --tag-titles)")
	addCmd.Flags().BoolVarP(&addSimple, "simple", "s", false, "Simple output format")
}

func runAdd(cmd *cobra.Command, args []string) error {
	if addRecord == "" {
		return fmt.Errorf("record ID is required. Use --record flag")
	}
	if addTagIDs == "" && addTagTitles == "" {
		return fmt.Errorf("either --tag-ids or --tag-titles must be provided")
	}
	if addTagTitles != "" && addWorkspace == "" {
		return fmt.Errorf("workspace is required when using --tag-titles. Use --workspace flag")
	}

	config, err := common.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	client := common.NewClient(config)

	if addWorkspace != "" {
		client.SetProject(addWorkspace)
	}

	mutation := `
		mutation SetTodoTags($input: SetTodoTagsInput!) {
			setTodoTags(input: $input)
		}
	`

	input := map[string]interface{}{
		"todoId": addRecord,
	}

	if addTagIDs != "" {
		ids := strings.Split(addTagIDs, ",")
		for i, id := range ids {
			ids[i] = strings.TrimSpace(id)
		}
		input["tagIds"] = ids
	}

	if addTagTitles != "" {
		titles := strings.Split(addTagTitles, ",")
		for i, t := range titles {
			titles[i] = strings.TrimSpace(t)
		}
		input["tagTitles"] = titles
	}

	variables := map[string]interface{}{
		"input": input,
	}

	var response struct {
		SetTodoTags bool `json:"setTodoTags"`
	}

	if err := client.ExecuteQueryWithResult(mutation, variables, &response); err != nil {
		return fmt.Errorf("failed to add tags: %w", err)
	}

	if response.SetTodoTags {
		if addSimple {
			fmt.Printf("Tags added to record %s\n", addRecord)
		} else {
			fmt.Printf("Tags successfully added to record!\n")
			fmt.Printf("Record ID: %s\n", addRecord)
		}
	} else {
		return fmt.Errorf("failed to add tags to record")
	}

	return nil
}
