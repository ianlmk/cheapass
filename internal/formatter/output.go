package formatter

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/ianlmk/cheapass/internal/aws"
)

func Print(costs []aws.CostBreakdown, total float64, format string) error {
	switch strings.ToLower(format) {
	case "table":
		return printTable(costs, total)
	case "json":
		return printJSON(costs, total)
	case "csv":
		return printCSV(costs, total)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}

func printTable(costs []aws.CostBreakdown, total float64) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer w.Flush()

	fmt.Fprintln(w, "SERVICE\tCOST\tUNIT")
	fmt.Fprintln(w, "---\t---\t---")

	for _, cost := range costs {
		fmt.Fprintf(w, "%s\t$%.2f\t%s\n", cost.Service, cost.Cost, cost.Unit)
	}

	fmt.Fprintln(w, "---\t---\t---")
	fmt.Fprintf(w, "%s\t$%.2f\t%s\n", "TOTAL", total, "USD")

	return nil
}

func printJSON(costs []aws.CostBreakdown, total float64) error {
	data := map[string]interface{}{
		"costs": costs,
		"total": total,
		"unit":  "USD",
	}

	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	fmt.Println(string(b))
	return nil
}

func printCSV(costs []aws.CostBreakdown, total float64) error {
	w := csv.NewWriter(os.Stdout)
	defer w.Flush()

	// Header
	w.Write([]string{"Service", "Cost", "Unit"})

	// Data
	for _, cost := range costs {
		w.Write([]string{
			cost.Service,
			fmt.Sprintf("%.2f", cost.Cost),
			cost.Unit,
		})
	}

	// Total
	w.Write([]string{"TOTAL", fmt.Sprintf("%.2f", total), "USD"})

	return nil
}
