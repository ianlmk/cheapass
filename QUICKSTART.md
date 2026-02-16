# Quick Start

## Build & Run

```bash
# Install Go 1.21+
# Clone repo
git clone https://github.com/ianlmk/cheapass.git
cd cheapass

# Build
make build

# Run
AWS_PROFILE=opentofu ./cheapass cost --format table
```

## Common Commands

### View costs for Ghost project (last 7 days)

```bash
AWS_PROFILE=opentofu cheapass cost --project ghost --env free-tier
```

Output:
```
SERVICE    COST      UNIT
---        ---       ---
route53    $0.42     USD
dynamodb   $0.15     USD
---        ---       ---
TOTAL      $0.57     USD
```

### Last 30 days as JSON

```bash
AWS_PROFILE=opentofu cheapass cost --days 30 --format json
```

### Export to CSV for analysis

```bash
AWS_PROFILE=opentofu cheapass cost --days 90 --format csv > costs.csv
```

### Check help

```bash
cheapass cost --help
```

## Prerequisites

1. **Go 1.21+** installed
2. **AWS credentials** configured:
   - `~/.aws/credentials` with `[opentofu]` profile, OR
   - `AWS_ACCESS_KEY_ID` + `AWS_SECRET_ACCESS_KEY` env vars
3. **IAM permissions**: The `opentofu` user needs `ce:GetCostAndUsage` permission

## Troubleshooting

### "InvalidParameterValueException"

Likely cause: Tags don't match your infrastructure. Verify:
```bash
AWS_PROFILE=opentofu cheapass cost --project ghost --env free-tier
```

Check your actual tags in AWS Console → Cost Explorer.

### "AccessDenied"

The `opentofu` IAM user needs Cost Explorer permissions. Add to your terraform:

```hcl
user_policies = {
  cost_explorer = {
    user        = "opentofu"
    policy_name = "cheapass-ce-read"
    policy      = jsonencode({
      Version = "2012-10-17"
      Statement = [
        {
          Effect = "Allow"
          Action = ["ce:GetCostAndUsage"]
          Resource = "*"
        }
      ]
    })
  }
}
```

Then apply: `tofu apply`

### "No data available"

Cost Explorer data can take 24 hours to appear. Check again tomorrow, or try a longer time range:

```bash
cheapass cost --days 90
```

## Next Steps

- Add to CI/CD for automated cost reports
- Create a cron job to track spending over time
- Set up budgets in AWS to alert on overspend
