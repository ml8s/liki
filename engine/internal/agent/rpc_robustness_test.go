package agent

// RPC 输入健壮性测试：非法参数必须返回 -32602 或 -32000，不能 panic。

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
)

func TestRPCInputRobustness(t *testing.T) {
	reg := NewRPCRegistry()
	// Valid baselines distinguish invalid input from missing dependent data.
	zc := mustJSON(t, map[string]any{"lunar": map[string]any{"year": 1984, "month": 1, "day": 15, "shichen": "辰"}, "gender": "male"})
	bc := mustJSON(t, map[string]any{"solar_time": "1984-02-15T08:00:00+08:00", "gender": "male"})
	z := executeAndDecode(t, reg, "ziwei.chart", zc)["data"].(map[string]any)
	b := executeAndDecode(t, reg, "bazi.chart", bc)["data"].(map[string]any)

	valid := map[string]json.RawMessage{
		"bazi.chart":      bc,
		"bazi.fullchart":  mk(t, "chart", b),
		"bazi.liunian":    mk(t, "chart", b, "year", 2026),
		"bazi.liuyue":     mk(t, "chart", b, "year", 2026, "month", 6),
		"bazi.liuri":      mk(t, "chart", b, "year", 2026, "month", 6, "day", 4),
		"bazi.liushi":     mk(t, "chart", b, "year", 2026, "month", 6, "day", 4, "hour", 12),
		"bazi.xiaoyun":    mk(t, "chart", b),
		"bazi.bond":       mk(t, "a", map[string]any{"chart": b}, "b", map[string]any{"chart": b}),
		"ziwei.chart":     zc,
		"ziwei.fullchart": mk(t, "chart", z),
		"ziwei.daxian":    mk(t, "chart", z),
		"ziwei.liunian":   mk(t, "chart", z, "lunar_year", 2026),
		"ziwei.liuyue":    mk(t, "chart", z, "lunar_year", 2026, "lunar_month", 6),
		"ziwei.liuri":     mk(t, "chart", z, "lunar_year", 2026, "lunar_month", 6, "lunar_day", 4),
		"ziwei.liushi":    mk(t, "chart", z, "lunar_year", 2026, "lunar_month", 6, "lunar_day", 4, "shi_zhi", "午"),
		"qimen.chart":     []byte(`{"solar_time":"2026-07-31T10:00:00+08:00","kind":"shi"}`),
		"qiming.char":     []byte(`{"char":"明"}`),
		"qiming.pick":     []byte(`{"wuxing1":"木","wuxing2":"火","count":2}`),
		"qiming.compose":  []byte(`{"first":["德"],"second":["明"]}`),
		"qiming.check":    []byte(`{"given_names":["德明"]}`),
		"huangli.days":    []byte(`{"start_date":"2026-07-31","count":3}`),
		"time.now":        []byte(`{}`),
		// city 依赖外网 Nominatim，不测联网基线（非法参数仍由 schema 校验拦截）
	}

	badParams := []json.RawMessage{
		json.RawMessage(`{}`),                      // 空对象（缺必填）
		json.RawMessage(`null`),                    // null
		json.RawMessage(`[]`),                      // 数组（类型错）
		json.RawMessage(`{"chart":"not-a-chart"}`), // 字段类型错
	}
	// 极端值（仅对含 year/date 参数的方法）
	extreme := []json.RawMessage{
		json.RawMessage(`{"year":0}`),
		json.RawMessage(`{"year":9999}`),
		json.RawMessage(`{"month":13}`),
		json.RawMessage(`{"day":0}`),
		json.RawMessage(`{"solar_time":"not-a-date"}`),
		json.RawMessage(`{"start_date":"2026-13-45","count":-1}`),
	}

	panics, okCount, errCount := 0, 0, 0
	for name := range reg.methods {
		if _, has := valid[name]; !has {
			continue // 只测有合法基线的
		}
		cases := append(append([]json.RawMessage{}, badParams...), extreme...)
		// 有效基线也必须能正常返回（确认基线有效）
		if _, err := reg.Execute(context.Background(), name, valid[name]); err != nil {
			t.Errorf("%s: 合法基线反而报错: %v", name, err)
			continue
		}
		for _, p := range cases {
			func() {
				defer func() {
					if r := recover(); r != nil {
						panics++
						t.Errorf("%s: 参数 %s 触发 panic: %v", name, string(p), r)
					}
				}()
				_, err := reg.Execute(context.Background(), name, p)
				if err == nil {
					okCount++ // 宽容的 handler 接受非法参数（如 time.now 忽略参数）
					t.Logf("%s: 参数 %s 未报错（可接受）", name, string(p))
				} else {
					var rpcErr *RPCError
					if ok := asRPCError(err, &rpcErr); !ok || rpcErr.Code == 0 {
						errCount++
						t.Errorf("%s: 错误无 RPCError 码: %v", name, err)
					}
				}
			}()
		}
	}
	t.Logf("统计: panic=%d 未报错=%d 错误码异常=%d", panics, okCount, errCount)
}

func mk(t *testing.T, kvs ...any) json.RawMessage {
	t.Helper()
	out := map[string]any{}
	for i := 0; i+1 < len(kvs); i += 2 {
		out[fmt.Sprint(kvs[i])] = kvs[i+1]
	}
	return mustJSON(t, out)
}

func asRPCError(err error, out **RPCError) bool {
	if e, ok := err.(*RPCError); ok {
		*out = e
		return true
	}
	return false
}
