# cheapass

AWS cost tracker for projects. Query AWS Cost Explorer via CLI and filter spending by project and environment tags.

## Features

- 📊 Query AWS Cost Explorer API
- 🏷️ Filter by project + environment tags
- 📈 Flexible time ranges (7, 30, 90 days)
- 📁 Multiple output formats (table, JSON, CSV)
- ⚡ Single binary, no dependencies at runtime
- 🔐 Uses AWS SDK config (respects `AWS_PROFILE`, credentials file, etc.)

## Installation

### From Source

```bash
git clone https://github.com/ianlmk/cheapass.git
cd cheapass
make install
```

Or:

```bash
go install github.com/ianlmk/cheapass/cmd/cheapass@latest
```

### Build Locally

```bash
make build
./cheapass cost --help
```

## Usage

### Basic Query (Last 7 Days)

```bash
cheapass cost
```

### Specify Project and Environment

```bash
cheapass cost --project ghost --env free-tier
```

### Query Different Time Ranges

```bash
cheapass cost --days 30   # Last 30 days
cheapass cost --days 90   # Last 90 days
cheapass cost --days 7    # Last 7 days (default)
```

### Different Output Formats

```bash
cheapass cost --format table   # Human-readable table (default)
cheapass cost --format json    # JSON for scripting
cheapass cost --format csv     # CSV for spreadsheets
```

### Full Example

```bash
cheapass cost \
  --project ghost \
  --env free-tier \
  --days 30 \
  --format json
```

## AWS Permissions

The `opentofu` IAM user needs the following permissions to use `cheapass`:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "CostExplorerRead",
      "Effect": "Allow",
      "Action": [
        "ce:GetCostAndUsage",
        "ce:GetCostForecast"
      ],
      "Resource": "*"
    }
  ]
}
```

Add this to your OpenTofu infrastructure code:

```hcl
user_policies = {
  cheapass_cost_explorer = {
    user        = "opentofu"
    policy_name = "cheapass-cost-explorer"
    policy      = jsonencode({
      Version = "2012-10-17"
      Statement = [
        {
          Sid    = "CostExplorerRead"
          Effect = "Allow"
          Action = [
            "ce:GetCostAndUsage",
            "ce:GetCostForecast"
          ]
          Resource = "*"
        }
      ]
    })
  }
}
```

## AWS Configuration

`cheapass` uses the standard AWS SDK config chain:

1. Environment variables (`AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, etc.)
2. `~/.aws/credentials` and `~/.aws/config` files
3. IAM instance role (if running on EC2)
4. Container credentials (if running in ECS/Fargate)

### Using with Named Profiles

```bash
AWS_PROFILE=opentofu cheapass cost --days 30
```

Or set in `~/.aws/config`:

```ini
[profile opentofu]
region = us-east-2
output = json
```

## Development

### Build

```bash
make build
```

### Test

```bash
make test
```

### Lint

```bash
make lint
```

### Clean

```bash
make clean
```

## Project Structure

```
cheapass/
├── cmd/
│   └── cheapass/
│       └── main.go          # Entry point
├── internal/
│   ├── cmd/                 # CLI commands (Cobra)
│   │   ├── root.go
│   │   └── cost.go
│   ├── aws/                 # AWS SDK wrapper
│   │   └── client.go
│   └── formatter/           # Output formatting
│       └── output.go
├── go.mod                   # Dependencies
├── go.sum
├── Makefile
├── README.md
└── LICENSE
```

## Architecture

**Minimal, efficient design:**
- Single binary (no runtime dependencies)
- Direct AWS SDK calls (no middleware)
- Concurrent formatting (fast output)
- Standard Go error handling
- Cobra CLI for ergonomic UX

## Roadmap

- [ ] Caching layer (avoid repeated API calls)
- [ ] Anomaly detection (flag unusual spikes)
- [ ] Budget alerts (warn if approaching limits)
- [ ] Multi-project support (aggregate costs)
- [ ] Cost forecasting
- [ ] Export to S3/Slack

## License

MIT

## Author

Created for efficient AWS cost tracking on personal projects.
