package items

import (
	"fmt"

	"blue-cli/common"

	"github.com/spf13/cobra"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a checklist item",
	Long:  "Create a new item within a checklist.",
	Example: `  blue checklists items create --checklist <id> --title "Review docs"
  blue checklists items create --checklist <id> --title "Run tests" --position 2000.0
  blue checklists items create --checklist <id> -t "Deploy" --simple`,
	RunE: runCreate,
}

var (
	createChecklist string
	createTitle     string
	createPosition  float64
	createWorkspace string
	createSimple    bool
)

func init() {
	createCmd.Flags().StringVar(&createChecklist, "checklist", "", "Checklist ID to add item to (required)")
	createCmd.Flags().StringVarP(&createTitle, "title", "t", "", "Checklist item title (required)")
	createCmd.Flags().Float64Var(&createPosition, "position", 1000.0, "Position of the checklist item (default: 1000.0)")
	createCmd.Flags().StringVarP(&createWorkspace, "workspace", "w", "", "Workspace ID or slug (optional)")
	createCmd.Flags().BoolVarP(&createSimple, "simple", "s", false, "Simple output format")
}

func runCreate(cmd *cobra.Command, args []string) error {
	if createChecklist == "" {
		return fmt.Errorf("checklist ID is required. Use --checklist flag")
	}
	if createTitle == "" {
		return fmt.Errorf("checklist item title is required. Use --title flag")
	}

	config, err := common.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	client := common.NewClient(config)

	if createWorkspace != "" {
		client.SetProject(createWorkspace)
	}

	mutation := `
		mutation CreateChecklistItem($input: CreateChecklistItemInput!) {
			createChecklistItem(input: $input) {
				id
				uid
				title
				position
				done
				startedAt
				duedAt
				createdAt
				updatedAt
				createdBy {
					id
					uid
					fullName
					email
				}
			}
		}
	`

	variables := map[string]interface{}{
		"input": map[string]interface{}{
			"checklistId": createChecklist,
			"title":       createTitle,
			"position":    createPosition,
		},
	}

	var response struct {
		CreateChecklistItem struct {
			ID        string      `json:"id"`
			UID       string      `json:"uid"`
			Title     string      `json:"title"`
			Position  float64     `json:"position"`
			Done      bool        `json:"done"`
			StartedAt *string     `json:"startedAt"`
			DuedAt    *string     `json:"duedAt"`
			CreatedAt string      `json:"createdAt"`
			UpdatedAt string      `json:"updatedAt"`
			CreatedBy common.User `json:"createdBy"`
		} `json:"createChecklistItem"`
	}

	if !createSimple {
		fmt.Printf("=== Creating Checklist Item ===\n")
		fmt.Printf("Checklist ID: %s\n", createChecklist)
		fmt.Printf("Title: %s\n", createTitle)
		fmt.Printf("Position: %.1f\n", createPosition)
		if createWorkspace != "" {
			fmt.Printf("Workspace: %s\n", createWorkspace)
		}
		fmt.Printf("\n")
	}

	if err := client.ExecuteQueryWithResult(mutation, variables, &response); err != nil {
		return fmt.Errorf("failed to create checklist item: %w", err)
	}

	item := response.CreateChecklistItem

	if createSimple {
		fmt.Printf("Checklist Item ID: %s\n", item.ID)
	} else {
		fmt.Printf("=== Checklist Item Created Successfully ===\n")
		fmt.Printf("ID: %s\n", item.ID)
		fmt.Printf("UID: %s\n", item.UID)
		fmt.Printf("Title: %s\n", item.Title)
		fmt.Printf("Position: %.1f\n", item.Position)
		fmt.Printf("Done: %t\n", item.Done)
		if item.StartedAt != nil {
			fmt.Printf("Started: %s\n", *item.StartedAt)
		}
		if item.DuedAt != nil {
			fmt.Printf("Due: %s\n", *item.DuedAt)
		}
		fmt.Printf("Created: %s\n", item.CreatedAt)
		fmt.Printf("Created By: %s (%s)\n", item.CreatedBy.FullName, item.CreatedBy.Email)
		fmt.Printf("Checklist item created successfully!\n")
	}

	return nil
}
