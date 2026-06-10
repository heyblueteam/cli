package files

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/heyblueteam/cli/common"
	"github.com/spf13/cobra"
)

var inventoryCmd = &cobra.Command{
	Use:   "inventory",
	Short: "Export a CSV inventory of files",
	Long:  "Export file metadata for a workspace or the active company as CSV for later download or auditing.",
	Example: `  blue files inventory > files.csv
  blue files inventory --workspace <id-or-slug> --output files.csv
  blue files inventory --workspace <id-or-slug> --search invoice`,
	RunE: runInventory,
}

var (
	inventoryWorkspace string
	inventoryOutput    string
	inventorySearch    string
	inventoryPageSize  int
)

type inventoryFile struct {
	ID        string `json:"id"`
	UID       string `json:"uid"`
	Name      string `json:"name"`
	Size      int64  `json:"size"`
	Type      string `json:"type"`
	Extension string `json:"extension"`
	Status    string `json:"status"`
	Shared    bool   `json:"shared"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
	Project   *struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		Slug string `json:"slug"`
	} `json:"project"`
	Folder *struct {
		ID    string `json:"id"`
		Title string `json:"title"`
	} `json:"folder"`
	Todo *struct {
		ID    string `json:"id"`
		Title string `json:"title"`
	} `json:"todo"`
	User *struct {
		ID       string `json:"id"`
		FullName string `json:"fullName"`
		Email    string `json:"email"`
	} `json:"user"`
}

func init() {
	inventoryCmd.Flags().StringVarP(&inventoryWorkspace, "workspace", "w", "", "Workspace ID or slug; omit to export company files")
	inventoryCmd.Flags().StringVarP(&inventoryOutput, "output", "o", "", "Output CSV file path; omit for stdout")
	inventoryCmd.Flags().StringVar(&inventorySearch, "search", "", "Filter files by name")
	inventoryCmd.Flags().IntVar(&inventoryPageSize, "page-size", 500, "Files to fetch per API request")
}

func runInventory(cmd *cobra.Command, args []string) error {
	if inventoryPageSize <= 0 {
		return fmt.Errorf("page-size must be greater than 0")
	}

	config, err := common.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}
	client := common.NewClient(config)

	companyID, err := client.ResolveCompanyID()
	if err != nil {
		return err
	}

	projectID := ""
	if inventoryWorkspace != "" {
		client.SetProject(inventoryWorkspace)
		projectID, err = client.ResolveProjectID(inventoryWorkspace)
		if err != nil {
			return fmt.Errorf("failed to resolve workspace: %w", err)
		}
		client.SetProject(projectID)
	}

	var files []inventoryFile
	if projectID != "" {
		files, err = fetchInventoryFiles(client, companyID, projectID)
	} else {
		files, err = fetchCompanyInventoryFiles(client, companyID, config.CompanyID)
	}
	if err != nil {
		return err
	}

	var writer io.Writer = os.Stdout
	var outputFile *os.File
	if inventoryOutput != "" {
		outputFile, err = os.Create(inventoryOutput)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer outputFile.Close()
		writer = outputFile
	}

	if err := writeInventoryCSV(writer, files); err != nil {
		return err
	}

	if inventoryOutput != "" {
		fmt.Fprintf(os.Stderr, "Exported %d files to %s\n", len(files), inventoryOutput)
	}

	return nil
}

func fetchCompanyInventoryFiles(client *common.Client, companyID, companyRef string) ([]inventoryFile, error) {
	projectIDs, err := fetchInventoryProjectIDs(client, companyRef)
	if err != nil {
		return nil, err
	}

	var allFiles []inventoryFile
	for _, projectID := range projectIDs {
		client.SetProjectID(projectID)
		files, err := fetchInventoryFiles(client, companyID, projectID)
		if err != nil {
			return nil, err
		}
		allFiles = append(allFiles, files...)
	}

	return allFiles, nil
}

func fetchInventoryProjectIDs(client *common.Client, companyRef string) ([]string, error) {
	query := `query InventoryWorkspaces($companyId: String!, $skip: Int!, $take: Int!) {
		projectList(
			filter: { companyIds: [$companyId], archived: false, isTemplate: false }
			skip: $skip
			take: $take
			sort: [name_ASC]
		) {
			items { id }
			pageInfo { totalItems hasNextPage }
		}
	}`

	const pageSize = 100
	var projectIDs []string
	for skip := 0; ; skip += pageSize {
		variables := map[string]interface{}{"companyId": companyRef, "skip": skip, "take": pageSize}
		var response struct {
			ProjectList struct {
				Items []struct {
					ID string `json:"id"`
				} `json:"items"`
				PageInfo struct {
					TotalItems  int  `json:"totalItems"`
					HasNextPage bool `json:"hasNextPage"`
				} `json:"pageInfo"`
			} `json:"projectList"`
		}

		if err := client.ExecuteQueryWithResult(query, variables, &response); err != nil {
			return nil, fmt.Errorf("failed to list workspaces: %w", err)
		}
		for _, project := range response.ProjectList.Items {
			projectIDs = append(projectIDs, project.ID)
		}
		if !response.ProjectList.PageInfo.HasNextPage || len(projectIDs) >= response.ProjectList.PageInfo.TotalItems || len(response.ProjectList.Items) == 0 {
			break
		}
	}

	return projectIDs, nil
}

func fetchInventoryFiles(client *common.Client, companyID, projectID string) ([]inventoryFile, error) {
	query := `query FileInventory($filter: FileFilterInput, $sort: [FileSort!], $skip: Int, $take: Int) {
		files(filter: $filter, sort: $sort, skip: $skip, take: $take) {
			items {
				id uid name size type extension status shared createdAt updatedAt
				project { id name slug }
				folder { id title }
				todo { id title }
				user { id fullName email }
			}
			pageInfo { totalItems hasNextPage }
		}
	}`

	filter := map[string]interface{}{"companyIds": []string{companyID}, "projectIds": []string{projectID}}
	if inventorySearch != "" {
		filter["q"] = inventorySearch
	}

	var allFiles []inventoryFile
	for skip := 0; ; skip += inventoryPageSize {
		variables := map[string]interface{}{
			"filter": filter,
			"sort":   []string{"createdAt_DESC"},
			"skip":   skip,
			"take":   inventoryPageSize,
		}

		var response struct {
			Files struct {
				Items    []inventoryFile `json:"items"`
				PageInfo struct {
					TotalItems  int  `json:"totalItems"`
					HasNextPage bool `json:"hasNextPage"`
				} `json:"pageInfo"`
			} `json:"files"`
		}

		if err := client.ExecuteQueryWithResult(query, variables, &response); err != nil {
			return nil, fmt.Errorf("failed to fetch files: %w", err)
		}

		allFiles = append(allFiles, response.Files.Items...)
		if !response.Files.PageInfo.HasNextPage || len(allFiles) >= response.Files.PageInfo.TotalItems || len(response.Files.Items) == 0 {
			break
		}
	}

	return allFiles, nil
}

func writeInventoryCSV(writer io.Writer, files []inventoryFile) error {
	csvWriter := csv.NewWriter(writer)
	if err := csvWriter.Write([]string{
		"id", "uid", "name", "size", "type", "extension", "status", "shared", "created_at", "updated_at",
		"workspace_id", "workspace_name", "workspace_slug", "folder_id", "folder_name", "record_id", "record_title",
		"uploaded_by_id", "uploaded_by_name", "uploaded_by_email", "download_url",
	}); err != nil {
		return err
	}

	for _, file := range files {
		projectID, projectName, projectSlug := "", "", ""
		if file.Project != nil {
			projectID = file.Project.ID
			projectName = file.Project.Name
			projectSlug = file.Project.Slug
		}
		folderID, folderName := "", ""
		if file.Folder != nil {
			folderID = file.Folder.ID
			folderName = file.Folder.Title
		}
		recordID, recordTitle := "", ""
		if file.Todo != nil {
			recordID = file.Todo.ID
			recordTitle = file.Todo.Title
		}
		userID, userName, userEmail := "", "", ""
		if file.User != nil {
			userID = file.User.ID
			userName = file.User.FullName
			userEmail = file.User.Email
		}

		if err := csvWriter.Write([]string{
			file.ID,
			file.UID,
			file.Name,
			strconv.FormatInt(file.Size, 10),
			file.Type,
			file.Extension,
			file.Status,
			strconv.FormatBool(file.Shared),
			file.CreatedAt,
			file.UpdatedAt,
			projectID,
			projectName,
			projectSlug,
			folderID,
			folderName,
			recordID,
			recordTitle,
			userID,
			userName,
			userEmail,
			fmt.Sprintf("https://api.blue.cc/uploads/%s", file.UID),
		}); err != nil {
			return err
		}
	}

	csvWriter.Flush()
	return csvWriter.Error()
}
