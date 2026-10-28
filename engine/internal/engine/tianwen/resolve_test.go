package tianwen

import (
	"strings"
	"testing"
	"time"
)

func hasReason(candidate ResolveCandidate, domain, detailPart string) bool {
	for _, reason := range candidate.Reasons {
		if reason.Domain != domain {
			continue
		}
		if detailPart == "" || strings.Contains(reason.Detail, detailPart) {
			return true
		}
	}
	return false
}

func inputCandidate(t *testing.T, result ResolveResult) ResolveCandidate {
	t.Helper()
	for _, candidate := range result.Candidates {
		if hasReason(candidate, reasonInput, "") {
			return candidate
		}
	}
	t.Fatalf("no input candidate: %+v", result.Candidates)
	return ResolveCandidate{}
}

func TestResolve_ExactTime(t *testing.T) {
	result, err := Resolve(ResolveInput{
		Calendar: "solar", Date: "1992-05-20", Timezone: "Asia/Shanghai", Time: "12:00", Longitude: 116.4,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Candidates) != 1 {
		t.Fatalf("candidates=%d, want 1", len(result.Candidates))
	}
	input := inputCandidate(t, result)
	if input.StandardTime.IsZero() || !hasReason(input, reasonInput, "输入时间") {
		t.Fatalf("input candidate = %+v", input)
	}
}

func TestResolve_TimePrecisionCrossesSolarTerm(t *testing.T) {
	// 2024 立春 ≈ 2024-02-04 16:27 CST；HH:MM 的分钟窗口跨交节。
	result, err := Resolve(ResolveInput{
		Calendar: "solar", Date: "2024-02-04", Timezone: "Asia/Shanghai", Time: "16:27", Longitude: 116.4,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Candidates) != 2 {
		t.Fatalf("candidates=%d, want 2", len(result.Candidates))
	}
	input := inputCandidate(t, result)
	if !strings.Contains(input.Reasons[0].Detail, "交节") {
		t.Errorf("input detail=%q, want 距交节提示", input.Reasons[0].Detail)
	}
	found := false
	for _, candidate := range result.Candidates {
		if hasReason(candidate, reasonSolarTerm, "立春后") {
			found = true
		}
	}
	if !found {
		t.Errorf("no 交节/立春后 candidate: %+v", result.Candidates)
	}
}

func TestResolve_ShichenAcrossSolarTerm(t *testing.T) {
	result, err := Resolve(ResolveInput{
		Calendar: "solar", Date: "2024-02-04", Timezone: "Asia/Shanghai", Shichen: "申", Longitude: 116.4,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Candidates) != 2 {
		t.Fatalf("candidates=%d, want 2", len(result.Candidates))
	}
	input := inputCandidate(t, result)
	if !hasReason(input, reasonInput, "交节") {
		t.Errorf("input near-boundary detail = %+v", input.Reasons)
	}
	found := false
	for _, candidate := range result.Candidates {
		if hasReason(candidate, reasonSolarTerm, "立春前") {
			found = true
		}
	}
	if !found {
		t.Errorf("no 交节/立春前 candidate: %+v", result.Candidates)
	}
}

func TestResolve_ZiShiSplitsAtSolarDay(t *testing.T) {
	result, err := Resolve(ResolveInput{
		Calendar: "solar", Date: "1990-05-20", Timezone: "Asia/Shanghai", Shichen: "子", Longitude: 116.4,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Candidates) < 2 {
		t.Fatalf("candidates=%d, want >=2", len(result.Candidates))
	}
	found := false
	for _, candidate := range result.Candidates {
		if hasReason(candidate, reasonSolarDay, "") {
			found = true
		}
	}
	if !found {
		t.Errorf("no 日界 reason: %+v", result.Candidates)
	}
}

func TestResolve_LunarInput(t *testing.T) {
	lunar := LunarDate{Year: 2020, Month: 4, Day: 12, Leap: true}
	result, err := Resolve(ResolveInput{
		Calendar:  "lunar",
		Lunar:     lunar,
		Time:      "06:00",
		Timezone:  "Asia/Shanghai",
		Longitude: 121.47,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Candidates) != 1 {
		t.Fatalf("candidates=%d, want 1", len(result.Candidates))
	}
	gt := CivilTimeFromLunarDate(lunar)
	want := time.Date(gt.Year(), gt.Month(), gt.Day(), 6, 0, 0, 0,
		time.FixedZone("CST", 8*3600))
	if got := inputCandidate(t, result).StandardTime; !got.Equal(want) {
		t.Errorf("gregorian = %s, want %s", got, want)
	}
}

func TestResolve_Range(t *testing.T) {
	result, err := Resolve(ResolveInput{
		Calendar: "solar", Date: "1992-05-20", Timezone: "Asia/Shanghai",
		Range: &ResolveRange{Start: "06:00", End: "07:00"}, Longitude: 116.4,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Candidates) != 1 || !hasReason(inputCandidate(t, result), reasonInput, "") {
		t.Fatalf("candidates=%+v", result.Candidates)
	}
}

func TestResolve_ChinaDSTFold(t *testing.T) {
	// 1988-09-11 回拨：01:00-02:00 重复，01:30 两个时刻。
	result, err := Resolve(ResolveInput{
		Calendar: "solar", Date: "1988-09-11", Timezone: "Asia/Shanghai", Time: "01:30", Longitude: 116.4,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Candidates) != 2 {
		t.Fatalf("candidates=%d, want 2", len(result.Candidates))
	}
	_ = inputCandidate(t, result)
	found := false
	for _, candidate := range result.Candidates {
		if hasReason(candidate, reasonTimeSystem, "夏令时") {
			found = true
		}
	}
	if !found {
		t.Errorf("no 时制/夏令时 candidate: %+v", result.Candidates)
	}
}

func TestResolve_ChinaDSTPeriodIsUnambiguous(t *testing.T) {
	// 1988-07-01 在夏令时期间，但 23:30 只存在于 +09 钟面；不是折叠。
	result, err := Resolve(ResolveInput{
		Calendar: "solar", Date: "1988-07-01", Timezone: "Asia/Shanghai", Time: "23:30", Longitude: 116.4,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Candidates) != 1 {
		t.Fatalf("candidates=%d, want 1", len(result.Candidates))
	}
	if _, offset := inputCandidate(t, result).StandardTime.Zone(); offset != 9*3600 {
		t.Fatalf("offset=%d, want 9h", offset)
	}
}

func TestResolve_NewYorkDSTFoldAndGap(t *testing.T) {
	fold, err := Resolve(ResolveInput{
		Calendar: "solar", Date: "2024-11-03", Timezone: "America/New_York", Time: "01:30", Longitude: -74,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(fold.Candidates) != 2 {
		t.Fatalf("fold candidates=%d, want 2", len(fold.Candidates))
	}
	found := false
	for _, candidate := range fold.Candidates {
		if hasReason(candidate, reasonTimeSystem, "夏令时") {
			found = true
		}
	}
	if !found {
		t.Errorf("fold missing 夏令时 reason: %+v", fold.Candidates)
	}

	_, err = Resolve(ResolveInput{
		Calendar: "solar", Date: "2024-03-10", Timezone: "America/New_York", Time: "02:30", Longitude: -74,
	})
	if err == nil || !strings.Contains(err.Error(), "does not exist") {
		t.Fatalf("gap err=%v, want DST gap error", err)
	}
}

func TestResolve_BerlinDSTFoldAndGap(t *testing.T) {
	fold, err := Resolve(ResolveInput{
		Calendar: "solar", Date: "2024-10-27", Timezone: "Europe/Berlin", Time: "02:30", Longitude: 13.4,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(fold.Candidates) != 2 {
		t.Fatalf("fold candidates=%d, want 2", len(fold.Candidates))
	}
	_, err = Resolve(ResolveInput{
		Calendar: "solar", Date: "2024-03-31", Timezone: "Europe/Berlin", Time: "02:30", Longitude: 13.4,
	})
	if err == nil || !strings.Contains(err.Error(), "does not exist") {
		t.Fatalf("gap err=%v, want DST gap error", err)
	}
}

func TestResolve_HistoricalTimezoneOffset(t *testing.T) {
	// 马来西亚 1982 年前使用 +07:30。
	result, err := Resolve(ResolveInput{
		Calendar: "solar", Date: "1981-06-01", Timezone: "Asia/Kuala_Lumpur", Time: "12:00", Longitude: 101.7,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Candidates) != 1 {
		t.Fatalf("candidates=%d, want 1", len(result.Candidates))
	}
	if _, offset := inputCandidate(t, result).StandardTime.Zone(); offset != 7*3600+1800 {
		t.Fatalf("offset=%d, want +07:30", offset)
	}
}

func TestResolve_RejectsTimezoneAlias(t *testing.T) {
	_, err := Resolve(ResolveInput{
		Calendar: "solar", Date: "1990-05-20", Timezone: "CST", Time: "12:00", Longitude: 116.4,
	})
	if err == nil || !strings.Contains(err.Error(), "IANA") {
		t.Fatalf("err=%v, want IANA timezone error", err)
	}
}

func TestResolve_LongitudeRangeAddsShichenCandidate(t *testing.T) {
	result, err := Resolve(ResolveInput{
		Calendar: "solar", Date: "1992-05-20", Timezone: "Asia/Shanghai",
		Shichen: "子", Longitude: 116.4,
		LongitudeRange: &LongitudeRange{West: 116.4, East: 120},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Candidates) < 2 {
		t.Fatalf("candidates=%d, want >=2", len(result.Candidates))
	}
	found := false
	for _, candidate := range result.Candidates {
		if hasReason(candidate, reasonShichen, "丑时") {
			found = true
		}
	}
	if !found {
		t.Errorf("no 时辰/丑时 candidate: %+v", result.Candidates)
	}
}

func TestResolve_InvalidLongitudeRange(t *testing.T) {
	_, err := Resolve(ResolveInput{
		Calendar: "solar", Date: "1990-05-20", Timezone: "Asia/Shanghai", Time: "12:00",
		Longitude:      116.4,
		LongitudeRange: &LongitudeRange{West: 120, East: 110},
	})
	if err == nil || !strings.Contains(err.Error(), "longitude_range") {
		t.Fatalf("err=%v, want longitude_range error", err)
	}
}

func TestResolve_LunarInputLocalAstronomical(t *testing.T) {
	ny, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatal(err)
	}
	lunar := LunarDate{Year: 2024, Month: 1, Day: 1}
	result, err := Resolve(ResolveInput{
		Calendar:  "lunar",
		Lunar:     lunar,
		Time:      "12:00",
		Timezone:  "America/New_York",
		Longitude: -74.0,
	})
	if err != nil {
		t.Fatal(err)
	}
	want := time.Date(2024, 2, 9, 12, 0, 0, 0, ny)
	got := inputCandidate(t, result).StandardTime
	if !got.Equal(want) {
		t.Fatalf("instant = %s, want %s", got.Format(time.RFC3339), want.Format(time.RFC3339))
	}
}

func TestResolve_LunarInputFixedUTC8MatchesOfficial(t *testing.T) {
	lunar := LunarDate{Year: 2025, Month: 6, Day: 15, Leap: true}
	result, err := Resolve(ResolveInput{
		Calendar:  "lunar",
		Lunar:     lunar,
		Time:      "06:00",
		Timezone:  "Etc/GMT-8",
		Longitude: 116.4,
	})
	if err != nil {
		t.Fatal(err)
	}
	gt := CivilTimeFromLunarDate(lunar)
	want := time.Date(gt.Year(), gt.Month(), gt.Day(), 6, 0, 0, 0, time.FixedZone("CST", 8*3600))
	if got := inputCandidate(t, result).StandardTime; !got.Equal(want) {
		t.Fatalf("fixed UTC+8 local_astronomical = %s, want official %s", got, want)
	}
}
