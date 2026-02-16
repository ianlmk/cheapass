package cmd

import (
	"github.com/spf13/cobra"
)

var (
	project     string
	environment string
	days        int
	format      string
)

var rootCmd = &cobra.Command{
	Use:   "cheapass",
	Short: "AWS spend tracker for projects",
	Long: `cheapass is a CLI tool to track AWS costs for your projects.

Query AWS Cost Explorer and filter by tags (project, environment) to see spending.`,
	Version: "0.1.0",
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(costCmd)
	
	costCmd.Flags().StringVar(&project, "project", "ghost", "Project name (tag filter)")
	costCmd.Flags().StringVar(&environment, "env", "free-tier", "Environment (tag filter)")
	costCmd.Flags().IntVar(&days, "days", 7, "Number of days back to query (7, 30, 90)")
	costCmd.Flags().StringVar(&format, "format", "table", "Output format: table, json, csv")
}
