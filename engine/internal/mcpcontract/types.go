package mcpcontract

// Wire-level scalar types carry JSON Schema constraints through TypeSchemas.
// Their runtime JSON representations remain ordinary strings and integers.
type (
	Instant      string
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

	LiuyaoQiguaMode        string
	LiuyaoCastMode         string
	LiuyaoOrderName        string
	LiuyaoCoinSymbol       string
	LiuyaoCoinValue        int
	LiuyaoYaoValue         int
	LiuyaoPosition         int
	LiuyaoYaoLabel         string
	LiuyaoCastingID        string
	LiuyaoOptionalPosition int
	LiuyaoDongRelationType string
	LiuyaoHexagramName     string
	LiuyaoGanName          string
	LiuyaoZhiName          string
	LiuyaoWangShuai        string
	LiuyaoShiYing          string
	LiuyaoLiuQinName       string
	LiuyaoLiuShouName      string
	LiuyaoMuKuType         string
	LiuyaoPatternType      string
	LiuyaoForceRole        string
	LiuyaoConflictID       string
	LiuyaoTombBranch       string
	LiuyaoTombSchool       string
	LiuyaoYongShenName     string
	LiuyaoTimingID         string

	LiuyaoHeadsValue int
	LiuyaoTailsValue int
)
