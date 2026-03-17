package tags

import (
	"fmt"
	"strings"

	"blue-cli/common"

	"github.com/spf13/cobra"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new tag",
	Long:  "Create a new tag within a workspace.",
	Example: `  blue tags create --workspace <id> --title "Bug" --color red
  blue tags create -w <id> -t "Feature" --color blue`,
	RunE: runCreate,
}

var (
	createWorkspace string
	createTitle     string
	createColor     string
)

func init() {
	createCmd.Flags().StringVarP(&createWorkspace, "workspace", "w", "", "Workspace ID (required)")
	createCmd.Flags().StringVarP(&createTitle, "title", "t", "", "Tag title (required)")
	createCmd.Flags().StringVar(&createColor, "color", "", "Tag color (required)")
}

func runCreate(cmd *cobra.Command, args []string) error {
	if createWorkspace == "" {
		return fmt.Errorf("workspace ID is required. Use --workspace flag")
	}
	if createTitle == "" {
		return fmt.Errorf("tag title is required. Use --title flag")
	}
	if createColor == "" {
		return fmt.Errorf("tag color is required. Use --color flag")
	}

	config, err := common.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	client := common.NewClient(config)
	client.SetProject(createWorkspace)

	mutation := `
		mutation CreateTag($input: CreateTagInput!) {
			createTag(input: $input) {
				id
				uid
				title
				color
				createdAt
			}
		}
	`

	variables := map[string]interface{}{
		"input": map[string]interface{}{
			"title": strings.TrimSpace(createTitle),
			"color": strings.TrimSpace(createColor),
		},
	}

	var response struct {
		CreateTag common.Tag `json:"createTag"`
	}

	if err := client.ExecuteQueryWithResult(mutation, variables, &response); err != nil {
		return fmt.Errorf("failed to create tag: %w", err)
	}

	tag := response.CreateTag
	fmt.Printf("Tag created successfully!\n")
	fmt.Printf("Title: %s\n", tag.Title)
	fmt.Printf("ID: %s\n", tag.ID)
	fmt.Printf("Color: %s\n", tag.Color)

	return nil
}
