package main

import (
	"encoding/json"
	"fmt"

	"liki-engine/internal/engine/qiming"
)

type caseOut struct {
	Name string      `json:"name"`
	Out  interface{} `json:"out"`
}

func main() {
	var results []caseOut

	// pick
	pcases := []struct {
		name, w1, w2 string
		count        int
	}{
		{"pick_mu_huo_2", "木", "火", 2},
		{"pick_shui_1", "水", "", 1},
		{"pick_jin_shui_2", "金", "水", 2},
	}
	for _, c := range pcases {
		out, err := qiming.PickChars(c.w1, c.w2, c.count)
		if err != nil {
			results = append(results, caseOut{c.name, map[string]string{"error": err.Error()}})
			continue
		}
		results = append(results, caseOut{c.name, out})
	}

	// compose
	ccases := []struct {
		name             string
		first, second    []string
		maxNames         int
	}{
		{"compose_double_5", []string{"子", "轩"}, []string{"然", "宇"}, 5},
		{"compose_single_3", []string{"安", "宁", "静"}, nil, 3},
	}
	for _, c := range ccases {
		out, err := qiming.ComposeNames(qiming.ComposeRequest{First: c.first, Second: c.second, MaxNames: c.maxNames})
		if err != nil {
			results = append(results, caseOut{c.name, map[string]string{"error": err.Error()}})
			continue
		}
		results = append(results, caseOut{c.name, out})
	}

	// evaluate
	ecases := []struct {
		name, yong string
		names      []string
		xi, ji     []string
	}{
		{"evaluate_zixuan", "木", []string{"子轩", "浩然"}, []string{"火"}, []string{"金"}},
		{"evaluate_invalid", "木", []string{"子轩", "死", "x", "子"}, []string{"火"}, []string{"金"}},
		{"evaluate_no_constraint", "", []string{"云"}, nil, nil},
	}
	for _, c := range ecases {
		out, err := qiming.EvaluateNames(c.names, c.yong, c.xi, c.ji)
		if err != nil {
			results = append(results, caseOut{c.name, map[string]string{"error": err.Error()}})
			continue
		}
		results = append(results, caseOut{c.name, out})
	}

	// match_surnames
	scases := []struct {
		name, source string
		max          int
	}{
		{"surname_zhang_3", "zhang", 3},
		{"surname_liu_4", "liu", 4},
		{"surname_qian_2", "qian", 2},
		{"surname_fallback_x", "zyxwv", 3},
	}
	for _, c := range scases {
		out, err := qiming.MatchSurnames(c.source, c.max)
		if err != nil {
			results = append(results, caseOut{c.name, map[string]string{"error": err.Error()}})
			continue
		}
		results = append(results, caseOut{c.name, out})
	}

	data, _ := json.MarshalIndent(results, "", "  ")
	fmt.Println(string(data))
}
