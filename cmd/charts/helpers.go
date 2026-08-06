package charts

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/heyblueteam/cli/common"
)

const filterFields = `assigneeIds unassigned dueStart dueEnd showCompleted projectIds q tagIds tagColors tagTitles todoListIds todoListTitles fields op groups groupLinks notAssigneeIds notTagIds notTodoListIds notProjectIds colors notColors hasTag hasColor hasDueDate hasDescription hasChecklist hasDependency hasReference createdStart createdEnd completedStart completedEnd updatedAt_gt updatedAt_gte recordName recordNameOp lastUpdatedByUserIds lastUpdatedByAutomationIds lastUpdatedByActorTypes`

const chartFields = `
 id title position type displayType width height isCalculating isCalculatingWithFilter needCalculation isOverBudget createdAt updatedAt
 display { type precision function currency { code name } }
 metadata {
  query {
   dimensions { title type interval customFieldName customFieldType customFieldReferenceProjectId }
   metrics { key title function customFieldName customFieldType customFieldReferenceProjectId color axis filter { ` + filterFields + ` } }
   breakout { title type customFieldName customFieldType customFieldReferenceProjectId }
   filters { ` + filterFields + ` }
  }
  presentation { stackMode direction target { mode value segmentUid } bands { atRisk onTrack } context { trend { dateField period } sparkline targetLabel showScope } }
 }
 chartSegments { id uid title seriesTitle seriesIsFold metricKey color formulaResult formula { logic { text html } display { type precision function currency { code name } } } chartSegmentValues { id uid title disabled projectId customFieldId function filter { ` + filterFields + ` } } }
`

type Chart struct {
	ID                      string                   `json:"id"`
	Title                   string                   `json:"title"`
	Position                float64                  `json:"position"`
	Type                    string                   `json:"type"`
	DisplayType             string                   `json:"displayType"`
	Width                   *int                     `json:"width"`
	Height                  *int                     `json:"height"`
	IsCalculating           bool                     `json:"isCalculating"`
	IsCalculatingWithFilter bool                     `json:"isCalculatingWithFilter"`
	NeedCalculation         bool                     `json:"needCalculation"`
	IsOverBudget            bool                     `json:"isOverBudget"`
	Display                 map[string]interface{}   `json:"display"`
	Metadata                map[string]interface{}   `json:"metadata"`
	ChartSegments           []map[string]interface{} `json:"chartSegments"`
	CreatedAt               string                   `json:"createdAt"`
	UpdatedAt               string                   `json:"updatedAt"`
}

func chartClient() (*common.Client, error) {
	config, err := common.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}
	return common.NewClient(config), nil
}
func splitCSV(value string) []string {
	var result []string
	for _, item := range strings.Split(value, ",") {
		if v := strings.TrimSpace(item); v != "" {
			result = append(result, v)
		}
	}
	return result
}
func resolveProjectIDs(client *common.Client, refs []string) ([]string, error) {
	ids := make([]string, 0, len(refs))
	for _, ref := range refs {
		client.SetProject(ref)
		id, err := client.ResolveProjectID(ref)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve workspace %q: %w", ref, err)
		}
		ids = append(ids, id)
	}
	return ids, nil
}
func printJSON(value interface{}) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}
func printChartSummary(c Chart) {
	fmt.Printf("ID: %s\nTitle: %s\nDisplay type: %s\n", c.ID, c.Title, c.DisplayType)
	if c.IsCalculating || c.NeedCalculation {
		fmt.Println("Status: Calculating...")
	} else if c.IsOverBudget {
		fmt.Println("Status: Over budget")
	} else {
		fmt.Println("Status: Ready")
	}
}
