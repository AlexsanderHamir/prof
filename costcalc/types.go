package costcalc

import "time"

// CloudProvider represents supported cloud providers
type CloudProvider string

const (
	ProviderAWS   CloudProvider = "aws"
	ProviderGCP   CloudProvider = "gcp"
	ProviderAzure CloudProvider = "azure"
)

// ServiceType represents the compute service billing model
type ServiceType string

const (
	// Virtual Machine services (hourly billing)
	ServiceEC2             ServiceType = "ec2"              // AWS EC2
	ServiceComputeEngine   ServiceType = "compute-engine"   // GCP Compute Engine
	ServiceVirtualMachines ServiceType = "virtual-machines" // Azure VMs

	// Container services (per-second billing)
	ServiceFargate            ServiceType = "fargate"             // AWS Fargate
	ServiceCloudRun           ServiceType = "cloud-run"           // GCP Cloud Run
	ServiceContainerApps      ServiceType = "container-apps"      // Azure Container Apps
	ServiceContainerInstances ServiceType = "container-instances" // Azure Container Instances
)

// PricingConfig is what users need to provide for CPU cost calculation
type PricingConfig struct {
	Provider CloudProvider `json:"provider"` // aws, gcp, azure
	Service  ServiceType   `json:"service"`  // fargate, ec2, etc.
	Region   string        `json:"region"`   // us-west-2, europe-west1, etc.

	// Resource specification
	ResourceConfig ResourceConfig `json:"resource_config"`
}

// ResourceConfig specifies the compute resources
type ResourceConfig struct {
	// For VMs: Instance Type
	InstanceType string `json:"instance_type,omitempty"` // t3.medium, e2-standard-2, Standard_D2s_v3

	// For Containers: CPU Specs
	VCPUs *float64 `json:"vcpus,omitempty"` // 0.25, 0.5, 1, 2, 4, etc.
}

// PricingResult contains the fetched pricing information
type PricingResult struct {
	CPUPricePerSecond float64   `json:"cpu_price_per_second"`
	Currency          string    `json:"currency"`
	Region            string    `json:"region"`
	Provider          string    `json:"provider"`
	Service           string    `json:"service"`
	InstanceType      string    `json:"instance_type,omitempty"`
	VCPUs             float64   `json:"vcpus,omitempty"`
	LastUpdated       time.Time `json:"last_updated"`
	Source            string    `json:"source"`
}
