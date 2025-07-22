package costcalc

// PriceFetcher interface for different cloud providers
type PriceFetcher interface {
	FetchPricing(config *PricingConfig) (*PricingResult, error)
}

// Main price fetcher factory
func NewPriceFetcher(provider CloudProvider) PriceFetcher {
	switch provider {
	case ProviderAWS:
		return NewAWSPriceFetcher()
	case ProviderGCP:
		return NewGCPPriceFetcher()
	case ProviderAzure:
		return NewAzurePriceFetcher()
	default:
		return nil
	}
}
