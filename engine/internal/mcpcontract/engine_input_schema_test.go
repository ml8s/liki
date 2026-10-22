package mcpcontract

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// The Python qimen orchestrator (tests/test_qimen_tools.py) validates the
// parameters it builds against the engine boundary. It cannot derive the
// typed schema itself, so the qimen.chart input schema is committed as a
// golden file and guarded here: if the DTO changes, this test fails and the
// Python check would otherwise silently validate against a stale contract.
//
// Regenerate with UPDATE_GOLDEN=1 go test ./internal/mcpcontract/.
var updateEngineInputGolden = os.Getenv("UPDATE_GOLDEN") == "1"

const qimenInputSchemaGolden = "testdata/qimen-input-schema.json"

func TestEngineQimenInputSchemaGolden(t *testing.T) {
	definition, ok := testEngine().Definition("qimen_chart")
	if !ok {
		t.Fatal("qimen.chart is not registered")
	}
	pretty, err := indentSchema(definition.InputSchema)
	if err != nil {
		t.Fatalf("render input schema: %v", err)
	}

	if updateEngineInputGolden {
		if err := os.MkdirAll(filepath.Dir(qimenInputSchemaGolden), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(qimenInputSchemaGolden, pretty, 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}

	onDisk, err := os.ReadFile(qimenInputSchemaGolden)
	if err != nil {
		t.Fatalf("read golden: %v — regenerate with UPDATE_GOLDEN=1", err)
	}
	if !bytes.Equal(bytes.TrimSpace(onDisk), bytes.TrimSpace(pretty)) {
		t.Fatalf("qimen.chart input schema golden is stale; regenerate with UPDATE_GOLDEN=1")
	}
}

func indentSchema(raw json.RawMessage) ([]byte, error) {
	var document any
	if err := json.Unmarshal(raw, &document); err != nil {
		return nil, err
	}
	return json.MarshalIndent(document, "", "  ")
}
