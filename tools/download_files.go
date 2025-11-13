package tools

import (
	"archive/zip"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	. "demo-builder/common"

	"github.com/manifoldco/promptui"
)

const filesQuery = `
query Files($filter: FileFilterInput, $sort: [FileSort!], $skip: Int, $take: Int) {
  files(filter: $filter, sort: $sort, skip: $skip, take: $take) {
    items {
      uid
      name
      extension
    }
    pageInfo {
      totalItems
      hasNextPage
    }
  }
}
`

// RunDownloadFiles downloads files from a project and creates a zip archive
func RunDownloadFiles(args []string) error {
	fs := flag.NewFlagSet("download-files", flag.ExitOnError)
	useEnv := fs.Bool("use-env", false, "Use credentials from .env file instead of prompts")
	outputPath := fs.String("output", "", "Output path for zip file (default: blue-files-TIMESTAMP.zip)")
	parallel := fs.Int("parallel", 5, "Number of concurrent downloads (default: 5)")

	fs.Usage = func() {
		fmt.Println("Usage: download-files [flags]")
		fmt.Println()
		fmt.Println("Downloads files from a Blue project and creates a zip archive.")
		fmt.Println()
		fmt.Println("Flags:")
		fs.PrintDefaults()
		fmt.Println()
		fmt.Println("Interactive Mode (default):")
		fmt.Println("  You will be prompted for:")
		fmt.Println("  - AUTH_TOKEN: Your personal access token (labeled 'Secret' in Blue)")
		fmt.Println("  - CLIENT_ID: Your client ID (labeled 'ID' in Blue)")
		fmt.Println("  - COMPANY_ID: Your company ID")
		fmt.Println("  - PROJECT_ID: The project ID or slug")
		fmt.Println("  - FOLDER_ID: (optional) Specific folder ID, leave empty for root")
		fmt.Println()
		fmt.Println("Environment Mode (-use-env):")
		fmt.Println("  Reads from .env file in current directory")
		fmt.Println()
	}

	if err := fs.Parse(args); err != nil {
		return err
	}

	var config *Config
	var projectID string
	var folderID string
	var err error

	if *useEnv {
		// Load from .env file
		config, err = LoadConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Try to read PROJECT_ID from environment first
		projectID = os.Getenv("PROJECT_ID")
		if projectID == "" {
			// Prompt if not set
			projectID, err = promptForInput("Project ID or slug", false)
			if err != nil {
				return err
			}
		}

		// Try to read FOLDER_ID from environment first
		// Use os.LookupEnv to distinguish between unset and empty
		var folderIDSet bool
		folderID, folderIDSet = os.LookupEnv("FOLDER_ID")
		if !folderIDSet {
			// Prompt if not set
			folderID, err = promptForInput("Folder ID (optional, press Enter to skip)", true)
			if err != nil {
				return err
			}
		}
	} else {
		// Interactive mode - prompt for all values
		authToken, err := promptForInput("AUTH_TOKEN", false)
		if err != nil {
			return err
		}

		clientID, err := promptForInput("CLIENT_ID", false)
		if err != nil {
			return err
		}

		companyID, err := promptForInput("COMPANY_ID", false)
		if err != nil {
			return err
		}

		projectID, err = promptForInput("PROJECT_ID or slug", false)
		if err != nil {
			return err
		}

		folderID, err = promptForInput("FOLDER_ID (optional, press Enter to skip)", true)
		if err != nil {
			return err
		}

		config = &Config{
			APIUrl:    "https://api.blue.cc/graphql",
			AuthToken: authToken,
			ClientID:  clientID,
			CompanyID: companyID,
		}
	}

	// Create client and set project context using the unified SetProject method
	client := NewClient(config)
	client.SetProject(projectID)

	PrintInfo(fmt.Sprintf("Fetching files from project: %s", projectID))
	if folderID != "" {
		PrintInfo(fmt.Sprintf("Folder: %s", folderID))
	} else {
		PrintInfo("Folder: root")
	}

	// Fetch files
	files, err := fetchFiles(client, config.CompanyID, projectID, folderID)
	if err != nil {
		return fmt.Errorf("failed to fetch files: %w", err)
	}

	if len(files) == 0 {
		PrintInfo("No files found")
		return nil
	}

	PrintSuccess(fmt.Sprintf("Found %d file(s)", len(files)))

	// Download files and create zip
	zipPath := *outputPath
	if zipPath == "" {
		timestamp := time.Now().Format("20060102-150405")
		zipPath = fmt.Sprintf("blue-files-%s.zip", timestamp)
	}

	err = downloadAndZipFiles(client, files, zipPath, *parallel)
	if err != nil {
		return fmt.Errorf("failed to download files: %w", err)
	}

	PrintSuccess(fmt.Sprintf("Files downloaded and zipped to: %s", zipPath))
	return nil
}

// promptForInput prompts the user for input
func promptForInput(label string, allowEmpty bool) (string, error) {
	prompt := promptui.Prompt{
		Label: label,
		Validate: func(input string) error {
			if !allowEmpty && strings.TrimSpace(input) == "" {
				return fmt.Errorf("value cannot be empty")
			}
			return nil
		},
	}

	result, err := prompt.Run()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(result), nil
}

// fetchFiles fetches all files from the project/folder
func fetchFiles(client *Client, companyID, projectID, folderID string) ([]File, error) {
	variables := map[string]interface{}{
		"filter": map[string]interface{}{
			"companyIds": []string{companyID},
			"projectIds": []string{projectID},
			"folderId":   nil,
			"q":          "",
		},
		"sort": []string{"createdAt_DESC"},
		"take": 10000000, // Set high limit to fetch all files
	}

	// Set folder ID if provided
	if folderID != "" {
		variables["filter"].(map[string]interface{})["folderId"] = folderID
	}

	// Execute query (without operationName parameter - use existing CLI signature)
	data, err := client.ExecuteQuery(filesQuery, variables)
	if err != nil {
		return nil, err
	}

	// Parse response
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("error marshaling data: %w", err)
	}

	var response FilesResponse
	if err := json.Unmarshal(jsonData, &response); err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %w", err)
	}

	return response.Files.Items, nil
}

// downloadAndZipFiles downloads all files and creates a zip archive
func downloadAndZipFiles(client *Client, files []File, zipPath string, parallel int) error {
	// Validate parallel parameter
	if parallel < 1 {
		parallel = 1
	}

	// Create zip file
	zipFile, err := os.Create(zipPath)
	if err != nil {
		return fmt.Errorf("error creating zip file: %w", err)
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// Create channel for jobs and results
	type downloadJob struct {
		index int
		file  File
	}

	type downloadResult struct {
		index    int
		filename string
		data     []byte
		err      error
	}

	jobs := make(chan downloadJob, len(files))
	results := make(chan downloadResult, len(files))

	// Start worker goroutines
	var wg sync.WaitGroup
	for w := 0; w < parallel; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobs {
				PrintInfo(fmt.Sprintf("[%d/%d] Downloading: %s", job.index+1, len(files), job.file.Name))

				// Download file
				fileURL := fmt.Sprintf("https://api.blue.cc/uploads/%s", job.file.UID)
				data, err := client.DownloadFile(fileURL)

				// Determine filename with extension
				filename := job.file.Name
				if job.file.Extension != "" && !strings.HasSuffix(strings.ToLower(filename), strings.ToLower(job.file.Extension)) {
					filename = fmt.Sprintf("%s.%s", filename, job.file.Extension)
				}

				results <- downloadResult{
					index:    job.index,
					filename: filename,
					data:     data,
					err:      err,
				}
			}
		}()
	}

	// Send jobs
	for i, file := range files {
		jobs <- downloadJob{index: i, file: file}
	}
	close(jobs)

	// Wait for all downloads to complete in a separate goroutine
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results and add to zip
	// Use mutex to ensure thread-safe zip writing
	var zipMutex sync.Mutex
	successCount := 0
	errorCount := 0

	for result := range results {
		if result.err != nil {
			PrintError(fmt.Sprintf("Failed to download %s: %v", result.filename, result.err))
			errorCount++
			continue
		}

		// Thread-safe zip writing
		zipMutex.Lock()
		writer, err := zipWriter.Create(sanitizeFilename(result.filename))
		if err != nil {
			zipMutex.Unlock()
			PrintError(fmt.Sprintf("Failed to add %s to zip: %v", result.filename, err))
			errorCount++
			continue
		}

		_, err = writer.Write(result.data)
		zipMutex.Unlock()

		if err != nil {
			PrintError(fmt.Sprintf("Failed to write %s to zip: %v", result.filename, err))
			errorCount++
			continue
		}

		PrintSuccess(fmt.Sprintf("Added to zip: %s (%d bytes)", result.filename, len(result.data)))
		successCount++
	}

	PrintInfo(fmt.Sprintf("Download complete: %d succeeded, %d failed", successCount, errorCount))

	return nil
}

// sanitizeFilename removes or replaces invalid characters from filename
func sanitizeFilename(filename string) string {
	// Replace invalid characters with underscore
	invalid := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	result := filename
	for _, char := range invalid {
		result = strings.ReplaceAll(result, char, "_")
	}
	return filepath.Clean(result)
}