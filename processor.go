package telex

type Processor interface {

	// Apply the filter to the given metric.
	Apply(in ...Metric) []Metric
}
