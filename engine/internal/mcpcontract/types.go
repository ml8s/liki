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

	GuaNumber             int
	MountainIndex         int
	XuankongYear          int
	EightGuaName          string
	MountainName          string
	StarColorName         string
	XuankongYuanName      string
	YunName               string
	XuankongSituationName string
	XuankongFuYinLayer    string
	XuankongFanYinLayer   string
	XuankongAssessment    string
	ChartDigest           string
	ElementName           string
	YinYangName           string
	DirectionName         string
	LifeGroupName         string
	YearBoundaryName      string
	FlyingStarNumber      int
	FlyingStarName        string
	YouxingName           string
	BazhaiRatingName      string
	TrigramGroupName      string
	BazhaiMatchName       string
	BazhaiBirthYear       int

	LiuyaoQiguaMode     string
	LiuyaoCastMode      string
	LiuyaoOrderName     string
	LiuyaoCoinSymbol    string
	LiuyaoCoinValue     int
	LiuyaoYaoValue      int
	LiuyaoPosition      int
	LiuyaoYaoLabel      string
	LiuyaoCastingID     string
	LiuyaoSchemaVersion string

	LiuyaoHeadsValue int
	LiuyaoTailsValue int
)
