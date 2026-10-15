package mcpcontract

// Wire-level scalar types carry JSON Schema constraints through TypeSchemas.
// Their runtime JSON representations remain ordinary strings and integers.
type (
	RFC3339Time  string
	DateString   string
	CityName     string
	HuangliEvent string
	HuangliCount int
)
