package aws

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ce"
	"github.com/aws/aws-sdk-go-v2/service/ce/types"
)

type Client struct {
	ce *ce.Client
}

type CostBreakdown struct {
	Service string
	Cost    float64
	Unit    string
}

func NewClient(ctx context.Context) (*Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK config: %w", err)
	}

	return &Client{
		ce: ce.NewFromConfig(cfg),
	}, nil
}

func (c *Client) GetProjectCosts(ctx context.Context, startDate, endDate time.Time, project, environment string) ([]CostBreakdown, float64, error) {
	// Format dates for AWS API (YYYY-MM-DD)
	start := startDate.Format("2006-01-02")
	end := endDate.Format("2006-01-02")

	// Query Cost Explorer API
	result, err := c.ce.GetCostAndUsage(ctx, &ce.GetCostAndUsageInput{
		TimePeriod: &types.DateInterval{
			Start: &start,
			End:   &end,
		},
		Granularity: types.GranularityDaily,
		Metrics:     []string{"UnblendedCost"},
		GroupBy: []types.GroupDefinition{
			{
				Type: types.GroupDefinitionTypeDimension,
				Key: &types.GroupDefinitionKeyOptions{
					Dimension: types.DimensionServiceCode,
				},
			},
		},
		Filter: &types.Expression{
			Tags: &types.TagValues{
				Key:    &project,
				Values: []string{project},
			},
		},
	})

	if err != nil {
		return nil, 0, fmt.Errorf("cost explorer query failed: %w", err)
	}

	// Parse results
	var costs []CostBreakdown
	var totalCost float64

	for _, result := range result.ResultsByTime {
		for _, group := range result.Total {
			if group.Amount == nil || group.Unit == nil {
				continue
			}

			amount := parseFloat(*group.Amount)
			totalCost += amount

			// Get service name from group keys
			serviceName := "Unknown"
			if len(result.Groups) > 0 && len(result.Groups[0].Keys) > 0 {
				serviceName = *result.Groups[0].Keys[0]
			}

			costs = append(costs, CostBreakdown{
				Service: serviceName,
				Cost:    amount,
				Unit:    *group.Unit,
			})
		}
	}

	return costs, totalCost, nil
}

func parseFloat(s string) float64 {
	var f float64
	fmt.Sscanf(s, "%f", &f)
	return f
}
