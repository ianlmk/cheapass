# cheapass

![Build](https://github.com/ianlmk/cheapass/actions/workflows/build.yml/badge.svg)

AWS cost tracking and resource auditing tools. Lean, efficient, open-source.

**Status:** ✅ Production-ready with multi-platform builds

---

## Tools

### 1. **cheapass** - Cost Tracker CLI

Track AWS spending by project, environment, or tag.

```bash
cheapass --region us-east-2 --profile myprofile

# Output:
# Project: app1-prod
#   Compute:  $15.32
#   Database: $25.10
#   Storage:   $2.44
# Total: $42.86/month
```

**Features:**
- Filter by AWS tags (project, environment, team)
- Multiple output formats (table, JSON, CSV)
- Cost forecasting (monthly projection)
- Multi-region support

### 2. **state_check** - Resource Auditor CLI

Audit all AWS resources in your account. Detects billable/orphaned resources.

```bash
state_check --region us-east-2 --profile myprofile

# Output:
# EC2 Instances: 2 running (t3.micro, t3.small)
# RDS: 1 database (db.t3.micro)
# EBS Volumes: 3 (50GB total)
# Elastic IPs: 1 unassociated ⚠️ (costs money!)
# NAT Gateways: 0
# S3 Buckets: 4 (25GB total)
# Total: $18-25/month estimated
```

**Features:**
- Comprehensive resource scan (9 AWS services)
- Flags unassociated EIPs (explicit cost warning)
- Estimates monthly cost
- Free tier eligibility check

---

## Installation

### From Source

```bash
# Clone repo
git clone https://github.com/ianlmk/cheapass.git
cd cheapass

# Build both tools
make build-all

# Install to $GOPATH/bin
make install

# Verify
cheapass --version
state_check --version
```

### Using Docker

```bash
docker build -t cheapass .
docker run cheapass cheapass --help
```

---

## Usage

### Cheapass (Cost Tracker)

```bash
# Basic cost tracking
cheapass --region us-east-2 --profile myprofile

# Filter by tag
cheapass --region us-east-2 --tag environment:production

# Filter by multiple tags
cheapass --region us-east-2 --tag project:app1 --tag team:platform

# JSON output (for scripts)
cheapass --region us-east-2 --format json | jq '.total'

# CSV output (for Excel)
cheapass --region us-east-2 --format csv > costs.csv

# Daily forecast
cheapass --region us-east-2 --forecast 30
```

### State Check (Resource Auditor)

```bash
# Scan current region
state_check --region us-east-2 --profile myprofile

# Scan all regions
state_check --region all --profile myprofile

# JSON output
state_check --region us-east-2 --format json

# Check for cost issues
state_check --region us-east-2 --warnings-only
```

---

## Configuration

### AWS Credentials

```bash
# Via environment
export AWS_ACCESS_KEY_ID=AKIAIOSFODNN7EXAMPLE
export AWS_SECRET_ACCESS_KEY=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY
export AWS_REGION=us-east-2

# Or via AWS profile
aws configure --profile myprofile
cheapass --profile myprofile

# Or via IAM role (EC2, Lambda, ECS)
cheapass  # Auto-detects role credentials
```

### IAM Permissions Required

Minimal permissions for **cheapass**:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "ce:GetCostAndUsage",
        "ce:GetCostForecast",
        "ce:DescribeCostCategoryDefinition",
        "ce:ListCostAllocationTags"
      ],
      "Resource": "*"
    }
  ]
}
```

Minimal permissions for **state_check**:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "ec2:DescribeInstances",
        "ec2:DescribeVolumes",
        "ec2:DescribeAddresses",
        "rds:DescribeDBInstances",
        "s3:ListAllMyBuckets",
        "s3:GetBucketLocation",
        "elasticloadbalancing:DescribeLoadBalancers",
        "eks:ListClusters",
        "ecs:ListClusters",
        "lambda:ListFunctions"
      ],
      "Resource": "*"
    }
  ]
}
```

---

## CI/CD Pipeline

![Build](https://github.com/ianlmk/cheapass/actions/workflows/build.yml/badge.svg)

**Workflow:** `.github/workflows/build.yml`

Every PR and push to main:
- ✅ Go linting (golangci-lint)
- ✅ Format validation (go fmt)
- ✅ Unit tests (go test -race)
- ✅ Code coverage report
- ✅ Binary builds (Linux/macOS amd64/arm64)

On release (git tag):
- ✅ Multi-platform builds
- ✅ Release assets to GitHub Releases
- ✅ Docker image build

**Duration:** 4-5 minutes  
**Cost:** FREE (public repo, unlimited CI/CD)

---

## Architecture

```
cmd/
├── cheapass/          # Cost tracker CLI
│   └── main.go
└── state_check/       # Resource auditor CLI
    └── main.go

internal/
├── aws/
│   ├── cost.go        # Cost Explorer API
│   ├── resources.go   # Resource enumeration
│   └── audit.go       # Shared audit logic
└── output/
    ├── table.go       # Table formatter
    ├── json.go        # JSON formatter
    └── csv.go         # CSV formatter
```

---

## Design Decisions

### AWS CLI Instead of SDK

Why? **Avoid Go version constraints.**

- cheapass uses `exec` to call AWS CLI
- Doesn't require Go 1.23 (works with Go 1.21)
- Faster development iteration
- No dependency version conflicts

### Single Audit Module

**internal/aws/audit.go** shared by both tools:
- DRY principle (don't repeat audit logic)
- Consistent resource enumeration
- Easy to extend (add service = update one place)

### Lean Binaries

- **cheapass:** ~5MB
- **state_check:** ~3MB
- No bloated dependencies
- Fast startup

---

## Performance

| Tool | Scan Time | Data Fetched |
|------|-----------|--------------|
| cheapass | 2-5 sec | 1 month of cost data |
| state_check | 5-30 sec | All AWS resources in region |

---

## Examples

### Track Costs for Multiple Projects

```bash
# Project A (dev environment)
cheapass --region us-east-2 --tag project:app1 --tag environment:dev

# Project B (prod environment)
cheapass --region us-east-2 --tag project:app1 --tag environment:prod

# All projects
cheapass --region us-east-2
```

### Monthly Cost Report

```bash
# Generate CSV report
cheapass --region us-east-2 --format csv > monthly-costs.csv

# View in Excel/Google Sheets
open monthly-costs.csv
```

### Find Cost Overruns

```bash
# Audit resources
state_check --region us-east-2

# Check for unassociated EIPs (costs money!)
state_check --region us-east-2 --warnings-only

# Forecast monthly cost
cheapass --region us-east-2 --forecast 30
```

### Free Tier Monitoring

```bash
# Check what's eligible for free tier
state_check --region us-east-2

# Verify expected services are free tier
# If seeing unexpected costs, investigate:
cheapass --region us-east-2 --format json | jq '.services'
```

---

## Testing

```bash
# Run all tests
make test

# Run with coverage
make test-coverage

# Lint code
make lint

# Build locally
make build-all
```

---

## Building & Distribution

### Local Build

```bash
make build-cheapass
make build-state-check

# Binaries in ./dist/
ls -lh dist/
```

### Multi-Platform Build (Release)

```bash
# Tag release
git tag v1.0.0
git push --tags

# CI automatically:
# 1. Builds for Linux/macOS amd64/arm64
# 2. Creates tarballs
# 3. Publishes to GitHub Releases
```

### Docker Build

```bash
docker build -t cheapass:latest .
docker run cheapass cheapass --help
```

---

## Troubleshooting

### AWS Credentials Not Found

```bash
# Check AWS profile
aws sts get-caller-identity --profile myprofile

# Set explicitly
export AWS_ACCESS_KEY_ID=...
export AWS_SECRET_ACCESS_KEY=...
cheapass
```

### Permission Denied Errors

```bash
# Verify IAM permissions
aws ce get-cost-and-usage --time-period Start=2024-01-01,End=2024-01-31 --granularity MONTHLY --metrics UnblendedCost

# If fails: add ce:GetCostAndUsage permission
```

### Region Not Found

```bash
# List available regions
aws ec2 describe-regions --all-regions

# Use valid region
cheapass --region us-east-2
```

---

## Repository Links

Related Projects:
- **aws-infra:** https://github.com/ianlmk/aws-infra (Infrastructure as code)
- **ansible-core:** https://github.com/ianlmk/ansible-core (Configuration management)
- **gcp-infra:** https://github.com/ianlmk/gcp-infra (GCP infrastructure)

---

## Documentation

- Inline help: `cheapass --help`, `state_check --help`
- Go doc: `go doc ./internal/aws`
- Examples: See `cmd/*/main.go` for usage patterns

---

## Best Practices

1. **Monitor regularly**
   ```bash
   # Weekly cost check
   cheapass --region us-east-2 --forecast 30
   ```

2. **Audit for orphaned resources**
   ```bash
   # Unassociated EIPs cost $3.50/month
   state_check --region us-east-2 --warnings-only
   ```

3. **Use tags for cost allocation**
   ```bash
   # Tag everything
   cheapass --region us-east-2 --tag project:app1 --tag team:platform
   ```

4. **Set up budget alerts**
   ```bash
   # AWS Console: Budgets > Create Budget
   # Alert when spending exceeds $50/month
   ```

5. **Review free tier eligibility**
   ```bash
   state_check --region us-east-2
   # Ensure only free-tier eligible resources running
   ```

---

## Contributing

Found a bug? Want to add a new feature?

1. Fork the repo
2. Create feature branch (`git checkout -b feature/add-feature`)
3. Implement changes
4. Run tests (`make test`)
5. Commit (`git commit -am "Add feature"`)
6. Push (`git push origin feature/add-feature`)
7. Open Pull Request

**CI/CD will automatically:**
- Lint your code
- Run tests
- Report coverage
- Build binaries

---

## License

_(Add your license here)_

---

## Support

- **GitHub Issues:** https://github.com/ianlmk/cheapass/issues
- **AWS Cost Explorer:** https://console.aws.amazon.com/cost-management/
- **AWS CLI Docs:** https://docs.aws.amazon.com/cli/
