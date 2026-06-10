package ids

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/heyblueteam/cli/common"
	"github.com/spf13/cobra"
)

// Cmd is the parent command for quick ID lookups.
var Cmd = &cobra.Command{
	Use:     "ids",
	Aliases: []string{"id"},
	Short:   "Resolve Blue names to IDs",
	Long:    "Look up commonly needed Blue IDs for workspaces, fields, lists, tags, users, and records.",
}

type idRow struct {
	Type      string `json:"type"`
	Name      string `json:"name"`
	ID        string `json:"id"`
	UID       string `json:"uid,omitempty"`
	Extra     string `json:"extra,omitempty"`
	Workspace string `json:"workspace,omitempty"`
}

var (
	idsWorkspace string
	idsSearch    string
	idsLimit     int
	idsFormat    string
)

func init() {
	Cmd.AddCommand(workspaceCmd)
	Cmd.AddCommand(fieldCmd)
	Cmd.AddCommand(listCmd)
	Cmd.AddCommand(tagCmd)
	Cmd.AddCommand(userCmd)
	Cmd.AddCommand(recordCmd)
}

func addCommonFlags(cmd *cobra.Command, workspaceRequired bool) {
	if workspaceRequired {
		cmd.Flags().StringVarP(&idsWorkspace, "workspace", "w", "", "Workspace ID or slug (required)")
	} else {
		cmd.Flags().StringVarP(&idsWorkspace, "workspace", "w", "", "Workspace ID or slug")
	}
	cmd.Flags().StringVar(&idsSearch, "search", "", "Filter by name/title/email")
	cmd.Flags().IntVar(&idsLimit, "limit", 50, "Maximum rows to return")
	cmd.Flags().StringVar(&idsFormat, "format", "text", "Output format: text, json, csv")
}

func newClient() (*common.Client, error) {
	config, err := common.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}
	return common.NewClient(config), nil
}

func requireWorkspace() error {
	if idsWorkspace == "" {
		return fmt.Errorf("workspace ID or slug is required. Use --workspace flag")
	}
	return nil
}

func matchesSearch(values ...string) bool {
	if idsSearch == "" {
		return true
	}
	needle := strings.ToLower(idsSearch)
	for _, value := range values {
		if strings.Contains(strings.ToLower(value), needle) {
			return true
		}
	}
	return false
}

func printRows(rows []idRow) error {
	switch idsFormat {
	case "text":
		if len(rows) == 0 {
			fmt.Println("No matching IDs found.")
			return nil
		}
		for _, row := range rows {
			fmt.Printf("%s\t%s\t%s", row.Type, row.Name, row.ID)
			if row.UID != "" {
				fmt.Printf("\tuid:%s", row.UID)
			}
			if row.Extra != "" {
				fmt.Printf("\t%s", row.Extra)
			}
			fmt.Println()
		}
		return nil
	case "json":
		out, err := json.MarshalIndent(rows, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(out))
		return nil
	case "csv":
		writer := csv.NewWriter(os.Stdout)
		if err := writer.Write([]string{"type", "name", "id", "uid", "extra", "workspace"}); err != nil {
			return err
		}
		for _, row := range rows {
			if err := writer.Write([]string{row.Type, row.Name, row.ID, row.UID, row.Extra, row.Workspace}); err != nil {
				return err
			}
		}
		writer.Flush()
		return writer.Error()
	default:
		return fmt.Errorf("invalid format %q. Use text, json, or csv", idsFormat)
	}
}

func limitRows(rows []idRow) []idRow {
	if idsLimit <= 0 || len(rows) <= idsLimit {
		return rows
	}
	return rows[:idsLimit]
}
