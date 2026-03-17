package checklists

import (
	"fmt"

	"blue-cli/common"

	"github.com/spf13/cobra"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a checklist on a record",
	Long:  "Create a new checklist on a specific record.",
	Example: `  blue checklists create --record <id> --title "Pre-launch Checklist"
  blue checklists create --record <id> --title "QA Tasks" --position 2000.0
  blue checklists create -r <id> -t "Deploy Steps" --workspace <id> --simple`,
	RunE: runCreate,
}

var (
	createRecord    string
	createTitle     string
	createPosition  float64
	createWorkspace string
	createSimple    bool
)

func init() {
	createCmd.Flags().StringVarP(&createRecord, "record", "r", "", "Record/Todo ID to add checklist to (required)")
	createCmd.Flags().StringVarP(&createTitle, "title", "t", "", "Checklist title (required)")
	createCmd.Flags().Float64Var(&createPosition, "position", 1000.0, "Position of the checklist (default: 1000.0)")
	createCmd.Flags().StringVarP(&createWorkspace, "workspace", "w", "", "Workspace ID or slug (optional)")
	createCmd.Flags().BoolVarP(&createSimple, "simple", "s", false, "Simple output format")
}

func runCreate(cmd *cobra.Command, args []string) error {
	if createRecord == "" {
		return fmt.Errorf("record ID is required. Use --record flag")
	}
	if createTitle == "" {
		return fmt.Errorf("checklist title is required. Use --title flag")
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
		mutation CreateChecklist($input: CreateChecklistInput!) {
			createChecklist(input: $input) {
				id
				uid
				title
				position
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
			"todoId":   createRecord,
			"title":    createTitle,
			"position": createPosition,
		},
	}

	var response struct {
		CreateChecklist struct {
			ID        string      `json:"id"`
			UID       string      `json:"uid"`
			Title     string      `json:"title"`
			Position  float64     `json:"position"`
			CreatedAt string      `json:"createdAt"`
			UpdatedAt string      `json:"updatedAt"`
			CreatedBy common.User `json:"createdBy"`
		} `json:"createChecklist"`
	}

	if !createSimple {
		fmt.Printf("=== Creating Checklist ===\n")
		fmt.Printf("Record ID: %s\n", createRecord)
		fmt.Printf("Title: %s\n", createTitle)
		fmt.Printf("Position: %.1f\n", createPosition)
		if createWorkspace != "" {
			fmt.Printf("Workspace: %s\n", createWorkspace)
		}
		fmt.Printf("\n")
	}

	if err := client.ExecuteQueryWithResult(mutation, variables, &response); err != nil {
		return fmt.Errorf("failed to create checklist: %w", err)
	}

	checklist := response.CreateChecklist

	if createSimple {
		fmt.Printf("Checklist ID: %s\n", checklist.ID)
	} else {
		fmt.Printf("=== Checklist Created Successfully ===\n")
		fmt.Printf("ID: %s\n", checklist.ID)
		fmt.Printf("UID: %s\n", checklist.UID)
		fmt.Printf("Title: %s\n", checklist.Title)
		fmt.Printf("Position: %.1f\n", checklist.Position)
		fmt.Printf("Created: %s\n", checklist.CreatedAt)
		fmt.Printf("Created By: %s (%s)\n", checklist.CreatedBy.FullName, checklist.CreatedBy.Email)
		fmt.Printf("Checklist created successfully!\n")
	}

	return nil
}
