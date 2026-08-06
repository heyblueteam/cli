package charts

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

type editOptions struct {
	chart, title, displayType, display, currency, input, format string
	precision                                                   float64
	width, height                                               int
}

var editCmd = newEditCommand()

func newEditCommand() *cobra.Command {
	o := &editOptions{width: -1, height: -1}
	cmd := &cobra.Command{Use: "edit", Short: "Edit a chart", Example: `  blue charts edit --chart <id> --title "New title" --display-type line
  blue charts edit --input edit.json --format json`, RunE: func(cmd *cobra.Command, args []string) error { return runEdit(cmd, o) }}
	f := cmd.Flags()
	f.StringVar(&o.chart, "chart", "", "Chart ID")
	f.StringVar(&o.title, "title", "", "New title")
	f.StringVar(&o.displayType, "display-type", "", "New display type")
	f.StringVar(&o.display, "display", "", "New number format")
	f.StringVar(&o.currency, "currency", "USD", "Currency code")
	f.Float64Var(&o.precision, "precision", 0, "Decimal precision")
	f.IntVar(&o.width, "width", -1, "Card width in grid columns")
	f.IntVar(&o.height, "height", -1, "Card height in grid rows")
	f.StringVar(&o.input, "input", "", "Exact EditChartInput JSON file, or - for stdin")
	f.StringVar(&o.format, "format", "", "Output format (json)")
	return cmd
}

func runEdit(cmd *cobra.Command, o *editOptions) error {
	var input map[string]interface{}
	var err error
	if o.input != "" {
		for _, name := range []string{"chart", "title", "display-type", "display", "currency", "precision", "width", "height"} {
			if cmd.Flags().Changed(name) {
				return fmt.Errorf("--input cannot be combined with --%s", name)
			}
		}
		input, err = loadJSONInput(o.input)
		if err != nil {
			return err
		}
	} else {
		if o.chart == "" {
			return fmt.Errorf("chart ID is required. Use --chart or --input")
		}
		input = map[string]interface{}{"id": o.chart}
		if cmd.Flags().Changed("title") {
			input["title"] = o.title
		}
		if cmd.Flags().Changed("display-type") {
			if _, _, err := resolveDisplayAndChartType(o.displayType, ""); err != nil {
				return err
			}
			input["displayType"] = strings.ToLower(o.displayType)
		}
		if !cmd.Flags().Changed("display") && (cmd.Flags().Changed("currency") || cmd.Flags().Changed("precision")) {
			return fmt.Errorf("--currency and --precision require --display")
		}
		if cmd.Flags().Changed("display") {
			input["display"] = displayInput(o.display, o.currency, o.precision)
		}
		if cmd.Flags().Changed("width") != cmd.Flags().Changed("height") {
			return fmt.Errorf("--width and --height must be sent together")
		}
		if cmd.Flags().Changed("width") && cmd.Flags().Changed("height") {
			if o.width < 1 || o.width > 4 {
				return fmt.Errorf("--width must be between 1 and 4")
			}
			if o.height < 1 || o.height > 8 {
				return fmt.Errorf("--height must be between 1 and 8")
			}
			input["width"], input["height"] = o.width, o.height
		}
		if len(input) == 1 {
			return fmt.Errorf("nothing to edit")
		}
	}
	client, err := chartClient()
	if err != nil {
		return err
	}
	query := fmt.Sprintf(`mutation EditChart($input:EditChartInput!){editChart(input:$input){%s}}`, chartFields)
	var response struct {
		Chart Chart `json:"editChart"`
	}
	if err := client.ExecuteQueryWithResult(query, map[string]interface{}{"input": input}, &response); err != nil {
		return fmt.Errorf("failed to edit chart: %w", err)
	}
	if strings.EqualFold(o.format, "json") {
		return printJSON(response.Chart)
	}
	fmt.Println("Chart updated")
	printChartSummary(response.Chart)
	return nil
}
