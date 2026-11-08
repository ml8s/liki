package mcpcontract

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEveryEngineToolActualOutputMatchesSchema(t *testing.T) {
	registry := testEngine()
	validated := make(map[string]bool, len(registry.names))
	call := func(name string, input any) json.RawMessage {
		t.Helper()
		raw, err := registry.Call(context.Background(), name, mustJSON(t, input))
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if err := registry.ValidateOutput(name, raw); err != nil {
			t.Fatalf("%s output contract: %v", name, err)
		}
		validated[name] = true
		return raw
	}

	call("time_now", map[string]any{"timezone": "Asia/Shanghai"})
	call("subject_candidates", civilResolveInput(t, "1990-05-20T03:49:00Z", "Asia/Shanghai", 116.4))
	call("city_info", CityInfoInput{City: "北京"})
	call("huangli_days", HuangliDaysInput{
		StandardTime: "2024-06-15T12:00:00+08:00",
		Count:        HuangliCount(1),
		Event:        "sign",
	})
	call("bazi_chart", BaziChartInput{Subject: testNatalSubject(t, standardInputFromInstant(t, "1990-05-20T03:49:00Z", "Asia/Shanghai"), "male")})
	call("bazhai_chart", BazhaiChartInput{Subject: BazhaiSubjectInput{StandardTime: "1984-06-15T12:00:00+08:00", Gender: "male"}})
	call("bazhai_layout", BazhaiLayoutInput{
		MingGua: "乾", DoorGua: "坤", MasterGua: "艮", StoveGua: "兑",
	})
	xuankongRaw := call("xuankong_chart", XuankongChartInput{
		StandardTime: standardInputFromInstant(t, "2026-07-31T04:00:00Z", "Asia/Shanghai"), ZuoShan: 2, XiangShan: 14,
	})
	var xuankongChart XuankongChart
	decode(t, xuankongRaw, &xuankongChart)
	call("xuankong_liunian", XuankongLiunianInput{
		StandardTime: standardInputFromInstant(t, "2026-06-15T04:00:00Z", "Asia/Shanghai"), ZuoShan: 2, XiangShan: 14,
	})
	qiguaRaw := call("liuyao_qigua", LiuyaoQiguaInput{Casting: LiuyaoCastingRequestInput{
		Mode: "yaos", Yaos: LiuyaoYaoValuesInput{7, 8, 7, 8, 7, 8},
	}})
	var qigua LiuyaoQiguaOutput
	decode(t, qiguaRaw, &qigua)
	call("liuyao_chart", LiuyaoChartInput{
		StandardTime: standardInputFromInstant(t, "2026-07-31T02:00:00Z", "Asia/Shanghai"),
		Longitude:    116.4, Casting: &qigua.Casting,
	})

	// Captured legacy fixtures cover the migrated high-volume bazi, ziwei, and
	// qimen families, including every downstream chart consumer.
	dir := filepath.Join("testdata", "wire_inputs")
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		name := fixtureTool(strings.TrimSuffix(entry.Name(), ".json"))
		if name == "" || name == "bazi_chart" {
			continue
		}
		raw, err := os.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			t.Fatal(err)
		}
		var fixture struct {
			Input json.RawMessage `json:"input"`
		}
		if err := json.Unmarshal(raw, &fixture); err != nil {
			t.Fatalf("%s: %v", entry.Name(), err)
		}
		call(name, fixture.Input)
	}

	for _, name := range registry.Names() {
		if !validated[name] {
			t.Errorf("%s has no actual-output schema validation", name)
		}
	}
}
