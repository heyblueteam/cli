package items

import (
	"fmt"

	"blue-cli/common"

	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a checklist item",
	Long:  "Update a checklist item's title, position, done status, or move it to another checklist.",
	Example: `  blue checklists items update --item <id> --done true
  blue checklists items update --item <id> --title "Updated title" --position 1500.0
  blue checklists items update --item <id> --move-to-checklist <checklist-id>
  blue checklists items update --item <id> --done false --simple`,
	RunE: runUpdate,
}

var (
	updateItem            string
	updateTitle           string
	updatePosition        float64
	updateDone            string
	updateMoveToChecklist string
	updateWorkspace       string
	updateSimple          bool
)

func init() {
	updateCmd.Flags().StringVar(&updateItem, "item", "", "Checklist item ID to update (required)")
	updateCmd.Flags().StringVarP(&updateTitle, "title", "t", "", "New title for the checklist item")
	updateCmd.Flags().Float64Var(&updatePosition, "position", -1, "New position for the checklist item")
	updateCmd.Flags().StringVar(&updateDone, "done", "", "Mark item as done (true/false)")
	updateCmd.Flags().StringVar(&updateMoveToChecklist, "move-to-checklist", "", "Move item to a different checklist (checklist ID)")
	updateCmd.Flags().StringVarP(&updateWorkspace, "workspace", "w", "", "Workspace ID or slug (optional)")
	updateCmd.Flags().BoolVarP(&updateSimple, "simple", "s", false, "Simple output format")
}

func runUpdate(cmd *cobra.Command, args []string) error {
	if updateItem == "" {
		return fmt.Errorf("checklist item ID is required. Use --item flag")
	}

	if updateTitle == "" && updatePosition == -1 && updateDone == "" && updateMoveToChecklist == "" {
		return fmt.Errorf("at least one field to update must be provided (--title, --position, --done, or --move-to-checklist)")
	}

	config, err := common.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	client := common.NewClient(config)

	if updateWorkspace != "" {
		client.SetProject(updateWorkspace)
	}

	// Build input map with only provided fields
	input := map[string]interface{}{
		"checklistItemId": updateItem,
	}

	if updateTitle != "" {
		input["title"] = updateTitle
	}
	if updatePosition != -1 {
		input["position"] = updatePosition
	}
	if updateMoveToChecklist != "" {
		input["checklistId"] = updateMoveToChecklist
	}
	if updateDone != "" {
		if updateDone == "true" {
			input["done"] = true
		} else if updateDone == "false" {
			input["done"] = false
		} else {
			return fmt.Errorf("--done flag must be 'true' or 'false'")
		}
	}

	mutation := `
		mutation EditChecklistItem($input: EditChecklistItemInput!) {
			editChecklistItem(input: $input) {
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
		"input": input,
	}

	var response struct {
		EditChecklistItem struct {
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
		} `json:"editChecklistItem"`
	}

	if !updateSimple {
		fmt.Printf("=== Updating Checklist Item ===\n")
		fmt.Printf("Item ID: %s\n", updateItem)
		if updateTitle != "" {
			fmt.Printf("New Title: %s\n", updateTitle)
		}
		if updatePosition != -1 {
			fmt.Printf("New Position: %.1f\n", updatePosition)
		}
		if updateMoveToChecklist != "" {
			fmt.Printf("Move to Checklist: %s\n", updateMoveToChecklist)
		}
		if updateDone != "" {
			fmt.Printf("Done Status: %s\n", updateDone)
		}
		if updateWorkspace != "" {
			fmt.Printf("Workspace: %s\n", updateWorkspace)
		}
		fmt.Printf("\n")
	}

	if err := client.ExecuteQueryWithResult(mutation, variables, &response); err != nil {
		return fmt.Errorf("failed to update checklist item: %w", err)
	}

	item := response.EditChecklistItem

	if updateSimple {
		fmt.Printf("Checklist Item Updated: %s\n", item.ID)
	} else {
		fmt.Printf("=== Checklist Item Updated Successfully ===\n")
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
		fmt.Printf("Updated: %s\n", item.UpdatedAt)
		fmt.Printf("Created By: %s (%s)\n", item.CreatedBy.FullName, item.CreatedBy.Email)
		fmt.Printf("Checklist item updated successfully!\n")
	}

	return nil
}
