package mcpcontract

// Wire-level scalar types carry JSON Schema constraints through TypeSchemas.
// Their runtime JSON representations remain ordinary strings and integers.
type (
	RFC3339Time  string
	DateString   string
	GenderName   string
	CityName     string
	HuangliEvent string
	HuangliCount int

	GuaNumber        int
	EightGuaName     string
	ElementName      string
	YinYangName      string
	DirectionName    string
	LifeGroupName    string
	YearBoundaryName string
	FlyingStarNumber int
	FlyingStarName   string
	YouxingName      string
	BazhaiRatingName string
	TrigramGroupName string
	BazhaiMatchName  string
	BazhaiBirthYear  int
)
