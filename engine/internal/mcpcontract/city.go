package mcpcontract

import (
	"context"

	"liki-engine/internal/engine/city"
)

// CityCoordsInput requests coordinates for one place name.
type CityCoordsInput struct {
	City CityName `json:"city" jsonschema:"城市名称（中英文均可）。"`
}

// CityCoordsOutput is the MCP wire shape for a resolved place.
type CityCoordsOutput struct {
	Name      string  `json:"name" jsonschema:"解析后的地名。"`
	Longitude float64 `json:"longitude" jsonschema:"WGS-84 经度，十进制度。"`
	Latitude  float64 `json:"latitude" jsonschema:"WGS-84 纬度，十进制度。"`
	Country   string  `json:"country" jsonschema:"国家或地区。"`
}

// CityCoords resolves built-in city names first and falls back to Nominatim.
func CityCoords(ctx context.Context, input CityCoordsInput) (CityCoordsOutput, error) {
	result, err := city.ResolveCoords(ctx, string(input.City))
	if err != nil {
		return CityCoordsOutput{}, err
	}
	return CityCoordsOutput(result), nil
}
