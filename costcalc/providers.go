package costcalc

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// AWS Price Fetcher
type AWSPriceFetcher struct {
	httpClient *http.Client
}

func NewAWSPriceFetcher() *AWSPriceFetcher {
	return &AWSPriceFetcher{
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

func (f *AWSPriceFetcher) FetchPricing(config *PricingConfig) (*PricingResult, error) {
	switch config.Service {
	case ServiceEC2:
		return f.fetchEC2Pricing(config)
	case ServiceFargate:
		return f.fetchFargatePricing(config)
	default:
		return nil, fmt.Errorf("unsupported AWS service: %s", config.Service)
	}
}

func (f *AWSPriceFetcher) fetchEC2Pricing(config *PricingConfig) (*PricingResult, error) {
	// For demo, we'll use approximate known pricing
	instancePricing := f.getEC2InstancePricing(config.ResourceConfig.InstanceType, config.Region)

	return &PricingResult{
		CPUPricePerSecond: instancePricing.CPUPricePerSecond,
		Currency:          "USD",
		Region:            config.Region,
		Provider:          string(config.Provider),
		Service:           string(config.Service),
		InstanceType:      config.ResourceConfig.InstanceType,
		VCPUs:             instancePricing.VCPUs,
		LastUpdated:       time.Now(),
		Source:            "AWS Pricing API (approximated)",
	}, nil
}

func (f *AWSPriceFetcher) fetchFargatePricing(config *PricingConfig) (*PricingResult, error) {
	// AWS Fargate has simpler pricing - per vCPU-second
	// Current pricing (as of 2025): ~$0.04048/vCPU-hour in us-east-1

	basePrice := 0.04048 / 3600 // Convert to per-second
	regionMultiplier := f.getAWSRegionMultiplier(config.Region)

	cpuPricePerSecond := basePrice * regionMultiplier

	vcpus := 1.0 // Default
	if config.ResourceConfig.VCPUs != nil {
		vcpus = *config.ResourceConfig.VCPUs
	}

	return &PricingResult{
		CPUPricePerSecond: cpuPricePerSecond,
		Currency:          "USD",
		Region:            config.Region,
		Provider:          string(config.Provider),
		Service:           string(config.Service),
		VCPUs:             vcpus,
		LastUpdated:       time.Now(),
		Source:            "AWS Fargate Pricing",
	}, nil
}

type instancePricing struct {
	CPUPricePerSecond float64
	VCPUs             float64
}

func (f *AWSPriceFetcher) getEC2InstancePricing(instanceType, region string) instancePricing {
	// Simplified pricing lookup - in production, fetch from AWS API
	pricing := map[string]instancePricing{
		"t3.nano":    {CPUPricePerSecond: 0.0052 / 3600, VCPUs: 2},
		"t3.micro":   {CPUPricePerSecond: 0.0104 / 3600, VCPUs: 2},
		"t3.small":   {CPUPricePerSecond: 0.0208 / 3600, VCPUs: 2},
		"t3.medium":  {CPUPricePerSecond: 0.0416 / 3600, VCPUs: 2},
		"t3.large":   {CPUPricePerSecond: 0.0832 / 3600, VCPUs: 2},
		"t3.xlarge":  {CPUPricePerSecond: 0.1664 / 3600, VCPUs: 4},
		"t3.2xlarge": {CPUPricePerSecond: 0.3328 / 3600, VCPUs: 8},
		"c5.large":   {CPUPricePerSecond: 0.085 / 3600, VCPUs: 2},
		"c5.xlarge":  {CPUPricePerSecond: 0.17 / 3600, VCPUs: 4},
		"c5.2xlarge": {CPUPricePerSecond: 0.34 / 3600, VCPUs: 8},
		"m5.large":   {CPUPricePerSecond: 0.096 / 3600, VCPUs: 2},
		"m5.xlarge":  {CPUPricePerSecond: 0.192 / 3600, VCPUs: 4},
		"m5.2xlarge": {CPUPricePerSecond: 0.384 / 3600, VCPUs: 8},
	}

	if p, exists := pricing[instanceType]; exists {
		regionMultiplier := f.getAWSRegionMultiplier(region)
		return instancePricing{
			CPUPricePerSecond: p.CPUPricePerSecond * regionMultiplier,
			VCPUs:             p.VCPUs,
		}
	}

	// Default fallback
	return instancePricing{CPUPricePerSecond: 0.0416 / 3600, VCPUs: 2}
}

func (f *AWSPriceFetcher) getAWSRegionMultiplier(region string) float64 {
	// Approximate AWS region pricing multipliers
	multipliers := map[string]float64{
		"us-east-1":      1.0,
		"us-east-2":      1.0,
		"us-west-1":      1.05,
		"us-west-2":      1.05,
		"eu-west-1":      1.08,
		"eu-central-1":   1.10,
		"ap-southeast-1": 1.15,
		"ap-northeast-1": 1.18,
	}

	if mult, exists := multipliers[region]; exists {
		return mult
	}
	return 1.1 // Default for unknown regions
}

// GCP Price Fetcher
type GCPPriceFetcher struct {
	httpClient *http.Client
}

func NewGCPPriceFetcher() *GCPPriceFetcher {
	return &GCPPriceFetcher{
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

func (f *GCPPriceFetcher) FetchPricing(config *PricingConfig) (*PricingResult, error) {
	switch config.Service {
	case ServiceComputeEngine:
		return f.fetchComputeEnginePricing(config)
	case ServiceCloudRun:
		return f.fetchCloudRunPricing(config)
	default:
		return nil, fmt.Errorf("unsupported GCP service: %s", config.Service)
	}
}

func (f *GCPPriceFetcher) fetchComputeEnginePricing(config *PricingConfig) (*PricingResult, error) {
	// GCP Compute Engine pricing
	instancePricing := f.getGCEInstancePricing(config.ResourceConfig.InstanceType, config.Region)

	return &PricingResult{
		CPUPricePerSecond: instancePricing.CPUPricePerSecond,
		Currency:          "USD",
		Region:            config.Region,
		Provider:          string(config.Provider),
		Service:           string(config.Service),
		InstanceType:      config.ResourceConfig.InstanceType,
		VCPUs:             instancePricing.VCPUs,
		LastUpdated:       time.Now(),
		Source:            "GCP Compute Engine Pricing",
	}, nil
}

func (f *GCPPriceFetcher) fetchCloudRunPricing(config *PricingConfig) (*PricingResult, error) {
	// GCP Cloud Run pricing: ~$0.0864/vCPU-hour
	basePrice := 0.0864 / 3600 // Convert to per-second
	regionMultiplier := f.getGCPRegionMultiplier(config.Region)

	cpuPricePerSecond := basePrice * regionMultiplier

	vcpus := 1.0
	if config.ResourceConfig.VCPUs != nil {
		vcpus = *config.ResourceConfig.VCPUs
	}

	return &PricingResult{
		CPUPricePerSecond: cpuPricePerSecond,
		Currency:          "USD",
		Region:            config.Region,
		Provider:          string(config.Provider),
		Service:           string(config.Service),
		VCPUs:             vcpus,
		LastUpdated:       time.Now(),
		Source:            "GCP Cloud Run Pricing",
	}, nil
}

func (f *GCPPriceFetcher) getGCEInstancePricing(instanceType, region string) instancePricing {
	// GCP instance pricing (simplified)
	pricing := map[string]instancePricing{
		"e2-micro":      {CPUPricePerSecond: 0.008468 / 3600, VCPUs: 1},
		"e2-small":      {CPUPricePerSecond: 0.016936 / 3600, VCPUs: 1},
		"e2-medium":     {CPUPricePerSecond: 0.033872 / 3600, VCPUs: 1},
		"e2-standard-2": {CPUPricePerSecond: 0.067744 / 3600, VCPUs: 2},
		"e2-standard-4": {CPUPricePerSecond: 0.135488 / 3600, VCPUs: 4},
		"n1-standard-1": {CPUPricePerSecond: 0.0475 / 3600, VCPUs: 1},
		"n1-standard-2": {CPUPricePerSecond: 0.095 / 3600, VCPUs: 2},
		"n1-standard-4": {CPUPricePerSecond: 0.19 / 3600, VCPUs: 4},
	}

	if p, exists := pricing[instanceType]; exists {
		regionMultiplier := f.getGCPRegionMultiplier(region)
		return instancePricing{
			CPUPricePerSecond: p.CPUPricePerSecond * regionMultiplier,
			VCPUs:             p.VCPUs,
		}
	}

	return instancePricing{CPUPricePerSecond: 0.067744 / 3600, VCPUs: 2}
}

func (f *GCPPriceFetcher) getGCPRegionMultiplier(region string) float64 {
	// GCP region pricing multipliers
	multipliers := map[string]float64{
		"us-central1":     1.0,
		"us-west1":        1.0,
		"us-east1":        1.0,
		"europe-west1":    1.08,
		"europe-west4":    1.08,
		"asia-southeast1": 1.15,
		"asia-northeast1": 1.18,
	}

	if mult, exists := multipliers[region]; exists {
		return mult
	}
	return 1.1
}

// Azure Price Fetcher (using Retail Prices API - easiest to use)
type AzurePriceFetcher struct {
	httpClient *http.Client
}

func NewAzurePriceFetcher() *AzurePriceFetcher {
	return &AzurePriceFetcher{
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

type AzurePriceResponse struct {
	Items []AzurePriceItem `json:"Items"`
}

type AzurePriceItem struct {
	RetailPrice   float64 `json:"retailPrice"`
	UnitPrice     float64 `json:"unitPrice"`
	ArmRegionName string  `json:"armRegionName"`
	ProductName   string  `json:"productName"`
	SkuName       string  `json:"skuName"`
	ServiceName   string  `json:"serviceName"`
	UnitOfMeasure string  `json:"unitOfMeasure"`
}

func (f *AzurePriceFetcher) FetchPricing(config *PricingConfig) (*PricingResult, error) {
	switch config.Service {
	case ServiceVirtualMachines:
		return f.fetchVMPricing(config)
	case ServiceContainerInstances:
		return f.fetchContainerInstancesPricing(config)
	default:
		return nil, fmt.Errorf("unsupported Azure service: %s", config.Service)
	}
}

func (f *AzurePriceFetcher) fetchVMPricing(config *PricingConfig) (*PricingResult, error) {
	// Azure Retail Prices API
	serviceName := "Virtual Machines"
	skuName := config.ResourceConfig.InstanceType
	region := config.Region

	url := fmt.Sprintf("https://prices.azure.com/api/retail/prices?$filter=serviceName eq '%s' and skuName eq '%s' and armRegionName eq '%s'",
		serviceName, skuName, region)

	resp, err := f.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Azure pricing: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var priceResp AzurePriceResponse
	if err := json.Unmarshal(body, &priceResp); err != nil {
		return nil, fmt.Errorf("failed to parse pricing response: %w", err)
	}

	if len(priceResp.Items) == 0 {
		return nil, fmt.Errorf("no pricing found for %s in %s", skuName, region)
	}

	// Take the first price result
	item := priceResp.Items[0]
	pricePerHour := item.UnitPrice

	// Estimate vCPUs from instance type (simplified)
	vcpus := f.estimateVCPUsFromInstanceType(skuName)

	return &PricingResult{
		CPUPricePerSecond: pricePerHour / 3600, // Convert to per-second
		Currency:          "USD",
		Region:            config.Region,
		Provider:          string(config.Provider),
		Service:           string(config.Service),
		InstanceType:      config.ResourceConfig.InstanceType,
		VCPUs:             vcpus,
		LastUpdated:       time.Now(),
		Source:            "Azure Retail Prices API",
	}, nil
}

func (f *AzurePriceFetcher) fetchContainerInstancesPricing(config *PricingConfig) (*PricingResult, error) {
	// Azure Container Instances pricing
	url := fmt.Sprintf("https://prices.azure.com/api/retail/prices?$filter=serviceName eq 'Container Instances' and armRegionName eq '%s'", config.Region)

	resp, err := f.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Azure Container Instances pricing: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var priceResp AzurePriceResponse
	if err := json.Unmarshal(body, &priceResp); err != nil {
		return nil, fmt.Errorf("failed to parse pricing response: %w", err)
	}

	// Find vCPU pricing
	var cpuPrice float64
	for _, item := range priceResp.Items {
		if strings.Contains(item.SkuName, "vCPU") && item.UnitOfMeasure == "1 Hour" {
			cpuPrice = item.UnitPrice / 3600 // Convert to per-second
			break
		}
	}

	if cpuPrice == 0 {
		return nil, fmt.Errorf("no CPU pricing found for Container Instances in %s", config.Region)
	}

	vcpus := 1.0
	if config.ResourceConfig.VCPUs != nil {
		vcpus = *config.ResourceConfig.VCPUs
	}

	return &PricingResult{
		CPUPricePerSecond: cpuPrice,
		Currency:          "USD",
		Region:            config.Region,
		Provider:          string(config.Provider),
		Service:           string(config.Service),
		VCPUs:             vcpus,
		LastUpdated:       time.Now(),
		Source:            "Azure Container Instances Pricing",
	}, nil
}

func (f *AzurePriceFetcher) estimateVCPUsFromInstanceType(instanceType string) float64 {
	// Simple mapping of Azure instance types to vCPUs
	vcpuMap := map[string]float64{
		"Standard_B1s":    1,
		"Standard_B1ms":   1,
		"Standard_B2s":    2,
		"Standard_B2ms":   2,
		"Standard_B4ms":   4,
		"Standard_D2s_v3": 2,
		"Standard_D4s_v3": 4,
		"Standard_D8s_v3": 8,
	}

	if vcpu, exists := vcpuMap[instanceType]; exists {
		return vcpu
	}

	return 2.0 // Default assumption
}
