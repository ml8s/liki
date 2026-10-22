package mcpcontract

import (
	"context"
	"fmt"

	"liki-engine/internal/engine/city"
)

// CityCoordsInput requests candidates for one place name.
type CityCoordsInput struct {
	City CityName `json:"city" jsonschema:"城市名称（中英文均可）；同名区县请用省/市限定（如 北京朝阳区）。"`
}

// CityLevelName is the administrative level of a candidate.
type CityLevelName string

// CityLongitudeRangeOutput is the east-west extent of the recorded place.
type CityLongitudeRangeOutput struct {
	West float64 `json:"west" jsonschema:"西界经度（-180..180）。"`
	East float64 `json:"east" jsonschema:"东界经度（-180..180）。"`
}

// CityCandidateOutput is one possible place match.
type CityCandidateOutput struct {
	Name           string                    `json:"name" jsonschema:"候选地名（原样，如 长沙市、长沙县）。"`
	Region         string                    `json:"region" jsonschema:"省/州/国家，用于消歧（如 湖南省、美国）。"`
	City           string                    `json:"city,omitempty" jsonschema:"所属地级市（内置表已知时；如 杭州市）。"`
	Country        string                    `json:"country" jsonschema:"国家或地区；境内候选沿用省份名。"`
	Level          CityLevelName             `json:"level" jsonschema:"层级：city（地级市）/county（区县）/world（海外或港澳台城市）。"`
	Longitude      float64                   `json:"longitude" jsonschema:"WGS-84 经度，十进制度（东经为正）。"`
	Latitude       float64                   `json:"latitude" jsonschema:"WGS-84 纬度，十进制度（北纬为正）。"`
	Timezone       string                    `json:"timezone,omitempty" jsonschema:"IANA 时区名；候选无时区数据时省略，调用方须询问用户。"`
	LongitudeRange *CityLongitudeRangeOutput `json:"longitude_range,omitempty" jsonschema:"地点东西经度范围（由行政边界换算）；用于时间边界的不确定性。"`
}

// CityCoordsOutput is the MCP wire shape for place candidates.
type CityCoordsOutput struct {
	Candidates []CityCandidateOutput `json:"candidates" jsonschema:"地点候选（按相关度排序）。长度 1 可直接使用；多于 1 时必须由用户确认，不得静默取第一个。"`
}

// CityCoords resolves a place name against the embedded table. It never calls
// external geocoding: empty results mean the place is not in the table.
func CityCoords(ctx context.Context, input CityCoordsInput) (CityCoordsOutput, error) {
	candidates, err := city.ResolveCandidates(ctx, string(input.City))
	if err != nil {
		return CityCoordsOutput{}, err
	}
	if len(candidates) == 0 {
		return CityCoordsOutput{}, fmt.Errorf(
			"city: 未找到 %q；请尝试附近大城市，或直接提供经纬度和 IANA 时区", input.City)
	}
	out := CityCoordsOutput{Candidates: make([]CityCandidateOutput, 0, len(candidates))}
	for _, candidate := range candidates {
		out.Candidates = append(out.Candidates, cityCandidateOutput(candidate))
	}
	return out, nil
}

func cityCandidateOutput(candidate city.Candidate) CityCandidateOutput {
	item := CityCandidateOutput{
		Name:      candidate.Name,
		Region:    candidate.Region,
		City:      candidate.City,
		Country:   candidate.Country,
		Level:     CityLevelName(candidate.Level),
		Longitude: candidate.Longitude,
		Latitude:  candidate.Latitude,
		Timezone:  candidate.Timezone,
	}
	if candidate.BBox != nil {
		item.LongitudeRange = &CityLongitudeRangeOutput{
			West: candidate.BBox.West,
			East: candidate.BBox.East,
		}
	}
	return item
}
