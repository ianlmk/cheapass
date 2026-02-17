package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/ianlmk/cheapass/internal/aws"
)

func main() {
	profile := flag.String("profile", "", "AWS CLI profile name to use (optional)")
	region := flag.String("region", "us-east-2", "AWS region to scan (default: us-east-2)")
	flag.Parse()

	auditor, err := aws.NewAuditor(*profile, *region)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] %v\n", err)
		os.Exit(2)
	}

	fmt.Println()
	results := auditor.Audit()

	anyFindings := false
	for _, result := range results {
		printSection(result.Service)
		if result.Error != "" {
			fmt.Printf(" - ERROR: %s\n", result.Error)
			continue
		}

		if len(result.Items) == 0 {
			fmt.Println(" - None found.")
			continue
		}

		anyFindings = true
		for _, item := range result.Items {
			fmt.Printf(" - FOUND: %s\n", item)
		}
	}

	printSection("Summary")
	if anyFindings {
		fmt.Println("Potentially billable resources were found in this region. Review the sections above.")
		os.Exit(1)
	}
	fmt.Println("No resources found by these checks (or you lack permissions for some services).")
	os.Exit(0)
}

func printSection(title string) {
	fmt.Println()
	fmt.Println("=" + repeatStr("=", len(title)-1))
	fmt.Println(title)
	fmt.Println("=" + repeatStr("=", len(title)-1))
}

func repeatStr(s string, count int) string {
	result := ""
	for i := 0; i < count; i++ {
		result += s
	}
	return result
}
