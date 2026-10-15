package mcpcontract

import (
	"context"
	"fmt"
	"time"
)

// TimeNowInput is deliberately empty: the server owns the clock. The generated
// schema still closes the object so unknown arguments fail validation.
type TimeNowInput struct{}

// TimeNowOutput is the MCP wire shape for the server clock.
type TimeNowOutput struct {
	UTC   string `json:"utc" jsonschema:"Current UTC timestamp in RFC3339 format."`
	Local string `json:"local" jsonschema:"Server-local timestamp in RFC3339 format."`
	CST   string `json:"cst" jsonschema:"Current China Standard Time (UTC+8) timestamp in RFC3339 format."`
}

// TimeNow returns the server clock in UTC, server-local time, and CST.
func TimeNow(ctx context.Context, _ TimeNowInput) (TimeNowOutput, error) {
	if err := ctx.Err(); err != nil {
		return TimeNowOutput{}, err
	}
	now := time.Now()
	return TimeNowOutput{
		UTC:   now.UTC().Format(time.RFC3339),
		Local: now.Format(time.RFC3339),
		CST:   now.In(time.FixedZone("CST", 8*3600)).Format(time.RFC3339),
	}, nil
}

// NewEngine returns the engine's typed MCP contract registry.
func NewEngine() *Registry {
	registry := New()
	mustRegister(registry, Tool[TimeNowInput, TimeNowOutput]{
		Name:        "time.now",
		Description: "获取服务端当前时间。返回 UTC、本地、北京时间，用于 AI agent 获取准确时间避免幻觉。",
		Handler:     TimeNow,
	})
	mustRegister(registry, Tool[HuangliDaysInput, []HuangliDayOutput]{
		Name:        "huangli.days",
		Description: "黄历查日。返回连续N天的黄历信息；event 给定时按建除事项规则输出 suitability。",
		Handler:     HuangliDays,
	})
	mustRegister(registry, Tool[TianwenTimeInput, TianwenTimeOutput]{
		Name:        "tianwen.time",
		Description: "根据时间和经度计算真太阳时，返回公历、真太阳时、农历三套时间。",
		Handler:     TianwenTime,
	})
	mustRegister(registry, Tool[CityCoordsInput, CityCoordsOutput]{
		Name:        "city.coords",
		Description: "根据城市名查询经纬度。优先使用内置城市表；未命中时可查询 Nominatim（OSM），部署方可用 LIKI_EXTERNAL_GEOCODING=off 禁用。",
		Handler:     CityCoords,
	})
	mustRegister(registry, Tool[BazhaiChartInput, BazhaiChartOutput]{
		Name:        "bazhai.chart",
		Description: "八宅排盘。按出生年份与性别排命卦 + 四吉四凶方 + 出生年紫白飞星。八宅不需要出生时辰，也不与八字四柱合参。",
		Handler:     BazhaiChart,
	})
	mustRegister(registry, Tool[BazhaiLayoutInput, BazhaiLayoutOutput]{
		Name:        "bazhai.layout",
		Description: "八宅门主灶配合。命卦 + 门/主/灶卦 → 方向、游年九星、吉凶。确定性计算。",
		Handler:     BazhaiLayout,
	})
	return registry
}

func mustRegister[In, Out any](registry *Registry, tool Tool[In, Out]) {
	if err := register(registry, tool); err != nil {
		panic(fmt.Sprintf("register engine MCP contract %s: %v", tool.Name, err))
	}
}
