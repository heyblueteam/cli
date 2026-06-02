package tags

import (
	"fmt"
	"strings"

	"github.com/heyblueteam/cli/common"

	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a tag",
	Long:  "Update a tag title or hex color.",
	Example: `  blue tags update --tag <id> --color "#ff0000"
  blue tags update --tag <id> --title "Urgent" --color "#ff8800"`,
	RunE: runUpdate,
}

var (
	updateTag   string
	updateTitle string
	updateColor string
)

func init() {
	updateCmd.Flags().StringVar(&updateTag, "tag", "", "Tag ID (required)")
	updateCmd.Flags().StringVarP(&updateTitle, "title", "t", "", "New tag title")
	updateCmd.Flags().StringVar(&updateColor, "color", "", "New tag hex color")
}

func runUpdate(cmd *cobra.Command, args []string) error {
	if updateTag == "" {
		return fmt.Errorf("tag ID is required. Use --tag flag")
	}
	if updateTitle == "" && updateColor == "" {
		return fmt.Errorf("at least one field must be specified for update (--title or --color)")
	}

	input := map[string]interface{}{
		"id": updateTag,
	}

	if updateTitle != "" {
		input["title"] = strings.TrimSpace(updateTitle)
	}

	if updateColor != "" {
		color, err := common.NormalizeHexColor(updateColor)
		if err != nil {
			return err
		}
		input["color"] = color
	}

	config, err := common.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	client := common.NewClient(config)

	mutation := `
		mutation EditTag($input: EditTagInput!) {
			editTag(input: $input) {
				id
				uid
				title
				color
				updatedAt
			}
		}
	`

	variables := map[string]interface{}{
		"input": input,
	}

	var response struct {
		EditTag common.Tag `json:"editTag"`
	}

	if err := client.ExecuteQueryWithResult(mutation, variables, &response); err != nil {
		return fmt.Errorf("failed to update tag: %w", err)
	}

	tag := response.EditTag
	fmt.Printf("Tag updated successfully!\n")
	fmt.Printf("Title: %s\n", tag.Title)
	fmt.Printf("ID: %s\n", tag.ID)
	fmt.Printf("Color: %s\n", tag.Color)

	return nil
}
