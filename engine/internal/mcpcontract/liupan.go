package mcpcontract

import (
	"encoding/json"
	"fmt"

	"github.com/google/jsonschema-go/jsonschema"
)

// Liupan tools expose one domain method with one schema form per time layer.
// The distinct forms keep required facts exact (for example, liu_shi needs a
// longitude while xiao_yun does not) without turning one input DTO into a bag
// of mutually optional fields.

type baziLiupanLiuNianForm struct {
	Scope     BaziLiupanScopeName `json:"scope" jsonschema:"流盘层级。"`
	Subject   NatalSubject        `json:"subject" jsonschema:"已归一化的出生主体。"`
	StartTime StandardTime        `json:"start_time" jsonschema:"流年选择起点：local standard time，RFC3339 fixed offset。"`
	Count     *SeriesCount        `json:"count,omitempty" jsonschema:"连续查询年数。"`
}

type baziLiupanLiuYueForm struct {
	Scope     BaziLiupanScopeName `json:"scope" jsonschema:"流盘层级。"`
	Subject   NatalSubject        `json:"subject" jsonschema:"已归一化的出生主体。"`
	StartTime StandardTime        `json:"start_time" jsonschema:"流月选择起点：local standard time，RFC3339 fixed offset。"`
	Count     *SeriesCount        `json:"count,omitempty" jsonschema:"连续查询月数。"`
}

type baziLiupanLiuRiForm struct {
	Scope     BaziLiupanScopeName `json:"scope" jsonschema:"流盘层级。"`
	Subject   NatalSubject        `json:"subject" jsonschema:"已归一化的出生主体。"`
	StartTime StandardTime        `json:"start_time" jsonschema:"流日选择起点：local standard time，RFC3339 fixed offset。"`
	Longitude float64             `json:"longitude" jsonschema:"地点经度（-180..180）；真太阳时校正用于日柱。"`
	Count     *SeriesCount        `json:"count,omitempty" jsonschema:"连续查询日数。"`
}

type baziLiupanLiuShiForm struct {
	Scope     BaziLiupanScopeName `json:"scope" jsonschema:"流盘层级。"`
	Subject   NatalSubject        `json:"subject" jsonschema:"已归一化的出生主体。"`
	StartTime StandardTime        `json:"start_time" jsonschema:"流时选择起点：local standard time，RFC3339 fixed offset。"`
	Longitude float64             `json:"longitude" jsonschema:"地点经度（-180..180）；真太阳时校正用于日柱/时柱。"`
	Count     *SeriesCount        `json:"count,omitempty" jsonschema:"连续查询时辰数。"`
}

type baziLiupanXiaoYunForm struct {
	Scope   BaziLiupanScopeName `json:"scope" jsonschema:"流盘层级。"`
	Subject NatalSubject        `json:"subject" jsonschema:"已归一化的出生主体。"`
	Count   *SeriesCount        `json:"count,omitempty" jsonschema:"连续查询小运岁数。"`
}

type ziweiLiupanLiuNianForm struct {
	Scope     ZiweiLiupanScopeName `json:"scope" jsonschema:"流盘层级。"`
	Subject   NatalSubject         `json:"subject" jsonschema:"已归一化的出生主体。"`
	StartTime StandardTime         `json:"start_time" jsonschema:"流年选择起点：local standard time，RFC3339 fixed offset。"`
	Count     *SeriesCount         `json:"count,omitempty" jsonschema:"连续查询年数。"`
}

type ziweiLiupanLiuYueForm struct {
	Scope     ZiweiLiupanScopeName `json:"scope" jsonschema:"流盘层级。"`
	Subject   NatalSubject         `json:"subject" jsonschema:"已归一化的出生主体。"`
	StartTime StandardTime         `json:"start_time" jsonschema:"流月选择起点：local standard time，RFC3339 fixed offset。"`
	Count     *SeriesCount         `json:"count,omitempty" jsonschema:"连续查询月数。"`
}

type ziweiLiupanLiuRiForm struct {
	Scope     ZiweiLiupanScopeName `json:"scope" jsonschema:"流盘层级。"`
	Subject   NatalSubject         `json:"subject" jsonschema:"已归一化的出生主体。"`
	StartTime StandardTime         `json:"start_time" jsonschema:"流日选择起点：local standard time，RFC3339 fixed offset。"`
	Count     *SeriesCount         `json:"count,omitempty" jsonschema:"连续查询日数。"`
}

type ziweiLiupanLiuShiForm struct {
	Scope     ZiweiLiupanScopeName `json:"scope" jsonschema:"流盘层级。"`
	Subject   NatalSubject         `json:"subject" jsonschema:"已归一化的出生主体。"`
	StartTime StandardTime         `json:"start_time" jsonschema:"流时选择起点：local standard time，RFC3339 fixed offset。"`
	Longitude float64              `json:"longitude" jsonschema:"地点经度（-180..180）；真太阳时校正用于时辰。"`
	Count     *SeriesCount         `json:"count,omitempty" jsonschema:"连续查询时辰数。"`
}

func constFormSchema[T any](options *jsonschema.ForOptions, scope string) (json.RawMessage, error) {
	_, raw, err := inputSchemaFor[T](options)
	if err != nil {
		return nil, err
	}
	return withSchemaConst(raw, "scope", scope)
}

func longitudeBoundFormSchema[T any](options *jsonschema.ForOptions, scope string) (json.RawMessage, error) {
	raw, err := constFormSchema[T](options, scope)
	if err != nil {
		return nil, err
	}
	return withSchemaNumberBounds(raw, "longitude", -180, 180)
}

func oneOfSchema(forms []json.RawMessage) (json.RawMessage, error) {
	return json.Marshal(map[string]any{"type": "object", "oneOf": forms})
}

func baziLiupanFormsSchema(options *jsonschema.ForOptions) (json.RawMessage, error) {
	builders := []func() (json.RawMessage, error){
		func() (json.RawMessage, error) {
			return constFormSchema[baziLiupanLiuNianForm](options, "liu_nian")
		},
		func() (json.RawMessage, error) {
			return constFormSchema[baziLiupanLiuYueForm](options, "liu_yue")
		},
		func() (json.RawMessage, error) {
			return longitudeBoundFormSchema[baziLiupanLiuRiForm](options, "liu_ri")
		},
		func() (json.RawMessage, error) {
			return longitudeBoundFormSchema[baziLiupanLiuShiForm](options, "liu_shi")
		},
		func() (json.RawMessage, error) {
			return constFormSchema[baziLiupanXiaoYunForm](options, "xiao_yun")
		},
	}
	forms := make([]json.RawMessage, 0, len(builders))
	for _, build := range builders {
		raw, err := build()
		if err != nil {
			return nil, err
		}
		forms = append(forms, raw)
	}
	return oneOfSchema(forms)
}

func ziweiLiupanFormsSchema(options *jsonschema.ForOptions) (json.RawMessage, error) {
	builders := []func() (json.RawMessage, error){
		func() (json.RawMessage, error) {
			return constFormSchema[ziweiLiupanLiuNianForm](options, "liu_nian")
		},
		func() (json.RawMessage, error) {
			return constFormSchema[ziweiLiupanLiuYueForm](options, "liu_yue")
		},
		func() (json.RawMessage, error) {
			return constFormSchema[ziweiLiupanLiuRiForm](options, "liu_ri")
		},
		func() (json.RawMessage, error) {
			return longitudeBoundFormSchema[ziweiLiupanLiuShiForm](options, "liu_shi")
		},
	}
	forms := make([]json.RawMessage, 0, len(builders))
	for _, build := range builders {
		raw, err := build()
		if err != nil {
			return nil, err
		}
		forms = append(forms, raw)
	}
	return oneOfSchema(forms)
}

func baziLiupanOutputSchema(options *jsonschema.ForOptions) (json.RawMessage, error) {
	roots := []func() (json.RawMessage, error){
		func() (json.RawMessage, error) { return seriesOutputSchema[BaziLiuNianOutput](options, "liu_nian") },
		func() (json.RawMessage, error) { return seriesOutputSchema[BaziLiuYueOutput](options, "liu_yue") },
		func() (json.RawMessage, error) { return seriesOutputSchema[BaziLiuRiOutput](options, "liu_ri") },
		func() (json.RawMessage, error) { return seriesOutputSchema[BaziLiuShiOutput](options, "liu_shi") },
		func() (json.RawMessage, error) { return seriesOutputSchema[BaziXiaoYunItem](options, "xiao_yun") },
	}
	return domainFormsOutputSchema("bazi.liupan output schema", roots)
}

func ziweiLiupanOutputSchema(options *jsonschema.ForOptions) (json.RawMessage, error) {
	roots := []func() (json.RawMessage, error){
		func() (json.RawMessage, error) { return seriesOutputSchema[ZiweiLiuNianOutput](options, "liu_nian") },
		func() (json.RawMessage, error) { return seriesOutputSchema[ZiweiLiuYueOutput](options, "liu_yue") },
		func() (json.RawMessage, error) { return seriesOutputSchema[ZiweiLiuRiOutput](options, "liu_ri") },
		func() (json.RawMessage, error) { return seriesOutputSchema[ZiweiLiuShiOutput](options, "liu_shi") },
	}
	return domainFormsOutputSchema("ziwei.liupan output schema", roots)
}

func domainFormsOutputSchema(name string, builders []func() (json.RawMessage, error)) (json.RawMessage, error) {
	forms := make([]json.RawMessage, 0, len(builders))
	for _, build := range builders {
		raw, err := build()
		if err != nil {
			return nil, fmt.Errorf("%s: %w", name, err)
		}
		forms = append(forms, raw)
	}
	return oneOfSchema(forms)
}
