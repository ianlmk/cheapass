package aws

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"time"
)

type Client struct{}

type CostBreakdown struct {
	Service string
	Cost    float64
	Unit    string
}

func NewClient() (*Client, error) {
	// Verify AWS CLI is available
	cmd := exec.Command("aws", "--version")
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("AWS CLI not found: %w", err)
	}
	return &Client{}, nil
}

func (c *Client) GetProjectCosts(startDate, endDate time.Time, project, environment string) ([]CostBreakdown, float64, error) {
	start := startDate.Format("2006-01-02")
	end := endDate.Format("2006-01-02")

	// Call AWS CLI
	cmd := exec.Command("aws", "ce", "get-cost-and-usage",
		"--time-period", fmt.Sprintf("Start=%s,End=%s", start, end),
		"--granularity", "DAILY",
		"--metrics", "UnblendedCost",
		"--group-by", "Type=DIMENSION,Key=SERVICE",
		"--filter", fmt.Sprintf(`{"Tags":{"Key":"%s","Values":["%s"]}}`, project, project),
		"--output", "json",
	)

	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	if err := cmd.Run(); err != nil {
		return nil, 0, fmt.Errorf("AWS CLI query failed: %w", err)
	}

	// Parse JSON response
	var result struct {
		ResultsByTime []struct {
			Groups []struct {
				Keys   []string
				Metrics map[string]struct {
					Amount string
					Unit   string
				}
			}
		}
	}

	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		return nil, 0, fmt.Errorf("failed to parse AWS response: %w", err)
	}

	var costs []CostBreakdown
	var totalCost float64

	for _, dayResult := range result.ResultsByTime {
		for _, group := range dayResult.Groups {
			if len(group.Keys) == 0 {
				continue
			}

			metrics := group.Metrics["UnblendedCost"]
			var amount float64
			fmt.Sscanf(metrics.Amount, "%f", &amount)

			totalCost += amount
			costs = append(costs, CostBreakdown{
				Service: group.Keys[0],
				Cost:    amount,
				Unit:    metrics.Unit,
			})
		}
	}

	return costs, totalCost, nil
}
