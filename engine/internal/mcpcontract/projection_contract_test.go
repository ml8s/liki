package mcpcontract

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// These assertions used to live in tests/test_engine_snapshot_sync.py, which
// parsed the legacy raw-JSON Result schemas out of tools_bazi.go/tools_ziwei.go.
// The typed registry now owns those schemas, so the projection contract is
// checked here against the DTOs directly. The remaining Python tests in that
// file only inspect the contract document itself.

const projectionContractPath = "../../../counsel/app/natal/tools/natal_projection_contract.json"

type projectionContract struct {
	Pillars     []string          `json:"四柱"`
	ValueFields map[string]string `json:"柱值字段"`
	ExistFields map[string]string `json:"柱存在字段"`
	Bazi        struct {
		Full  map[string]string `json:"full"`
		Chart map[string]string `json:"chart"`
	} `json:"八字字段映射"`
	Ziwei struct {
		Ziwei map[string]string `json:"ziwei"`
		Board map[string]string `json:"board"`
	} `json:"紫微字段映射"`
}

func loadProjectionContract(t *testing.T) projectionContract {
	t.Helper()
	raw, err := os.ReadFile(filepath.FromSlash(projectionContractPath))
	if err != nil {
		t.Fatalf("read projection contract: %v", err)
	}
	var contract projectionContract
	if err := json.Unmarshal(raw, &contract); err != nil {
		t.Fatalf("decode projection contract: %v", err)
	}
	return contract
}

// toolProperties returns the declared properties of a tool's output schema.
func toolProperties(t *testing.T, tool string) map[string]any {
	t.Helper()
	return outputProperties(t, testEngine(), tool)
}

func TestProjectionContractCoversEngineFullChartFields(t *testing.T) {
	contract := loadProjectionContract(t)
	full := toolProperties(t, "bazi_fullchart")

	pillarFields := map[string]bool{}
	for name := range contract.ValueFields {
		pillarFields[name] = true
	}
	for name := range contract.ExistFields {
		pillarFields[name] = true
	}
	mappedFull := map[string]bool{}
	for name := range contract.Bazi.Full {
		mappedFull[name] = true
	}
	for name := range contract.Bazi.Chart {
		mappedFull[name] = true
	}

	// gender is projected into context; birth_year is available through chart
	// mapping. Other new top-level engine facts must be added to the projection
	// mapping deliberately.
	allowedUnprojected := map[string]bool{"gender": true}
	pillars := map[string]bool{}
	for _, pillar := range contract.Pillars {
		pillars[pillar] = true
	}

	for name := range full {
		if pillars[name] || mappedFull[name] || allowedUnprojected[name] {
			continue
		}
		t.Errorf("engine fullchart field %q is not in the projection contract", name)
	}

	for _, pillar := range contract.Pillars {
		pillarSchema, ok := full[pillar].(map[string]any)
		if !ok {
			t.Fatalf("fullchart.%s schema = %#v", pillar, full[pillar])
		}
		properties, _ := pillarSchema["properties"].(map[string]any)
		for name := range properties {
			if !pillarFields[name] {
				t.Errorf("fullchart.%s.%s is not in the projection contract", pillar, name)
			}
		}
	}
}

func TestProjectionContractPathsExistInEngineSchema(t *testing.T) {
	contract := loadProjectionContract(t)
	full := toolProperties(t, "bazi_fullchart")
	chart := toolProperties(t, "bazi_chart")
	ziwei := toolProperties(t, "ziwei_chart")

	var missing []string
	for _, pillar := range contract.Pillars {
		pillarSchema, ok := full[pillar].(map[string]any)
		if !ok {
			missing = append(missing, "full."+pillar)
			continue
		}
		properties, _ := pillarSchema["properties"].(map[string]any)
		for source, label := range contract.ValueFields {
			if _, ok := properties[source]; !ok {
				missing = append(missing, "full."+pillar+"."+source+" ("+label+")")
			}
		}
		for source, label := range contract.ExistFields {
			if _, ok := properties[source]; !ok {
				missing = append(missing, "full."+pillar+"."+source+" ("+label+")")
			}
		}
	}
	for source, label := range contract.Bazi.Full {
		if _, ok := full[source]; !ok {
			missing = append(missing, "full."+source+" ("+label+")")
		}
	}
	for source, label := range contract.Bazi.Chart {
		if _, ok := chart[source]; !ok {
			missing = append(missing, "chart."+source+" ("+label+")")
		}
	}
	for source, label := range contract.Ziwei.Ziwei {
		if _, ok := ziwei[source]; !ok {
			missing = append(missing, "ziwei."+source+" ("+label+")")
		}
	}

	if len(missing) > 0 {
		t.Errorf("projection contract references missing engine fields: %v", missing)
	}
}
