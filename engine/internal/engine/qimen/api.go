// Package qimen computes Qimen charts across hour, quarter, day, month, and
// year scopes with rotating, flying, and Golden Mirror plates.
package qimen

import (
	"fmt"
	"time"

	"liki-engine/internal/engine/ganzhi"
	"liki-engine/internal/engine/tianwen"
)

// DingjuMethod selects a stable method for determining the Qimen ju (局数).
type DingjuMethod string

const (
	DingjuChaiBu  DingjuMethod = "chaibu"
	DingjuZhiRun  DingjuMethod = "zhirun"
	DingjuMaoShan DingjuMethod = "maoshan"
	DingjuNone    DingjuMethod = "none"
)

// ParseDingjuMethod validates a dingju-method identifier.
func ParseDingjuMethod(value string) (DingjuMethod, error) {
	switch DingjuMethod(value) {
	case DingjuChaiBu, DingjuZhiRun, DingjuMaoShan, DingjuNone:
		return DingjuMethod(value), nil
	default:
		return "", fmt.Errorf(
			"unknown dingju method %q: must be chaibu, zhirun, maoshan or none", value,
		)
	}
}

func (m DingjuMethod) String() string { return string(m) }

// QuarterRule selects a named Kejia time-division rule.
type QuarterRule string

const (
	QuarterTenMinuteSanYuan        QuarterRule = "ten_minute_sanyuan"
	QuarterTwelveMinuteTenDivision QuarterRule = "twelve_minute_ten_division"
)

// ParseQuarterRule validates a Kejia rule identifier.
func ParseQuarterRule(value string) (QuarterRule, error) {
	switch QuarterRule(value) {
	case QuarterTenMinuteSanYuan, QuarterTwelveMinuteTenDivision:
		return QuarterRule(value), nil
	default:
		return "", fmt.Errorf(
			"unknown quarter rule %q: must be ten_minute_sanyuan or twelve_minute_ten_division", value,
		)
	}
}

func (r QuarterRule) String() string { return string(r) }

// Scope selects the chart's time layer and duty pillar.
type Scope string

const (
	ScopeHour    Scope = "hour"
	ScopeDay     Scope = "day"
	ScopeQuarter Scope = "quarter"
	ScopeMonth   Scope = "month"
	ScopeYear    Scope = "year"
)

// ParseScope validates a time-scope identifier.
func ParseScope(value string) (Scope, error) {
	switch Scope(value) {
	case ScopeHour, ScopeDay, ScopeQuarter, ScopeMonth, ScopeYear:
		return Scope(value), nil
	default:
		return "", fmt.Errorf("unknown scope %q: must be hour, day, quarter, month or year", value)
	}
}

func (s Scope) String() string { return string(s) }

// School selects a plate geometry and its spirit convention.
type School string

const (
	SchoolZhuanPan     School = "zhuanpan"
	SchoolLuoShuFeiPan School = "luoshu_feipan"
	SchoolMingFaFeiPan School = "mingfa_feipan"
	SchoolJinhanYuJing School = "jinhan_yujing"
)

// ParseSchool validates a plate-school identifier.
func ParseSchool(value string) (School, error) {
	switch School(value) {
	case SchoolZhuanPan, SchoolLuoShuFeiPan, SchoolMingFaFeiPan, SchoolJinhanYuJing:
		return School(value), nil
	default:
		return "", fmt.Errorf("unknown school %q: must be zhuanpan, luoshu_feipan, mingfa_feipan or jinhan_yujing", value)
	}
}

func (s School) String() string { return string(s) }

// Method is a closed Qimen method combination.
type Method struct {
	Scope        Scope
	School       School
	Dingju       DingjuMethod
	QuarterRule  QuarterRule
	BaseDingju   DingjuMethod
	DunSource    string
	HourBoundary string
}

// ParseMethod validates a method combination against the catalog.
func ParseMethod(scope, school, dingjuMethod, quarterRule string) (Method, error) {
	return ParseMethodWithOptions(scope, school, dingjuMethod, quarterRule, "", "", "")
}

// ParseMethodWithOptions validates a method and its Kejia-specific options.
func ParseMethodWithOptions(
	scope, school, dingjuMethod, quarterRule, baseDingjuMethod, dunSource, hourBoundary string,
) (Method, error) {
	parsedScope, err := ParseScope(scope)
	if err != nil {
		return Method{}, err
	}
	parsedSchool, err := ParseSchool(school)
	if err != nil {
		return Method{}, err
	}
	scopeEntry, ok := scopeMethods[parsedScope]
	if !ok {
		return Method{}, fmt.Errorf("unknown scope %q", parsedScope)
	}
	schoolEntry, ok := schoolMethods[parsedSchool]
	if !ok {
		return Method{}, fmt.Errorf("unknown school %q", parsedSchool)
	}
	if !schoolEntry.allowsScope(parsedScope) {
		return Method{}, fmt.Errorf("school %q does not support scope %q", parsedSchool, parsedScope)
	}

	result := Method{Scope: parsedScope, School: parsedSchool}
	switch parsedScope {
	case ScopeQuarter:
		if dingjuMethod != "" {
			return Method{}, fmt.Errorf("scope %q does not support dingju_method", parsedScope)
		}
		if quarterRule == "" {
			return Method{}, fmt.Errorf("scope %q requires quarter_rule", parsedScope)
		}
		rule, err := ParseQuarterRule(quarterRule)
		if err != nil {
			return Method{}, err
		}
		if !scopeEntry.allowsQuarterRule(rule) || !schoolEntry.allowsQuarterRule(rule) {
			return Method{}, fmt.Errorf("school %q does not support quarter rule %q", parsedSchool, rule)
		}
		result.QuarterRule = rule
		if rule == QuarterTwelveMinuteTenDivision {
			base := DingjuChaiBu
			if baseDingjuMethod != "" {
				parsedBase, err := ParseDingjuMethod(baseDingjuMethod)
				if err != nil {
					return Method{}, fmt.Errorf("invalid base dingju method: %w", err)
				}
				base = parsedBase
			}
			if !quarterRules[rule].BaseDingjuMethods[base] {
				return Method{}, fmt.Errorf("twelve-minute quarter does not support base dingju method %q", base)
			}
			if dunSource == "" {
				dunSource = quarterRules[rule].DefaultDunSource
			}
			if !quarterRules[rule].DunSources[dunSource] {
				return Method{}, fmt.Errorf("twelve-minute quarter does not support dun source %q", dunSource)
			}
			if hourBoundary == "" {
				hourBoundary = quarterRules[rule].DefaultHourBoundary
			}
			if _, exists := quarterRules[rule].HourBoundaries[hourBoundary]; !exists {
				return Method{}, fmt.Errorf("twelve-minute quarter does not support hour boundary %q", hourBoundary)
			}
			result.BaseDingju, result.DunSource, result.HourBoundary = base, dunSource, hourBoundary
		} else if baseDingjuMethod != "" || dunSource != "" || hourBoundary != "" {
			return Method{}, fmt.Errorf("quarter rule %q does not support Kejia options", rule)
		}
	default:
		if quarterRule != "" {
			return Method{}, fmt.Errorf("scope %q does not support quarter_rule", parsedScope)
		}
		if baseDingjuMethod != "" || dunSource != "" || hourBoundary != "" {
			return Method{}, fmt.Errorf("scope %q does not support Kejia options", parsedScope)
		}
		if scopeEntry.allowsNoDingjuSchool(parsedSchool) {
			if dingjuMethod != "" {
				return Method{}, fmt.Errorf("school %q must omit dingju_method", parsedSchool)
			}
			result.Dingju = DingjuNone
		} else if len(scopeEntry.DingjuMethods) == 0 {
			if dingjuMethod != "" {
				return Method{}, fmt.Errorf("scope %q does not support dingju_method", parsedScope)
			}
		} else {
			method, err := ParseDingjuMethod(dingjuMethod)
			if err != nil {
				return Method{}, err
			}
			if !scopeEntry.allowsDingju(method) {
				return Method{}, fmt.Errorf("scope %q does not support dingju method %q", parsedScope, method)
			}
			result.Dingju = method
		}
		if result.Dingju == DingjuNone {
			if !scopeEntry.allowsNoDingjuSchool(parsedSchool) || !schoolEntry.allowsDingju(DingjuNone) {
				return Method{}, fmt.Errorf("school %q must omit dingju_method", parsedSchool)
			}
		} else if result.Dingju != "" && !schoolEntry.allowsDingju(result.Dingju) {
			return Method{}, fmt.Errorf("school %q does not support dingju method %q", parsedSchool, result.Dingju)
		}
	}
	return result, nil
}

// ResolveMethod fills omitted API dimensions from the catalog and validates the
// resulting closed combination.
func ResolveMethod(scope, school, dingjuMethod, quarterRule string) (Method, error) {
	if scope == "" {
		scope = defaultScope.String()
	}
	if school == "" {
		school = defaultSchool.String()
	}
	parsedScope, err := ParseScope(scope)
	if err != nil {
		return Method{}, err
	}
	parsedSchool, err := ParseSchool(school)
	if err != nil {
		return Method{}, err
	}
	entry := scopeMethods[parsedScope]
	if parsedScope == ScopeQuarter {
		if quarterRule == "" {
			quarterRule = entry.DefaultQuarterRule.String()
		}
		return ParseMethod(scope, school, "", quarterRule)
	}
	if dingjuMethod == "" && entry.allowsNoDingjuSchool(parsedSchool) {
		return ParseMethod(scope, school, "", "")
	}
	if dingjuMethod == "" && entry.DefaultDingju != "" {
		dingjuMethod = entry.DefaultDingju.String()
	}
	return ParseMethod(scope, school, dingjuMethod, "")
}

func ResolveMethodWithOptions(
	scope, school, dingjuMethod, quarterRule, baseDingjuMethod, dunSource, hourBoundary string,
) (Method, error) {
	resolved, err := ResolveMethod(scope, school, dingjuMethod, quarterRule)
	if err != nil {
		return Method{}, err
	}
	dingju := resolved.Dingju.String()
	if resolved.Dingju == DingjuNone {
		dingju = ""
	}
	return ParseMethodWithOptions(
		resolved.Scope.String(), resolved.School.String(), dingju,
		resolved.QuarterRule.String(), baseDingjuMethod, dunSource, hourBoundary,
	)
}

// ComputeChart computes the default hour-scope rotating chai-bu chart.
func ComputeChart(st tianwen.SolarTime) Chart {
	method, _ := ResolveMethod("", "", "", "")
	chart, _ := ComputeChartWithMethod(st, method)
	return chart
}

// ComputeChartWithMethod computes a chart for a catalog method and rejects
// combinations absent from the catalog.
func ComputeChartWithMethod(st tianwen.SolarTime, method Method) (Chart, error) {
	parsed, err := ParseMethodWithOptions(
		method.Scope.String(), method.School.String(), method.Dingju.String(),
		method.QuarterRule.String(), method.BaseDingju.String(), method.DunSource, method.HourBoundary,
	)
	if err != nil {
		return Chart{}, fmt.Errorf("invalid qimen method: %w", err)
	}
	if parsed.School == SchoolJinhanYuJing {
		return Chart{}, fmt.Errorf("jinhan_yujing has a distinct result model; use ComputeJinhanChart")
	}
	bz, chartTime := chartInputs(st)
	return computeChart(bz, chartTime, parsed), nil
}

func chartInputs(st tianwen.SolarTime) (ganzhi.Bazi, time.Time) {
	bz := tianwen.ComputeBazi(st)
	if qimenDayBoundary == "late_zi_rolls_day_pillar" && st.Time().Hour() >= 23 {
		dayIndex := ganzhi.SixtyCycleIndex(bz.Ri.Gan, bz.Ri.Zhi) + 1
		bz.Ri = ganzhi.SixtyToZhu(dayIndex)
	}
	return bz, st.Time()
}

// ComputeChartWithYongShenAndMethod aggregates selected Yong Shen under a
// catalog method and rejects combinations absent from the catalog.
func ComputeChartWithYongShenAndMethod(
	st tianwen.SolarTime, syms []YongShenSymbol, birthDate BirthDate, method Method,
) (Chart, error) {
	parsed, err := ParseMethodWithOptions(
		method.Scope.String(), method.School.String(), method.Dingju.String(),
		method.QuarterRule.String(), method.BaseDingju.String(), method.DunSource, method.HourBoundary,
	)
	if err != nil {
		return Chart{}, err
	}
	if parsed.School == SchoolJinhanYuJing {
		return Chart{}, fmt.Errorf("jinhan_yujing has a distinct result model; use ComputeJinhanChart")
	}
	bz, chartTime := chartInputs(st)
	return computeChartWithYongShen(bz, chartTime, parsed, syms, birthDate), nil
}
