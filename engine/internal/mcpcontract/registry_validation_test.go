package mcpcontract

import (
	"context"
	"encoding/json"
	"strings"
	"sync"
	"testing"
)

func TestRegistryValidateOutputAcceptsActualJSON(t *testing.T) {
	registry := testEngine()
	raw, err := registry.Call(context.Background(), "time_now", json.RawMessage(`{"timezone":"Asia/Shanghai"}`))
	if err != nil {
		t.Fatalf("call time_now: %v", err)
	}
	if err := registry.ValidateOutput("time_now", raw); err != nil {
		t.Fatalf("validate actual output: %v", err)
	}
}

func TestRegistryValidateOutputRejectsContractViolation(t *testing.T) {
	registry := testEngine()
	err := registry.ValidateOutput("time_now", json.RawMessage(`{"instant":12,"timezone":""}`))
	if err == nil {
		t.Fatal("expected invalid output document to fail")
	}
	if !strings.Contains(err.Error(), "output contract violation") {
		t.Fatalf("error does not identify output contract violation: %v", err)
	}
}

func TestRegistryValidateOutputRejectsUnknownTool(t *testing.T) {
	err := testEngine().ValidateOutput("missing_tool", json.RawMessage(`{}`))
	if err == nil || !strings.Contains(err.Error(), "unknown tool") {
		t.Fatalf("unknown tool error = %v", err)
	}
}

func TestRegistryMaterializeSchemasIsConcurrencySafe(t *testing.T) {
	registry := NewEngine()
	if err := registry.MaterializeSchemas(); err != nil {
		t.Fatalf("materialize schemas: %v", err)
	}
	var wg sync.WaitGroup
	for _, name := range registry.Names() {
		wg.Add(1)
		go func(name string) {
			defer wg.Done()
			if _, err := registry.Definition(name); err != nil {
				t.Errorf("%s: %v", name, err)
			}
		}(name)
	}
	wg.Wait()
}
