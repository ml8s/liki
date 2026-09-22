package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"testing"
	"time"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

func TestQimenPatternSchemaMatchesTable(t *testing.T) {
	tablePath := filepath.Join(
		"..", "..", "..", "engine", "internal", "engine", "qimen",
		"data", "patterns.json",
	)
	rawTable, err := os.ReadFile(tablePath)
	if err != nil {
		t.Fatal(err)
	}
	var table []struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(rawTable, &table); err != nil {
		t.Fatal(err)
	}
	tableNames := make([]string, 0, len(table))
	for _, rule := range table {
		tableNames = append(tableNames, rule.Name)
	}

	var document struct {
		Result struct {
			OneOf []struct {
				Properties struct {
					Patterns struct {
						Items struct {
							Properties struct {
								Name struct {
									Enum []string `json:"enum"`
								} `json:"name"`
							} `json:"properties"`
						} `json:"items"`
					} `json:"patterns"`
				} `json:"properties"`
			} `json:"oneOf"`
		} `json:"result"`
	}
	decodeJSON(t, qimenSchemaJSON, &document)

	var schemaNames []string
	for _, branch := range document.Result.OneOf {
		if names := branch.Properties.Patterns.Items.Properties.Name.Enum; len(names) > 0 {
			schemaNames = names
			break
		}
	}
	if len(schemaNames) == 0 {
		t.Fatal("qimen result schema lacks pattern name enum")
	}
	if !reflect.DeepEqual(schemaNames, tableNames) {
		t.Fatalf("pattern schema enum = %v, want table order %v", schemaNames, tableNames)
	}
}

func TestQimenSchemaParamsMatchMethodMatrix(t *testing.T) {
	params := qimenParamsSchema(t)
	if additional := params.AdditionalProperties; additional == nil || *additional {
		t.Fatal("qimen.chart params must reject undeclared fields")
	}
	if !reflect.DeepEqual(params.Required, []string{"solar_time"}) {
		t.Fatalf("required = %v, want solar_time", params.Required)
	}
	if format := params.Properties["solar_time"].(map[string]any)["format"]; format != "date-time" {
		t.Fatalf("solar_time format = %v, want date-time", format)
	}

	assertEnum(t, params.Properties["scope"], []any{"hour", "day", "quarter", "month", "year"})
	assertEnum(t, params.Properties["school"], []any{"zhuanpan", "luoshu_feipan", "mingfa_feipan", "jinhan_yujing"})
	assertEnum(t, params.Properties["dingju_method"], []any{"chaibu", "zhirun", "maoshan"})
	assertEnum(t, params.Properties["quarter_rule"], []any{"ten_minute_sanyuan", "twelve_minute_ten_division"})
	assertEnum(t, params.Properties["base_dingju_method"], []any{"chaibu", "zhirun", "maoshan"})
	assertEnum(t, params.Properties["dun_source"], []any{"hour_branch", "solar_term"})
	assertEnum(t, params.Properties["hour_boundary"], []any{"late_zi", "zi_zheng"})
	yongShen := params.Properties["yong_shen"].(map[string]any)
	if minItems, ok := yongShen["minItems"].(float64); !ok || minItems != 1 {
		t.Fatalf("yong_shen minItems = %v, want 1", yongShen["minItems"])
	}
	if unique, ok := yongShen["uniqueItems"].(bool); !ok || !unique {
		t.Fatalf("yong_shen uniqueItems = %v, want true", yongShen["uniqueItems"])
	}
	assertSameKeys(t, params.Properties, map[string]any{
		"solar_time": nil, "scope": nil, "school": nil,
		"dingju_method": nil, "quarter_rule": nil, "base_dingju_method": nil,
		"dun_source": nil, "hour_boundary": nil, "yong_shen": nil, "birth_date": nil,
	})
	if _, exists := params.Properties["matter"]; exists {
		t.Fatal("qimen.chart engine API must not own matter routing; use the skill orchestration table")
	}

	conditions := params.AllOf
	find := func(t *testing.T, want string) {
		t.Helper()
		var expected any
		if err := json.Unmarshal([]byte(want), &expected); err != nil {
			t.Fatal(err)
		}
		normalized, _ := json.Marshal(expected)
		for _, condition := range conditions {
			encoded, _ := json.Marshal(condition)
			if string(encoded) == string(normalized) {
				return
			}
		}
		t.Fatalf("qimen.chart lacks condition %s", normalized)
	}
	find(t, `{"if":{"properties":{"scope":{"enum":["hour","day"]}},"required":["scope"]},"then":{"properties":{"dingju_method":{"enum":["chaibu","zhirun","maoshan"]},"quarter_rule":{"not":{}},"base_dingju_method":{"not":{}},"dun_source":{"not":{}},"hour_boundary":{"not":{}}}}}`)
	find(t, `{"if":{"not":{"required":["scope"]}},"then":{"properties":{"quarter_rule":{"not":{}},"base_dingju_method":{"not":{}},"dun_source":{"not":{}},"hour_boundary":{"not":{}}}}}`)
	find(t, `{"if":{"properties":{"scope":{"const":"quarter"}},"required":["scope"]},"then":{"properties":{"dingju_method":{"not":{}},"quarter_rule":{"enum":["ten_minute_sanyuan","twelve_minute_ten_division"]},"school":{"const":"zhuanpan"}}}}`)
	find(t, `{"if":{"properties":{"scope":{"enum":["month","year"]}},"required":["scope"]},"then":{"properties":{"base_dingju_method":{"not":{}},"dingju_method":{"not":{}},"dun_source":{"not":{}},"hour_boundary":{"not":{}},"quarter_rule":{"not":{}}}}}`)
	find(t, `{"if":{"properties":{"quarter_rule":{"const":"ten_minute_sanyuan"}},"required":["quarter_rule"]},"then":{"properties":{"base_dingju_method":{"not":{}},"dun_source":{"not":{}},"hour_boundary":{"not":{}}}}}`)
	find(t, `{"if":{"properties":{"quarter_rule":{"const":"twelve_minute_ten_division"}},"required":["quarter_rule"]},"then":{"properties":{"base_dingju_method":{"enum":["chaibu","zhirun","maoshan"]},"dun_source":{"enum":["hour_branch","solar_term"]},"hour_boundary":{"enum":["late_zi","zi_zheng"]}}}}`)
	find(t, `{"if":{"properties":{"dingju_method":{"const":"maoshan"}},"required":["dingju_method"]},"then":{"properties":{"school":{"const":"zhuanpan"},"scope":{"const":"hour"}}}}`)
}

func TestQimenParamsSchemaDefaultScopeMatchesHourScope(t *testing.T) {
	schema := compileQimenParamsSchema(t)
	valid := []map[string]any{
		{"solar_time": "2026-09-07T12:00:00+08:00"},
		{"solar_time": "2026-09-07T12:00:00+08:00", "dingju_method": "zhirun"},
		{"solar_time": "2026-09-07T12:00:00+08:00", "school": "luoshu_feipan"},
	}
	for _, params := range valid {
		if err := schema.Validate(params); err != nil {
			t.Errorf("default-hour params rejected: %v", err)
		}
	}

	invalid := []map[string]any{
		{"solar_time": "2026-09-07T12:00:00+08:00", "quarter_rule": "ten_minute_sanyuan"},
		{"solar_time": "2026-09-07T12:00:00+08:00", "quarter_rule": "twelve_minute_ten_division"},
		{"solar_time": "2026-09-07T12:00:00+08:00", "base_dingju_method": "zhirun"},
		{"solar_time": "2026-09-07T12:00:00+08:00", "dun_source": "solar_term"},
		{"solar_time": "2026-09-07T12:00:00+08:00", "hour_boundary": "zi_zheng"},
	}
	for _, params := range invalid {
		if err := schema.Validate(params); err == nil {
			t.Errorf("default-hour params accepted quarter-only field: %v", params)
		}
	}
}

func TestQimenSchemaDocumentIsSelfContained(t *testing.T) {
	var document map[string]any
	decodeJSON(t, qimenSchemaJSON, &document)
	assertSameKeys(t, document, map[string]any{"params": nil, "result": nil})
	assertNoSchemaRef(t, document)
}

func TestQimenPublicSchemaExcludesProjectNames(t *testing.T) {
	var document map[string]any
	decodeJSON(t, qimenSchemaJSON, &document)
	projectNames := regexp.MustCompile(`(?i)horosa|kinqimen|atopx|potuo|mingyu|deminzhang|大师奇门`)
	assertNoProjectName(t, document, projectNames)
}

func TestQimenSchemaParamExamplesCoverUserJourneys(t *testing.T) {
	var document struct {
		Params struct {
			Examples []map[string]any `json:"examples"`
		} `json:"params"`
	}
	decodeJSON(t, qimenSchemaJSON, &document)
	if len(document.Params.Examples) != 7 {
		t.Fatalf("qimen params examples count = %d, want 7", len(document.Params.Examples))
	}

	wantExamples := []map[string]any{
		{"solar_time": "2026-09-07T12:00:00+08:00"},
		{
			"solar_time": "2026-09-07T12:00:00+08:00",
			"scope":      "hour", "school": "zhuanpan", "dingju_method": "zhirun",
		},
		{"solar_time": "2026-09-07T12:00:00+08:00", "scope": "day"},
		{
			"solar_time": "2026-09-07T12:00:00+08:00", "scope": "quarter",
			"school": "zhuanpan", "quarter_rule": "ten_minute_sanyuan",
		},
		{
			"solar_time": "2026-09-07T12:00:00+08:00", "scope": "quarter",
			"school": "zhuanpan", "quarter_rule": "twelve_minute_ten_division",
			"base_dingju_method": "zhirun", "dun_source": "solar_term", "hour_boundary": "zi_zheng",
		},
		{"solar_time": "2026-09-07T12:00:00+08:00", "scope": "day", "school": "jinhan_yujing"},
		{
			"solar_time": "2026-09-07T12:00:00+08:00",
			"yong_shen":  []any{"生门", "戊"}, "birth_date": "1984-02-15",
		},
	}
	for i, want := range wantExamples {
		got := document.Params.Examples[i]
		if len(got) != len(want) {
			t.Fatalf("example %d field count = %d, want %d", i, len(got), len(want))
		}
		for key, value := range want {
			if !reflect.DeepEqual(got[key], value) {
				t.Fatalf("example %d field %s = %v, want %v", i, key, got[key], value)
			}
		}
	}

	reg := NewRPCRegistry()
	for _, example := range document.Params.Examples {
		encoded, err := json.Marshal(example)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := reg.Execute(context.Background(), "qimen.chart", encoded); err != nil {
			t.Fatalf("example %s rejected by handler: %v", encoded, err)
		}
	}
}

func assertNoProjectName(t *testing.T, value any, pattern *regexp.Regexp) {
	t.Helper()
	switch typed := value.(type) {
	case map[string]any:
		for key, child := range typed {
			if pattern.MatchString(key) || pattern.MatchString(fmt.Sprint(child)) {
				t.Fatalf("public qimen schema contains project name at %q", key)
			}
			assertNoProjectName(t, child, pattern)
		}
	case []any:
		for _, child := range typed {
			assertNoProjectName(t, child, pattern)
		}
	case string:
		if pattern.MatchString(typed) {
			t.Fatalf("public qimen schema contains project name %q", typed)
		}
	}
}

func assertNoSchemaRef(t *testing.T, value any) {
	t.Helper()
	switch typed := value.(type) {
	case map[string]any:
		if _, exists := typed["$ref"]; exists {
			t.Fatal("qimen schema must stay self-contained and cannot use $ref")
		}
		for _, child := range typed {
			assertNoSchemaRef(t, child)
		}
	case []any:
		for _, child := range typed {
			assertNoSchemaRef(t, child)
		}
	}
}

func TestQimenSchemaYongShenSymbolsMatchParser(t *testing.T) {
	params := qimenParamsSchema(t)
	items := params.Properties["yong_shen"].(map[string]any)["items"].(map[string]any)
	if _, exists := items["enum"]; exists {
		t.Fatal("yong_shen symbols must be school-specific, not a global enum")
	}
	shared := []any{
		"休门", "生门", "伤门", "杜门", "景门", "死门", "惊门", "开门",
		"天蓬", "天芮", "天禽", "天冲", "天辅", "天心", "天柱", "天任", "天英",
		"值符", "螣蛇", "太阴", "六合", "勾陈", "朱雀", "白虎", "玄武", "九地", "九天", "太常",
		"乙", "丙", "丁", "戊", "己", "庚", "辛", "壬", "癸",
	}

	reg := NewRPCRegistry()
	for _, symbol := range shared {
		params := mustJSON(t, map[string]any{
			"solar_time": "2026-03-02T18:30:00+08:00",
			"yong_shen":  []string{symbol.(string)},
		})
		if _, err := reg.Execute(context.Background(), "qimen.chart", params); err != nil {
			t.Errorf("symbol %v rejected by engine: %v", symbol, err)
		}
	}
	centerParams := mustJSON(t, map[string]any{
		"solar_time": "2026-03-02T18:30:00+08:00",
		"school":     "mingfa_feipan",
		"yong_shen":  []string{"中门"},
	})
	if _, err := reg.Execute(context.Background(), "qimen.chart", centerParams); err != nil {
		t.Errorf("Ming Fa center door rejected by engine: %v", err)
	}
	assertError(t, reg, "qimen.chart", []byte(`{"solar_time":"2026-03-02T18:30:00+08:00","yong_shen":["中门"]}`))
	assertError(t, reg, "qimen.chart", []byte(`{"solar_time":"2026-03-02T18:30:00+08:00","yong_shen":["甲"]}`))
	assertError(t, reg, "qimen.chart", []byte(`{"solar_time":"2026-03-02T18:30:00+08:00","yong_shen":[]}`))
	assertError(t, reg, "qimen.chart", []byte(`{"solar_time":"2026-03-02T18:30:00+08:00","yong_shen":["生门","生门"]}`))
}

func TestQimenSchemaAndHandlerMethodMatrix(t *testing.T) {
	reg := NewRPCRegistry()
	schema := compileQimenResultSchema(t)
	type expected struct {
		method     string
		source     string
		spiritMode string
		hasTianQin bool
		wuBuYuShi  bool
	}
	tests := map[string]expected{
		`{"scope":"hour","school":"zhuanpan","dingju_method":"chaibu"}`:                       {"chaibu", "solar_term", "eight_spirits_yang_gouchen_zhuque_yin_baihu_xuanwu", true, true},
		`{"scope":"hour","school":"zhuanpan","dingju_method":"zhirun"}`:                       {"zhirun", "solar_term", "eight_spirits_yang_gouchen_zhuque_yin_baihu_xuanwu", true, true},
		`{"scope":"hour","school":"luoshu_feipan","dingju_method":"chaibu"}`:                  {"chaibu", "solar_term", "nine_spirits_baihu_taichang_xuanwu", false, true},
		`{"scope":"hour","school":"luoshu_feipan","dingju_method":"zhirun"}`:                  {"zhirun", "solar_term", "nine_spirits_baihu_taichang_xuanwu", false, true},
		`{"scope":"day","school":"zhuanpan","dingju_method":"chaibu"}`:                        {"chaibu", "solar_term", "eight_spirits_yang_gouchen_zhuque_yin_baihu_xuanwu", true, false},
		`{"scope":"day","school":"zhuanpan","dingju_method":"zhirun"}`:                        {"zhirun", "solar_term", "eight_spirits_yang_gouchen_zhuque_yin_baihu_xuanwu", true, false},
		`{"scope":"day","school":"luoshu_feipan","dingju_method":"chaibu"}`:                   {"chaibu", "solar_term", "nine_spirits_baihu_taichang_xuanwu", false, false},
		`{"scope":"day","school":"luoshu_feipan","dingju_method":"zhirun"}`:                   {"zhirun", "solar_term", "nine_spirits_baihu_taichang_xuanwu", false, false},
		`{"scope":"quarter","school":"zhuanpan","quarter_rule":"ten_minute_sanyuan"}`:         {"ten_minute_sanyuan", "quarter_rule", "eight_spirits_yang_gouchen_zhuque_yin_baihu_xuanwu", true, false},
		`{"scope":"quarter","school":"zhuanpan","quarter_rule":"twelve_minute_ten_division"}`: {"twelve_minute_ten_division", "quarter_rule", "eight_spirits_yang_gouchen_zhuque_yin_baihu_xuanwu", true, false},
		`{"scope":"month","school":"zhuanpan"}`:                                               {"", "month_cycle", "eight_spirits_yang_gouchen_zhuque_yin_baihu_xuanwu", true, false},
		`{"scope":"month","school":"luoshu_feipan"}`:                                          {"", "month_cycle", "nine_spirits_baihu_taichang_xuanwu", false, false},
		`{"scope":"year","school":"zhuanpan"}`:                                                {"", "year_cycle", "eight_spirits_yang_gouchen_zhuque_yin_baihu_xuanwu", true, false},
		`{"scope":"year","school":"luoshu_feipan"}`:                                           {"", "year_cycle", "nine_spirits_baihu_taichang_xuanwu", false, false},
		`{"scope":"hour","school":"mingfa_feipan"}`:                                           {"chaibu", "solar_term", "nine_spirits_gouchen_taichang_zhuque", false, true},
	}

	actualTopFields := make(map[string]any)
	actualMethodFields := make(map[string]any)
	actualPanFields := make(map[string]any)
	actualGongFields := make(map[string]any)
	for methodParams, want := range tests {
		t.Run(methodParams, func(t *testing.T) {
			params := []byte(`{"solar_time":"2026-03-02T18:30:00+08:00",` + methodParams[1:])
			output := executeAndDecode(t, reg, "qimen.chart", params)
			if err := schema.Validate(output); err != nil {
				t.Fatalf("result schema: %v", err)
			}
			data := output["data"].(map[string]any)
			collectKeys(data, actualTopFields)
			method := data["method"].(map[string]any)
			collectKeys(method, actualMethodFields)
			pan := data["pan"].(map[string]any)
			collectKeys(pan, actualPanFields)
			for _, field := range []string{
				"school", "lead_gan", "lead_zhi",
				"zhi_fu_xing_real_gong", "zhi_shi_men_real_gong",
			} {
				if _, exists := pan[field]; exists {
					t.Fatalf("pan.%s duplicates a method or top-level fact", field)
				}
			}
			for _, palace := range pan["gong_wei"].([]any) {
				collectKeys(palace.(map[string]any), actualGongFields)
			}
			var requested struct {
				Scope  string `json:"scope"`
				School string `json:"school"`
			}
			decodeJSON(t, []byte(methodParams), &requested)
			if method["scope"] != requested.Scope || method["school"] != requested.School {
				t.Fatalf("method scope/school = %v/%v, want %v/%v", method["scope"], method["school"], requested.Scope, requested.School)
			}
			wantScopeName := map[string]string{
				"hour": "时家奇门", "day": "日家奇门", "quarter": "刻家奇门",
				"month": "月家周期奇门", "year": "年家周期奇门",
			}[requested.Scope]
			wantSchoolName := map[string]string{
				"zhuanpan": "转盘法", "luoshu_feipan": "洛书飞盘法", "mingfa_feipan": "鸣法飞盘",
			}[requested.School]
			wantSchoolRules := map[string]map[string]string{
				"zhuanpan": {
					"geometry":      "luoshu_ring_rotation",
					"star_mode":     "tian_qin_lodges_kun",
					"door_mode":     "center_projects_kun",
					"spirit_flight": "yang_forward_yin_reverse",
				},
				"mingfa_feipan": {
					"geometry":      "mingfa_star_door_forward_flight",
					"star_mode":     "nine_stars_forward_with_center",
					"door_mode":     "nine_doors_with_center_forward",
					"spirit_flight": "yang_forward_yin_reverse",
				},
				"luoshu_feipan": {
					"geometry":      "palace_number_shift",
					"star_mode":     "star_gan_shift_together",
					"door_mode":     "eight_doors_real_center_step",
					"spirit_flight": "yang_forward_yin_reverse",
				},
			}[requested.School]
			methodNameMatches := false
			if requested.Scope == "quarter" {
				methodNameMatches = method["quarter_rule_name"] == map[string]string{
					"ten_minute_sanyuan":         "十分钟三元刻家",
					"twelve_minute_ten_division": "十二分钟十分局刻家",
				}[want.method]
				if _, exists := method["dingju_method_name"]; exists {
					t.Fatal("quarter chart exposes a top-level dingju method name")
				}
			} else if want.method == "" {
				_, exists := method["dingju_method_name"]
				methodNameMatches = !exists
			} else {
				methodNameMatches = method["dingju_method_name"] == map[string]string{
					"chaibu": "拆补定局", "zhirun": "置闰定局",
				}[want.method]
			}
			if method["scope_name"] != wantScopeName || method["school_name"] != wantSchoolName || !methodNameMatches {
				t.Fatalf("method display names = %v/%v/%v/%v", method["scope_name"], method["school_name"], method["dingju_method_name"], method["quarter_rule_name"])
			}
			if method["method_source"] != want.source || method["spirit_mode"] != want.spiritMode {
				t.Fatalf("method metadata = %+v", method)
			}
			for field, value := range wantSchoolRules {
				if method[field] != value {
					t.Fatalf("method %s = %v, want %v", field, method[field], value)
				}
			}
			if got := method["lead_pillar"].(map[string]any)["name"]; got == "" {
				t.Fatal("lead_pillar name missing")
			}
			quarter, hasQuarter := method["quarter"].(map[string]any)
			if requested.Scope == "quarter" {
				index, indexOK := quarter["index"].(float64)
				minutes, minutesOK := quarter["minutes_per_quarter"].(float64)
				leadMode, leadModeOK := quarter["lead_pillar_mode"].(string)
				wantLeadMode := "quarter"
				wantMinutes := 10.0
				if requestedQuarterRule(t, methodParams) == "twelve_minute_ten_division" {
					wantLeadMode = "hour"
					wantMinutes = 12
				}
				if !hasQuarter || !indexOK || index < 1 || index > 12 ||
					!minutesOK || minutes != wantMinutes || !leadModeOK || leadMode != wantLeadMode {
					t.Fatalf("quarter metadata = %#v", method["quarter"])
				}
			} else if hasQuarter {
				t.Fatalf("non-quarter chart exposes quarter metadata: %#v", quarter)
			}
			if requested.Scope == "quarter" {
				if _, exists := method["dingju_method"]; exists {
					t.Fatal("quarter chart exposes a top-level dingju method")
				}
				if method["quarter_rule"] != want.method {
					t.Fatalf("quarter rule = %v, want %q", method["quarter_rule"], want.method)
				}
			} else if got, exists := method["dingju_method"]; want.method == "" {
				if exists && got != "" {
					t.Fatalf("dingju method = %v, want omitted", got)
				}
			} else if got != want.method {
				t.Fatalf("dingju method = %v, want %q", got, want.method)
			}
			_, hasTianQin := method["tian_qin"]
			if hasTianQin != want.hasTianQin {
				t.Fatalf("tian_qin present = %v", hasTianQin)
			}
			if !want.wuBuYuShi && data["pan"].(map[string]any)["wu_bu_yu_shi"].(bool) {
				t.Fatal("non-shi chart must force wu_bu_yu_shi=false")
			}
		})
	}

	yongParams := []byte(`{"solar_time":"2026-03-02T18:30:00+08:00","yong_shen":["生门"],"birth_date":"1984-02-15"}`)
	collectKeys(executeAndDecode(t, reg, "qimen.chart", yongParams)["data"].(map[string]any), actualTopFields)

	schemaProperties := qimenDataSchema(t).Properties
	assertSameKeys(t, actualTopFields, schemaProperties)
	assertSameKeys(t, actualMethodFields, schemaProperties["method"].(map[string]any)["properties"].(map[string]any))
	assertSameKeys(t, actualPanFields, schemaProperties["pan"].(map[string]any)["properties"].(map[string]any))
	assertSameKeys(t, actualGongFields, schemaProperties["pan"].(map[string]any)["properties"].(map[string]any)["gong_wei"].(map[string]any)["items"].(map[string]any)["properties"].(map[string]any))
}

func TestQimenSchemaAndHandlerYongShenPath(t *testing.T) {
	reg := NewRPCRegistry()
	schema := compileQimenResultSchema(t)
	params := []byte(`{
		"solar_time":"2026-03-02T18:30:00+08:00",
		"scope":"hour","school":"luoshu_feipan","dingju_method":"zhirun",
		"yong_shen":["生门","天辅","六合","戊"],
		"birth_date":"1984-02-04T06:00:00+08:00"
	}`)
	output := executeAndDecode(t, reg, "qimen.chart", params)
	if err := schema.Validate(output); err != nil {
		t.Fatalf("result schema: %v", err)
	}
	data := output["data"].(map[string]any)
	yongShen, ok := data["yong_shen"].(map[string]any)
	if !ok {
		t.Fatal("yong_shen missing")
	}
	if yongShen["nian_gan_gong"] == nil {
		t.Fatal("birth_date did not produce nian_gan_gong")
	}
	if len(yongShen["symbols"].([]any)) != 4 {
		t.Fatalf("symbols = %v", yongShen["symbols"])
	}
	assertSchemaAllows(t, "qimen.chart", data, qimenDataSchema(t).Properties, "data")
}

func TestQimenSchemaResultAcrossBoundaryDates(t *testing.T) {
	reg := NewRPCRegistry()
	schema := compileQimenResultSchema(t)
	methods := []string{
		`{"scope":"hour","school":"zhuanpan","dingju_method":"chaibu"}`,
		`{"scope":"hour","school":"zhuanpan","dingju_method":"zhirun"}`,
		`{"scope":"hour","school":"luoshu_feipan","dingju_method":"chaibu"}`,
		`{"scope":"hour","school":"luoshu_feipan","dingju_method":"zhirun"}`,
		`{"scope":"hour","school":"zhuanpan","dingju_method":"maoshan"}`,
		`{"scope":"hour","school":"mingfa_feipan"}`,
		`{"scope":"day","school":"zhuanpan","dingju_method":"chaibu"}`,
		`{"scope":"day","school":"zhuanpan","dingju_method":"zhirun"}`,
		`{"scope":"day","school":"luoshu_feipan","dingju_method":"chaibu"}`,
		`{"scope":"day","school":"luoshu_feipan","dingju_method":"zhirun"}`,
		`{"scope":"day","school":"jinhan_yujing"}`,
		`{"scope":"quarter","school":"zhuanpan","quarter_rule":"ten_minute_sanyuan"}`,
		`{"scope":"quarter","school":"zhuanpan","quarter_rule":"twelve_minute_ten_division"}`,
		`{"scope":"month","school":"zhuanpan"}`,
		`{"scope":"month","school":"luoshu_feipan"}`,
		`{"scope":"year","school":"zhuanpan"}`,
		`{"scope":"year","school":"luoshu_feipan"}`,
	}
	if len(methods) != 17 {
		t.Fatalf("boundary-date method coverage = %d, want all 17 public method combinations", len(methods))
	}
	for day := 1; day <= 365; day += 15 {
		when := time.Date(2026, 1, 1, 0, 0, 0, 0, time.FixedZone("CST", 8*3600)).
			AddDate(0, 0, day-1).Format(time.RFC3339)
		for _, method := range methods {
			params := []byte(`{"solar_time":"` + when + `",` + method[1:])
			output := executeAndDecode(t, reg, "qimen.chart", params)
			if err := schema.Validate(output); err != nil {
				t.Fatalf("%s %s: %v", when, method, err)
			}
			assertJiaStaysHidden(t, output["data"].(map[string]any))
		}
	}
}

func requestedQuarterRule(t *testing.T, params string) string {
	t.Helper()
	var requested struct {
		QuarterRule string `json:"quarter_rule"`
	}
	decodeJSON(t, []byte(params), &requested)
	return requested.QuarterRule
}

func TestQimenResultSchemaRejectsMethodMatrixDrift(t *testing.T) {
	reg := NewRPCRegistry()
	schema := compileQimenResultSchema(t)
	result := func(params string) map[string]any {
		t.Helper()
		return executeAndDecode(t, reg, "qimen.chart", []byte(params))
	}
	yearFly := result(`{"solar_time":"2026-03-02T18:30:00+08:00","scope":"year","school":"luoshu_feipan"}`)
	hourRotate := result(`{"solar_time":"2026-03-02T18:30:00+08:00","scope":"hour","school":"zhuanpan","dingju_method":"chaibu"}`)
	dayRotate := result(`{"solar_time":"2026-03-02T18:30:00+08:00","scope":"day","school":"zhuanpan","dingju_method":"chaibu"}`)
	tenMinute := result(`{"solar_time":"2026-09-07T00:00:00+08:00","scope":"quarter","school":"zhuanpan"}`)
	twelveMinute := result(`{"solar_time":"2002-10-08T18:27:00+08:00","scope":"quarter","school":"zhuanpan","quarter_rule":"twelve_minute_ten_division"}`)

	yearWithJu := cloneResult(t, yearFly)
	yearWithJu["data"].(map[string]any)["method"].(map[string]any)["dingju_method"] = "chaibu"
	if err := schema.Validate(yearWithJu); err == nil {
		t.Fatal("schema accepted dingju method on a fixed-calendar year chart")
	}

	flyWithTianQin := cloneResult(t, yearFly)
	flyWithTianQin["data"].(map[string]any)["method"].(map[string]any)["tian_qin"] =
		hourRotate["data"].(map[string]any)["method"].(map[string]any)["tian_qin"]
	if err := schema.Validate(flyWithTianQin); err == nil {
		t.Fatal("schema accepted tian_qin on a flying chart")
	}

	rotateWithoutTianQin := cloneResult(t, hourRotate)
	delete(rotateWithoutTianQin["data"].(map[string]any)["method"].(map[string]any), "tian_qin")
	if err := schema.Validate(rotateWithoutTianQin); err == nil {
		t.Fatal("schema accepted a rotating chart without tian_qin")
	}

	chaibuWithState := cloneResult(t, hourRotate)
	chaibuWithState["data"].(map[string]any)["method"].(map[string]any)["zhi_run_state"] = "正授"
	if err := schema.Validate(chaibuWithState); err == nil {
		t.Fatal("schema accepted zhi_run_state on a chaibu chart")
	}

	dayWithWuBu := cloneResult(t, dayRotate)
	dayWithWuBu["data"].(map[string]any)["pan"].(map[string]any)["wu_bu_yu_shi"] = true
	if err := schema.Validate(dayWithWuBu); err == nil {
		t.Fatal("schema accepted wu_bu_yu_shi=true on a day chart")
	}

	tenMinuteWithHourLead := cloneResult(t, tenMinute)
	tenMinuteWithHourLead["data"].(map[string]any)["method"].(map[string]any)["quarter"].(map[string]any)["lead_pillar_mode"] = "hour"
	if err := schema.Validate(tenMinuteWithHourLead); err == nil {
		t.Fatal("schema accepted a ten-minute quarter with hour lead mode")
	}

	twelveMinuteWithQuarterLead := cloneResult(t, twelveMinute)
	twelveMinuteWithQuarterLead["data"].(map[string]any)["method"].(map[string]any)["quarter"].(map[string]any)["lead_pillar_mode"] = "quarter"
	if err := schema.Validate(twelveMinuteWithQuarterLead); err == nil {
		t.Fatal("schema accepted 十二分钟十分局 quarter with quarter lead mode")
	}

	twelveMinuteWithTenMinutes := cloneResult(t, twelveMinute)
	twelveMinuteWithTenMinutes["data"].(map[string]any)["method"].(map[string]any)["quarter"].(map[string]any)["minutes_per_quarter"] = 10
	if err := schema.Validate(twelveMinuteWithTenMinutes); err == nil {
		t.Fatal("schema accepted 十二分钟十分局 quarter with ten-minute quarters")
	}

	tenMinuteWithTwelveMinuteBase := cloneResult(t, tenMinute)
	tenMinuteMethod := tenMinuteWithTwelveMinuteBase["data"].(map[string]any)["method"].(map[string]any)
	tenMinuteMethod["base_dingju_method"] = "zhirun"
	tenMinuteMethod["base_dingju_method_name"] = "置闰定局"
	if err := schema.Validate(tenMinuteWithTwelveMinuteBase); err == nil {
		t.Fatal("schema accepted a ten-minute quarter with a twelve-minute base method")
	}

	tenMinuteWithTwelveMinuteOptions := cloneResult(t, tenMinute)
	tenMinuteOptions := tenMinuteWithTwelveMinuteOptions["data"].(map[string]any)["method"].(map[string]any)["quarter"].(map[string]any)
	tenMinuteOptions["dun_source"] = "solar_term"
	tenMinuteOptions["hour_boundary"] = "zi_zheng"
	if err := schema.Validate(tenMinuteWithTwelveMinuteOptions); err == nil {
		t.Fatal("schema accepted a ten-minute quarter with twelve-minute options")
	}
}

func TestQimenResultSchemaRejectsSchoolDrift(t *testing.T) {
	reg := NewRPCRegistry()
	schema := compileQimenResultSchema(t)
	result := func(params string) map[string]any {
		t.Helper()
		return executeAndDecode(t, reg, "qimen.chart", []byte(params))
	}
	rotate := result(`{"solar_time":"2026-03-02T18:30:00+08:00","scope":"hour","school":"zhuanpan","dingju_method":"chaibu"}`)
	luoShu := result(`{"solar_time":"2026-03-02T18:30:00+08:00","scope":"hour","school":"luoshu_feipan","dingju_method":"chaibu"}`)
	mingFa := result(`{"solar_time":"2026-08-17T22:00:00+08:00","scope":"hour","school":"mingfa_feipan"}`)

	for name, base := range map[string]map[string]any{
		"rotating": rotate,
		"luo-shu":  luoShu,
	} {
		withCenterDoor := cloneResult(t, base)
		withCenterDoor["data"].(map[string]any)["pan"].(map[string]any)["gong_wei"].([]any)[0].(map[string]any)["men"] = "中门"
		if err := schema.Validate(withCenterDoor); err == nil {
			t.Fatalf("schema accepted a center door on a non-Ming Fa %s chart", name)
		}

		withHiddenPillar := cloneResult(t, base)
		withHiddenPillar["data"].(map[string]any)["pan"].(map[string]any)["gong_wei"].([]any)[0].(map[string]any)["an_gan_zhi"] = "甲子"
		if err := schema.Validate(withHiddenPillar); err == nil {
			t.Fatalf("schema accepted an_gan_zhi on a non-Ming Fa %s chart", name)
		}

		withCenterDutyDoor := cloneResult(t, base)
		withCenterDutyDoor["data"].(map[string]any)["pan"].(map[string]any)["zhi_shi_men"] = "中门"
		if err := schema.Validate(withCenterDutyDoor); err == nil {
			t.Fatalf("schema accepted a center duty door on a non-Ming Fa %s chart", name)
		}

		withCenterDoorInteraction := cloneResult(t, base)
		interactions := withCenterDoorInteraction["data"].(map[string]any)["men_interaction"].([]any)
		if len(interactions) == 0 {
			interactions = append(interactions, map[string]any{
				"door": "中门", "gong": "坎", "name": "invalid", "meaning": "invalid",
			})
			withCenterDoorInteraction["data"].(map[string]any)["men_interaction"] = interactions
		} else {
			interactions[0].(map[string]any)["door"] = "中门"
		}
		if err := schema.Validate(withCenterDoorInteraction); err == nil {
			t.Fatalf("schema accepted a center door interaction on a non-Ming Fa %s chart", name)
		}
	}

	for palace := 0; palace < 9; palace++ {
		withoutHiddenPillar := cloneResult(t, mingFa)
		delete(withoutHiddenPillar["data"].(map[string]any)["pan"].(map[string]any)["gong_wei"].([]any)[palace].(map[string]any), "an_gan_zhi")
		if err := schema.Validate(withoutHiddenPillar); err == nil {
			t.Fatalf("schema accepted Ming Fa palace %d without an_gan_zhi", palace+1)
		}
	}

	dayMingFa := cloneResult(t, mingFa)
	dayMingFa["data"].(map[string]any)["method"].(map[string]any)["scope"] = "day"
	if err := schema.Validate(dayMingFa); err == nil {
		t.Fatal("schema accepted Ming Fa with scope=day")
	}

	zhiRunMingFa := cloneResult(t, mingFa)
	zhiRunMingFa["data"].(map[string]any)["method"].(map[string]any)["dingju_method"] = "zhirun"
	if err := schema.Validate(zhiRunMingFa); err == nil {
		t.Fatal("schema accepted Ming Fa with dingju_method=zhirun")
	}
}

func TestQimenMaoShanHandlerAndResultContract(t *testing.T) {
	reg := NewRPCRegistry()
	schema := compileQimenResultSchema(t)
	output := executeAndDecode(t, reg, "qimen.chart", []byte(
		`{"solar_time":"2024-06-21T06:00:00+08:00","scope":"hour","school":"zhuanpan","dingju_method":"maoshan"}`,
	))
	if err := schema.Validate(output); err != nil {
		t.Fatalf("result schema: %v", err)
	}
	method := output["data"].(map[string]any)["method"].(map[string]any)
	if method["dingju_method"] != "maoshan" || method["dingju_method_name"] != "茅山定局" ||
		method["jie_qi"] != "夏至" || method["yong_ju_jie_qi"] != "夏至" {
		t.Fatalf("Mao Shan method = %+v", method)
	}

	assertError(t, reg, "qimen.chart", []byte(`{"solar_time":"2024-06-21T06:00:00+08:00","scope":"day","school":"zhuanpan","dingju_method":"maoshan"}`))
	assertError(t, reg, "qimen.chart", []byte(`{"solar_time":"2024-06-21T06:00:00+08:00","scope":"hour","school":"luoshu_feipan","dingju_method":"maoshan"}`))
	assertError(t, reg, "qimen.chart", []byte(`{"solar_time":"2024-06-21T06:00:00+08:00","scope":"hour","school":"mingfa_feipan","dingju_method":"maoshan"}`))
}

func TestQimenQuarterHandlerAndResultContract(t *testing.T) {
	reg := NewRPCRegistry()
	schema := compileQimenResultSchema(t)
	output := executeAndDecode(t, reg, "qimen.chart", []byte(
		`{"solar_time":"2026-09-07T00:00:00+08:00","scope":"quarter","school":"zhuanpan"}`,
	))
	if err := schema.Validate(output); err != nil {
		t.Fatalf("result schema: %v", err)
	}
	data := output["data"].(map[string]any)
	method := data["method"].(map[string]any)
	if method["scope"] != "quarter" || method["scope_name"] != "刻家奇门" ||
		method["school"] != "zhuanpan" || method["quarter_rule"] != "ten_minute_sanyuan" ||
		method["quarter_rule_name"] != "十分钟三元刻家" || method["method_source"] != "quarter_rule" {
		t.Fatalf("quarter method = %+v", method)
	}
	quarter := method["quarter"].(map[string]any)
	index, _ := quarter["index"].(float64)
	minutes, _ := quarter["minutes_per_quarter"].(float64)
	if index != 7 || minutes != 10 {
		t.Fatalf("quarter metadata = %+v", quarter)
	}
	if lead := method["lead_pillar"].(map[string]any)["name"]; lead != "庚午" {
		t.Fatalf("quarter lead = %v, want 庚午", lead)
	}
	if mode := quarter["lead_pillar_mode"]; mode != "quarter" {
		t.Fatalf("quarter lead mode = %v, want quarter", mode)
	}

	twelveMinute := executeAndDecode(t, reg, "qimen.chart", []byte(
		`{"solar_time":"2002-10-08T18:27:00+08:00","scope":"quarter","school":"zhuanpan","quarter_rule":"twelve_minute_ten_division"}`,
	))
	if err := schema.Validate(twelveMinute); err != nil {
		t.Fatalf("十二分钟十分局 result schema: %v", err)
	}
	twelveMinuteMethod := twelveMinute["data"].(map[string]any)["method"].(map[string]any)
	if twelveMinuteMethod["quarter_rule_name"] != "十二分钟十分局刻家" ||
		twelveMinuteMethod["method_source"] != "quarter_rule" {
		t.Fatalf("十二分钟十分局 quarter method = %+v", twelveMinuteMethod)
	}
	twelveMinuteQuarter := twelveMinuteMethod["quarter"].(map[string]any)
	twelveMinuteIndex, _ := twelveMinuteQuarter["index"].(float64)
	twelveMinuteMinutes, _ := twelveMinuteQuarter["minutes_per_quarter"].(float64)
	if twelveMinuteIndex != 8 || twelveMinuteMinutes != 12 || twelveMinuteQuarter["lead_pillar_mode"] != "hour" {
		t.Fatalf("十二分钟十分局 quarter metadata = %+v", twelveMinuteQuarter)
	}
	if lead := twelveMinuteMethod["lead_pillar"].(map[string]any)["name"]; lead != "癸酉" {
		t.Fatalf("十二分钟十分局 quarter lead = %v, want 癸酉", lead)
	}
	twelveMinuteJu, _ := twelveMinute["data"].(map[string]any)["pan"].(map[string]any)["jushu"].(float64)
	if twelveMinuteJu != 9 {
		t.Fatalf("十二分钟十分局 quarter ju = %v, want 9", twelveMinuteJu)
	}

	zhiRunBase := executeAndDecode(t, reg, "qimen.chart", []byte(
		`{"solar_time":"2026-01-05T00:00:00+08:00","scope":"quarter","school":"zhuanpan","quarter_rule":"twelve_minute_ten_division","base_dingju_method":"zhirun"}`,
	))
	if err := schema.Validate(zhiRunBase); err != nil {
		t.Fatalf("十二分钟十分局 zhirun result schema: %v", err)
	}
	zhiRunMethod := zhiRunBase["data"].(map[string]any)["method"].(map[string]any)
	if zhiRunMethod["base_dingju_method"] != "zhirun" || zhiRunMethod["base_dingju_method_name"] != "置闰定局" {
		t.Fatalf("十二分钟十分局 zhirun base method = %+v", zhiRunMethod)
	}
	if ju := zhiRunBase["data"].(map[string]any)["pan"].(map[string]any)["jushu"].(float64); ju != 7 {
		t.Fatalf("十二分钟十分局 zhirun quarter ju = %v, want 7", ju)
	}

	options := executeAndDecode(t, reg, "qimen.chart", []byte(
		`{"solar_time":"2026-09-07T00:00:00+08:00","scope":"quarter","school":"zhuanpan","quarter_rule":"twelve_minute_ten_division","dun_source":"solar_term","hour_boundary":"zi_zheng"}`,
	))
	if err := schema.Validate(options); err != nil {
		t.Fatalf("十二分钟十分局 options result schema: %v", err)
	}
	optionsQuarter := options["data"].(map[string]any)["method"].(map[string]any)["quarter"].(map[string]any)
	if optionsQuarter["dun_source"] != "solar_term" || optionsQuarter["hour_boundary"] != "zi_zheng" {
		t.Fatalf("十二分钟十分局 option metadata = %+v", optionsQuarter)
	}
	if ju := options["data"].(map[string]any)["pan"].(map[string]any)["jushu"].(float64); ju != 4 {
		t.Fatalf("十二分钟十分局 option quarter ju = %v, want 4", ju)
	}

	for _, params := range []string{
		`{"solar_time":"2026-09-07T00:00:00+08:00","scope":"quarter","school":"zhuanpan","dingju_method":"chaibu"}`,
		`{"solar_time":"2026-09-07T00:00:00+08:00","scope":"quarter","school":"luoshu_feipan"}`,
		`{"solar_time":"2026-09-07T00:00:00+08:00","scope":"quarter","school":"mingfa_feipan"}`,
		`{"solar_time":"2026-09-07T00:00:00+08:00","scope":"hour","school":"zhuanpan","quarter_rule":"ten_minute_sanyuan"}`,
		`{"solar_time":"2026-09-07T00:00:00+08:00","scope":"hour","school":"zhuanpan","quarter_rule":"twelve_minute_ten_division"}`,
		`{"solar_time":"2026-09-07T00:00:00+08:00","scope":"quarter","school":"luoshu_feipan","quarter_rule":"twelve_minute_ten_division"}`,
		`{"solar_time":"2026-01-05T00:00:00+08:00","scope":"hour","school":"zhuanpan","dingju_method":"chaibu","base_dingju_method":"zhirun"}`,
		`{"solar_time":"2026-01-05T00:00:00+08:00","scope":"hour","school":"zhuanpan","dingju_method":"chaibu","dun_source":"solar_term"}`,
		`{"solar_time":"2026-01-05T00:00:00+08:00","scope":"hour","school":"zhuanpan","dingju_method":"chaibu","hour_boundary":"zi_zheng"}`,
		`{"solar_time":"2026-01-05T00:00:00+08:00","scope":"quarter","school":"zhuanpan","quarter_rule":"twelve_minute_ten_division","base_dingju_method":"ten_minute_sanyuan"}`,
	} {
		assertError(t, reg, "qimen.chart", []byte(params))
	}
}

func TestQimenJinhanHandlerAndResultContract(t *testing.T) {
	reg := NewRPCRegistry()
	schema := compileQimenResultSchema(t)
	output := executeAndDecode(t, reg, "qimen.chart", []byte(
		`{"solar_time":"2024-06-29T12:00:00+08:00","scope":"day","school":"jinhan_yujing"}`,
	))
	if err := schema.Validate(output); err != nil {
		t.Fatalf("result schema: %v", err)
	}
	data := output["data"].(map[string]any)
	method := data["method"].(map[string]any)
	if method["scope"] != "day" || method["school"] != "jinhan_yujing" ||
		method["dingju_method"] != "none" || method["dingju_method_name"] != "无定局" ||
		method["dun_source"] != "winter_solstice_yang_summer_solstice_yin" {
		t.Fatalf("Jin Han method = %+v", method)
	}
	pan := data["pan"].(map[string]any)
	if pan["yin_dun"] != true || pan["ri_gan"] != "甲" || pan["ri_zhi"] != "子" {
		t.Fatalf("Jin Han pan header = %+v", pan)
	}
	for _, field := range []string{
		"jushu", "zhi_fu_xing", "zhi_shi_men", "kong_wang", "ma_xing", "wu_bu_yu_shi",
	} {
		if _, exists := pan[field]; exists {
			t.Fatalf("Jin Han pan must not expose standard field %q", field)
		}
	}
	if _, exists := pan["day_spirits"]; !exists {
		t.Fatal("Jin Han pan lacks day_spirits")
	}
	for _, field := range []string{
		"gan_interaction", "men_interaction", "xing_interaction", "xing_gong_wu_xing",
		"men_po", "men_zhi", "patterns", "ying_qi", "ri_gan_gong", "shi_gan_gong",
		"kong_wang_affected", "ma_xing_affected", "zhi_fu_xing_gong", "zhi_shi_men_gong",
		"yong_shen",
	} {
		if _, exists := data[field]; exists {
			t.Fatalf("Jin Han chart must not expose standard analysis field %q", field)
		}
	}

	assertError(t, reg, "qimen.chart", []byte(`{"solar_time":"2024-06-29T12:00:00+08:00","scope":"hour","school":"jinhan_yujing"}`))
	assertError(t, reg, "qimen.chart", []byte(`{"solar_time":"2024-06-29T12:00:00+08:00","scope":"day","school":"jinhan_yujing","dingju_method":"none"}`))
	assertError(t, reg, "qimen.chart", []byte(`{"solar_time":"2024-06-29T12:00:00+08:00","scope":"day","school":"jinhan_yujing","yong_shen":["太乙"]}`))
}

func TestQimenConsumesTianwenSolarTime(t *testing.T) {
	reg := NewRPCRegistry()
	timeResult := executeAndDecode(t, reg, "tianwen.time", mustJSON(t, map[string]any{
		"time": "2026-03-02T18:30:00+08:00", "longitude": 116.4,
	}))
	solar, ok := timeResult["data"].(map[string]any)["solar"].(string)
	if !ok || solar == "" {
		t.Fatalf("tianwen.time solar = %#v", timeResult["data"])
	}
	params := mustJSON(t, map[string]any{"solar_time": solar})
	result := executeAndDecode(t, reg, "qimen.chart", params)
	if result["_product"] != "qimen" || result["data"] == nil {
		t.Fatalf("qimen result from tianwen solar = %#v", result)
	}
}

func cloneResult(t *testing.T, value map[string]any) map[string]any {
	t.Helper()
	encoded, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	var result map[string]any
	decodeJSON(t, encoded, &result)
	return result
}

func assertJiaStaysHidden(t *testing.T, data map[string]any) {
	t.Helper()
	pan := data["pan"].(map[string]any)
	for _, item := range pan["gong_wei"].([]any) {
		palace := item.(map[string]any)
		for _, field := range []string{"di_pan_gan", "an_gan", "tian_pan_gan"} {
			if palace[field] == "甲" {
				t.Fatalf("%s exposes 甲", field)
			}
		}
		if tianPan, ok := palace["tian_pan"].([]any); ok {
			for _, item := range tianPan {
				if item.(map[string]any)["gan"] == "甲" {
					t.Fatal("tian_pan exposes 甲")
				}
			}
		}
	}
	if interactions, ok := data["gan_interaction"].([]any); ok {
		for _, item := range interactions {
			interaction := item.(map[string]any)
			if interaction["di_pan_gan"] == "甲" || interaction["tian_pan_gan"] == "甲" {
				t.Fatal("gan_interaction exposes 甲")
			}
		}
	}
	if yongShen, ok := data["yong_shen"].(map[string]any); ok {
		for _, item := range yongShen["symbols"].([]any) {
			for _, gan := range item.(map[string]any)["tian_gan"].([]any) {
				if gan == "甲" {
					t.Fatal("yong_shen tian_gan exposes 甲")
				}
			}
		}
	}
}

type qimenParamsContract struct {
	AdditionalProperties *bool          `json:"additionalProperties"`
	Required             []string       `json:"required"`
	Properties           map[string]any `json:"properties"`
	AllOf                []any          `json:"allOf"`
}

type qimenObjectContract struct {
	Properties map[string]any `json:"properties"`
	Required   []string       `json:"required"`
}

func qimenMethodSchema(t *testing.T) map[string]any {
	t.Helper()
	var document struct {
		Methods []map[string]any `json:"methods"`
	}
	decodeJSON(t, NewRPCRegistry().OpenRPCDocument(), &document)
	for _, method := range document.Methods {
		if method["name"] == "qimen.chart" {
			return method
		}
	}
	t.Fatal("qimen.chart missing")
	return nil
}

func qimenParamsSchema(t *testing.T) qimenParamsContract {
	t.Helper()
	var params qimenParamsContract
	decodeAny(t, qimenMethodSchema(t)["params"], &params)
	return params
}

func qimenDataSchema(t *testing.T) qimenObjectContract {
	t.Helper()
	var result struct {
		Result struct {
			Properties struct {
				Data json.RawMessage `json:"data"`
			} `json:"properties"`
		} `json:"result"`
	}
	decodeAny(t, qimenMethodSchema(t), &result)
	var union struct {
		OneOf []json.RawMessage `json:"oneOf"`
	}
	if err := json.Unmarshal(result.Result.Properties.Data, &union); err != nil {
		t.Fatal(err)
	}
	if len(union.OneOf) == 0 {
		t.Fatal("qimen result data lacks standard/Jin Han union branches")
	}
	var data qimenObjectContract
	decodeJSON(t, union.OneOf[0], &data)
	return data
}

func compileQimenParamsSchema(t *testing.T) *jsonschema.Schema {
	t.Helper()
	compiler := jsonschema.NewCompiler()
	if err := compiler.AddResource("schema.json", qimenMethodSchema(t)["params"]); err != nil {
		t.Fatal(err)
	}
	schema, err := compiler.Compile("schema.json")
	if err != nil {
		t.Fatal(err)
	}
	return schema
}

func compileQimenResultSchema(t *testing.T) *jsonschema.Schema {
	t.Helper()
	compiler := jsonschema.NewCompiler()
	if err := compiler.AddResource("schema.json", qimenMethodSchema(t)["result"]); err != nil {
		t.Fatal(err)
	}
	schema, err := compiler.Compile("schema.json")
	if err != nil {
		t.Fatal(err)
	}
	return schema
}

func decodeAny(t *testing.T, value any, target any) {
	t.Helper()
	encoded, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(encoded, target); err != nil {
		t.Fatal(err)
	}
}

func assertEnum(t *testing.T, schema any, want []any) {
	t.Helper()
	got := schema.(map[string]any)["enum"].([]any)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("enum = %v, want %v", got, want)
	}
}

func collectKeys(value map[string]any, output map[string]any) {
	for key, child := range value {
		output[key] = child
	}
}

func assertSameKeys(t *testing.T, actual, expected map[string]any) {
	t.Helper()
	actualKeys := make([]string, 0, len(actual))
	for key := range actual {
		actualKeys = append(actualKeys, key)
	}
	expectedKeys := make([]string, 0, len(expected))
	for key := range expected {
		expectedKeys = append(expectedKeys, key)
	}
	sort.Strings(actualKeys)
	sort.Strings(expectedKeys)
	if !reflect.DeepEqual(actualKeys, expectedKeys) {
		t.Fatalf("keys = %v, want %v", actualKeys, expectedKeys)
	}
}
