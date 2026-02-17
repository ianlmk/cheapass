package aws

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"sort"
)

// AuditResult represents a single found resource
type AuditResult struct {
	Service string
	Items   []string
	Error   string
}

// Auditor checks AWS account for billable resources
type Auditor struct {
	profile string
	region  string
}

// NewAuditor creates an auditor for the given profile and region
func NewAuditor(profile, region string) (*Auditor, error) {
	// Verify credentials with AWS CLI
	cmd := exec.Command("aws", "sts", "get-caller-identity", "--region", region)
	if profile != "" {
		cmd.Env = append(cmd.Environ(), "AWS_PROFILE="+profile)
	}

	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("unable to verify AWS credentials: %w", err)
	}

	var identity struct {
		Account string `json:"Account"`
		Arn     string `json:"Arn"`
	}
	if err := json.Unmarshal(out.Bytes(), &identity); err != nil {
		return nil, fmt.Errorf("failed to parse STS response: %w", err)
	}

	fmt.Printf("Using account=%s arn=%s region=%s profile=%s\n",
		identity.Account, identity.Arn, region, profile)

	return &Auditor{
		profile: profile,
		region:  region,
	}, nil
}

// runAWSCommand executes an AWS CLI command and returns JSON output
func (a *Auditor) runAWSCommand(service, operation string, args ...string) ([]byte, error) {
	cmdArgs := []string{service, operation, "--region", a.region, "--output", "json"}
	cmdArgs = append(cmdArgs, args...)
	cmd := exec.Command("aws", cmdArgs...)
	if a.profile != "" {
		cmd.Env = append(cmd.Environ(), "AWS_PROFILE="+a.profile)
	}

	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("AWS CLI failed: %w\n%s", err, out.String())
	}

	return out.Bytes(), nil
}

// Audit runs all checks and returns results
func (a *Auditor) Audit() []AuditResult {
	checks := []struct {
		name string
		fn   func() ([]string, error)
	}{
		{"EC2 Instances", a.checkEC2Instances},
		{"EBS Volumes", a.checkEBSVolumes},
		{"Elastic IPs", a.checkEIPs},
		{"NAT Gateways", a.checkNATGateways},
		{"Load Balancers (ALB/NLB)", a.checkLoadBalancers},
		{"RDS DB Instances", a.checkRDSInstances},
		{"EKS Clusters", a.checkEKSClusters},
		{"ECS Services (desired/running > 0)", a.checkECSServices},
		{"Lambda Functions (presence)", a.checkLambdaFunctions},
	}

	var results []AuditResult
	for _, check := range checks {
		items, err := check.fn()
		result := AuditResult{Service: check.name}
		if err != nil {
			result.Error = err.Error()
		} else {
			result.Items = items
		}
		results = append(results, result)
	}
	return results
}

func (a *Auditor) checkEC2Instances() ([]string, error) {
	out, err := a.runAWSCommand("ec2", "describe-instances",
		"--filters", "Name=instance-state-name,Values=pending,running,stopping,stopped")
	if err != nil {
		return nil, err
	}

	var result struct {
		Reservations []struct {
			Instances []struct {
				InstanceId string
				State      struct{ Name string }
				InstanceType string
				Placement  struct{ AvailabilityZone string }
				Tags       []struct {
					Key   string
					Value string
				}
			}
		}
	}

	if err := json.Unmarshal(out, &result); err != nil {
		return nil, fmt.Errorf("failed to parse EC2 response: %w", err)
	}

	var results []string
	for _, res := range result.Reservations {
		for _, inst := range res.Instances {
			name := ""
			for _, tag := range inst.Tags {
				if tag.Key == "Name" {
					name = tag.Value
					break
				}
			}

			label := fmt.Sprintf("%s [%s] %s %s", inst.InstanceId, inst.State.Name, inst.InstanceType, inst.Placement.AvailabilityZone)
			if name != "" {
				label += fmt.Sprintf(" Name=%s", name)
			}
			results = append(results, label)
		}
	}

	return results, nil
}

func (a *Auditor) checkEBSVolumes() ([]string, error) {
	out, err := a.runAWSCommand("ec2", "describe-volumes",
		"--filters", "Name=status,Values=in-use,available")
	if err != nil {
		return nil, err
	}

	var result struct {
		Volumes []struct {
			VolumeId         string
			State            string
			Size             int
			VolumeType       string
			AvailabilityZone string
			Attachments      []interface{}
		}
	}

	if err := json.Unmarshal(out, &result); err != nil {
		return nil, fmt.Errorf("failed to parse EBS response: %w", err)
	}

	var results []string
	for _, vol := range result.Volumes {
		attached := "attached"
		if len(vol.Attachments) == 0 {
			attached = "unattached"
		}

		label := fmt.Sprintf("%s [%s/%s] %dGiB %s %s", vol.VolumeId, vol.State, attached, vol.Size, vol.VolumeType, vol.AvailabilityZone)
		results = append(results, label)
	}

	return results, nil
}

func (a *Auditor) checkEIPs() ([]string, error) {
	out, err := a.runAWSCommand("ec2", "describe-addresses")
	if err != nil {
		return nil, err
	}

	var result struct {
		Addresses []struct {
			PublicIp      string
			AssociationId string
			InstanceId    string
			NetworkInterfaceId string
		}
	}

	if err := json.Unmarshal(out, &result); err != nil {
		return nil, fmt.Errorf("failed to parse EIP response: %w", err)
	}

	var results []string
	for _, addr := range result.Addresses {
		status := "associated"
		if addr.AssociationId == "" {
			status = "UNASSOCIATED (billable)"
		}

		label := fmt.Sprintf("%s [%s] instance=%s eni=%s", addr.PublicIp, status, addr.InstanceId, addr.NetworkInterfaceId)
		results = append(results, label)
	}

	return results, nil
}

func (a *Auditor) checkNATGateways() ([]string, error) {
	out, err := a.runAWSCommand("ec2", "describe-nat-gateways")
	if err != nil {
		return nil, err
	}

	var result struct {
		NatGateways []struct {
			NatGatewayId string
			State        string
			VpcId        string
			SubnetId     string
			NatGatewayAddresses []struct {
				PublicIp string
			}
		}
	}

	if err := json.Unmarshal(out, &result); err != nil {
		return nil, fmt.Errorf("failed to parse NAT Gateway response: %w", err)
	}

	var results []string
	for _, ngw := range result.NatGateways {
		if ngw.State != "available" && ngw.State != "pending" && ngw.State != "deleting" {
			continue
		}

		eips := ""
		for _, addr := range ngw.NatGatewayAddresses {
			if eips != "" {
				eips += ","
			}
			eips += addr.PublicIp
		}
		if eips == "" {
			eips = "-"
		}

		label := fmt.Sprintf("%s [%s] vpc=%s subnet=%s eips=%s", ngw.NatGatewayId, ngw.State, ngw.VpcId, ngw.SubnetId, eips)
		results = append(results, label)
	}

	return results, nil
}

func (a *Auditor) checkLoadBalancers() ([]string, error) {
	out, err := a.runAWSCommand("elbv2", "describe-load-balancers")
	if err != nil {
		return nil, err
	}

	var result struct {
		LoadBalancers []struct {
			LoadBalancerName string
			Scheme           string
			Type             string
			LoadBalancerArn  string
			State            struct {
				Code string
			}
		}
	}

	if err := json.Unmarshal(out, &result); err != nil {
		return nil, fmt.Errorf("failed to parse Load Balancer response: %w", err)
	}

	var results []string
	for _, lb := range result.LoadBalancers {
		label := fmt.Sprintf("%s [%s] type=%s scheme=%s arn=%s", lb.LoadBalancerName, lb.State.Code, lb.Type, lb.Scheme, lb.LoadBalancerArn)
		results = append(results, label)
	}

	return results, nil
}

func (a *Auditor) checkRDSInstances() ([]string, error) {
	out, err := a.runAWSCommand("rds", "describe-db-instances")
	if err != nil {
		return nil, err
	}

	var result struct {
		DBInstances []struct {
			DBInstanceIdentifier string
			DBInstanceStatus     string
			Engine               string
			DBInstanceClass      string
			MultiAZ              bool
		}
	}

	if err := json.Unmarshal(out, &result); err != nil {
		return nil, fmt.Errorf("failed to parse RDS response: %w", err)
	}

	var results []string
	for _, db := range result.DBInstances {
		label := fmt.Sprintf("%s [%s] %s %s multiAZ=%v", db.DBInstanceIdentifier, db.DBInstanceStatus, db.Engine, db.DBInstanceClass, db.MultiAZ)
		results = append(results, label)
	}

	return results, nil
}

func (a *Auditor) checkEKSClusters() ([]string, error) {
	out, err := a.runAWSCommand("eks", "list-clusters")
	if err != nil {
		// Many accounts don't have EKS permissions; treat as non-fatal
		return nil, nil
	}

	var result struct {
		Clusters []string
	}

	if err := json.Unmarshal(out, &result); err != nil {
		return nil, fmt.Errorf("failed to parse EKS response: %w", err)
	}

	var results []string
	for _, name := range result.Clusters {
		out, err := a.runAWSCommand("eks", "describe-cluster", "--name", name)
		if err != nil {
			results = append(results, fmt.Sprintf("%s [ERROR: %v]", name, err))
			continue
		}

		var desc struct {
			Cluster struct {
				Status  string
				Version string
			}
		}

		if err := json.Unmarshal(out, &desc); err != nil {
			results = append(results, fmt.Sprintf("%s [ERROR: %v]", name, err))
			continue
		}

		label := fmt.Sprintf("%s [status=%s] version=%s", name, desc.Cluster.Status, desc.Cluster.Version)
		results = append(results, label)
	}

	return results, nil
}

func (a *Auditor) checkLambdaFunctions() ([]string, error) {
	out, err := a.runAWSCommand("lambda", "list-functions")
	if err != nil {
		return nil, err
	}

	var result struct {
		Functions []struct {
			FunctionName string
			Runtime      string
			MemorySize   int
		}
	}

	if err := json.Unmarshal(out, &result); err != nil {
		return nil, fmt.Errorf("failed to parse Lambda response: %w", err)
	}

	var results []string
	for _, fn := range result.Functions {
		label := fmt.Sprintf("%s runtime=%s memory=%dMB", fn.FunctionName, fn.Runtime, fn.MemorySize)
		results = append(results, label)
	}

	return results, nil
}

func (a *Auditor) checkECSServices() ([]string, error) {
	out, err := a.runAWSCommand("ecs", "list-clusters")
	if err != nil {
		return nil, err
	}

	var clusters struct {
		ClusterArns []string
	}

	if err := json.Unmarshal(out, &clusters); err != nil {
		return nil, fmt.Errorf("failed to parse ECS clusters response: %w", err)
	}

	var results []string
	for _, clusterArn := range clusters.ClusterArns {
		out, err := a.runAWSCommand("ecs", "list-services", "--cluster", clusterArn)
		if err != nil {
			results = append(results, fmt.Sprintf("%s [ERROR: %v]", clusterArn, err))
			continue
		}

		var services struct {
			ServiceArns []string
		}

		if err := json.Unmarshal(out, &services); err != nil {
			results = append(results, fmt.Sprintf("%s [ERROR: %v]", clusterArn, err))
			continue
		}

		if len(services.ServiceArns) == 0 {
			continue
		}

		// Batch describe services (10 at a time)
		for i := 0; i < len(services.ServiceArns); i += 10 {
			end := i + 10
			if end > len(services.ServiceArns) {
				end = len(services.ServiceArns)
			}

			batch := services.ServiceArns[i:end]
			args := []string{"--cluster", clusterArn, "--services"}
			args = append(args, batch...)
			out, err := a.runAWSCommand("ecs", "describe-services", args...)
			if err != nil {
				results = append(results, fmt.Sprintf("%s [ERROR: %v]", clusterArn, err))
				continue
			}

			var desc struct {
				Services []struct {
					ServiceName  string
					Status       string
					DesiredCount int
					RunningCount int
				}
			}

			if err := json.Unmarshal(out, &desc); err != nil {
				results = append(results, fmt.Sprintf("%s [ERROR: %v]", clusterArn, err))
				continue
			}

			for _, svc := range desc.Services {
				if svc.DesiredCount > 0 || svc.RunningCount > 0 {
					label := fmt.Sprintf("%s [%s] desired=%d running=%d cluster=%s",
						svc.ServiceName, svc.Status, svc.DesiredCount, svc.RunningCount, clusterArn)
					results = append(results, label)
				}
			}
		}
	}

	// Sort results for consistent output
	sort.Strings(results)
	return results, nil
}
