package main

import (
	"fmt"
	"os"
	"os/exec"
	
	"demo-builder/tools"
)

func printUsage() {
	fmt.Println("Blue Demo Builder - CLI Tool")
	fmt.Println()
	fmt.Println("Usage: go run . <command> [flags]")
	fmt.Println()
	fmt.Println("Available commands:")
	fmt.Println()
	fmt.Println("READ operations:")
	fmt.Println("  read-projects               List all projects")
	fmt.Println("  read-lists                  List todo lists in a project")
	fmt.Println("  read-record                 Get detailed record information")
	fmt.Println("  read-records                Query records with advanced filtering and statistics")
	fmt.Println("  read-list-records           List records in a specific list")
	fmt.Println("  read-project-records        List all records in a project by list")
	fmt.Println("")
	fmt.Println("  read-records-count          Count records in a project")
	fmt.Println("  read-tags                   List tags in a project")
	fmt.Println("  read-project-custom-fields  List custom fields in a project")
	fmt.Println("  read-custom-fields          Enhanced custom fields reference for record operations")
	fmt.Println("  read-automations            List automations in a project")
	fmt.Println("  read-user-profiles          List user profiles in a company")
	fmt.Println("  read-project-user-roles     List custom user roles in projects")
	fmt.Println()
	fmt.Println("CREATE operations:")
	fmt.Println("  create-project              Create a new project")
	fmt.Println("  create-list                 Create a new todo list")
	fmt.Println("  create-record               Create a new record/todo")
	fmt.Println("  create-comment              Create a comment on a record")
	fmt.Println("  create-tags                 Create new tags")
	fmt.Println("  create-record-tags          Add tags to a record")
	fmt.Println("  create-custom-field         Create a custom field")
	fmt.Println("  create-custom-field-options Create options for existing custom fields")
	fmt.Println("  create-automation           Create a new automation")
	fmt.Println("  create-automation-multi     Create automation with multiple actions")
	fmt.Println("  invite-user                 Invite a user to the company or project")
	fmt.Println()
	fmt.Println("UPDATE operations:")
	fmt.Println("  update-project              Update project settings")
	fmt.Println("  update-record               Update a record/todo")
	fmt.Println("  update-comment              Update a comment")
	fmt.Println("  update-custom-field         Update custom field properties")
	fmt.Println("  update-list                 Update list properties")
	fmt.Println("  update-automation           Update an existing automation")
	fmt.Println("  update-automation-multi     Update automation with multiple actions")
	fmt.Println()
	fmt.Println("DELETE operations:")
	fmt.Println("  delete-project              Delete a project")
	fmt.Println("  delete-record               Delete a record/todo")
	fmt.Println("  delete-custom-field         Delete a custom field")
	fmt.Println("  delete-custom-field-options Delete options from custom fields")
	fmt.Println("  delete-automation           Delete an automation")
	fmt.Println()
	fmt.Println("Testing:")
	fmt.Println("  e2e                         Run end-to-end tests")
	fmt.Println()
	fmt.Println("Use '<command> -h' for help with a specific command")
}

func runE2E(args []string) error {
	// Run the e2e test from test directory
	cmd := exec.Command("go", append([]string{"run", "test/e2e.go"}, args...)...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]
	args := os.Args[2:]

	var err error
	switch command {
	// READ operations
	case "read-projects":
		err = tools.RunReadProjects(args)
	case "read-lists":
		err = tools.RunReadLists(args)
	case "read-record":
		err = tools.RunReadRecord(args)
	case "read-records":
		err = tools.RunReadRecords(args)
	case "read-list-records":
		err = tools.RunReadTodos(args)
	case "read-project-records":
		err = tools.RunReadProjectRecords(args)
	case "read-records-count":
		err = tools.RunReadRecordsCount(args)
	case "read-tags":
		err = tools.RunReadTags(args)
	case "read-project-custom-fields":
		err = tools.RunReadProjectCustomFields(args)
	case "read-custom-fields":
		err = tools.RunReadCustomFields(args)
	case "read-automations":
		err = tools.RunReadAutomations(args)
	case "read-user-profiles":
		err = tools.RunReadUserProfiles(args)
	case "read-project-user-roles":
		err = tools.RunReadProjectUserRoles(args)
	
	// CREATE operations
	case "create-project":
		err = tools.RunCreateProject(args)
	case "create-list":
		err = tools.RunCreateList(args)
	case "create-record":
		err = tools.RunCreateRecord(args)
	case "create-comment":
		err = tools.RunCreateComment(args)
	case "create-tags":
		err = tools.RunCreateTags(args)
	case "create-record-tags":
		err = tools.RunCreateRecordTags(args)
	case "create-custom-field":
		err = tools.RunCreateCustomField(args)
	case "create-custom-field-options":
		err = tools.RunCreateCustomFieldOptions(args)
	case "create-automation":
		err = tools.RunCreateAutomation(args)
	case "create-automation-multi":
		err = tools.RunCreateAutomationMulti(args)
	case "invite-user":
		err = tools.RunInviteUser(args)
	
	// UPDATE operations
	case "update-project":
		err = tools.RunUpdateProject(args)
	case "update-record":
		err = tools.RunUpdateRecord(args)
	case "update-comment":
		err = tools.RunUpdateComment(args)
	case "update-custom-field":
		err = tools.RunUpdateCustomField(args)
	case "update-list":
		err = tools.RunUpdateList(args)
	case "update-automation":
		err = tools.RunUpdateAutomation(args)
	case "update-automation-multi":
		err = tools.RunUpdateAutomationMulti(args)
	case "test-custom-fields":
		err = tools.RunTestCustomFields(args)
	
	// DELETE operations
	case "delete-project":
		err = tools.RunDeleteProject(args)
	case "delete-record":
		err = tools.RunDeleteRecord(args)
	case "delete-custom-field":
		err = tools.RunDeleteCustomField(args)
	case "delete-custom-field-options":
		err = tools.RunDeleteCustomFieldOptions(args)
	case "delete-list":
		err = tools.RunDeleteList(args)
	case "delete-automation":
		err = tools.RunDeleteAutomation(args)
	
	// Testing
	case "e2e":
		err = runE2E(args)
	
	// Help
	case "-h", "--help", "help":
		printUsage()
		os.Exit(0)
	default:
		fmt.Printf("Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}