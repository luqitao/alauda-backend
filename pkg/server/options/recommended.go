package options

// RecommendedOptions recommended server options
type RecommendedOptions struct {
	*Options
}

// NewRecommendedOptions constructor for recommended options used on the server
func NewRecommendedOptions() *RecommendedOptions {
	return &RecommendedOptions{
		Options: With(
			NewLogOptions(),
			NewKlogOptions(),
			NewInsecureServingOptions(),
			NewClientOptions(),
			NewDebugOptions(),
			NewMetricsOptions(),
			NewAPIRegistryOptions(),
			NewOpenAPIOptions(),
			NewErrorOptions(),
			NewAuditOptions(),
		),
	}
}

// Unshift add optioners to the beginning of the options slice
func (o *RecommendedOptions) Unshift(opts ...Optioner) *RecommendedOptions {
	o.Options.Unshift(opts...)
	return o
}

// Add add new options to the recommended options
func (o *RecommendedOptions) Add(opts ...Optioner) *RecommendedOptions {
	o.Options.Add(opts...)
	return o
}
