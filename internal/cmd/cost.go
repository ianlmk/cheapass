package cmd

import (
	"fmt"
	"time"

	"github.com/ianlmk/cheapass/internal/aws"
	"github.com/ianlmk/cheapass/internal/formatter"
	"github.com/spf13/cobra"
)

var costCmd = &cobra.Command{
	Use:   "cost",
	Short: "Get AWS spending for a project",
	Long:  `Query AWS Cost Explorer API to get project spending over a time period.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Create AWS client
		client, err := aws.NewClient()
		if err != nil {
			return fmt.Errorf("failed to create AWS client: %w", err)
		}

		// Validate days parameter
		if days != 7 && days != 30 && days != 90 {
			return fmt.Errorf("invalid days: must be 7, 30, or 90")
		}

		// Calculate date range
		endDate := time.Now().UTC()
		startDate := endDate.AddDate(0, 0, -days)

		// Query costs
		costs, totalCost, err := client.GetProjectCosts(startDate, endDate, project, environment)
		if err != nil {
			return fmt.Errorf("failed to get costs: %w", err)
		}

		// Format and print output
		if err := formatter.Print(costs, totalCost, format); err != nil {
			return fmt.Errorf("failed to format output: %w", err)
		}

		return nil
	},
}
