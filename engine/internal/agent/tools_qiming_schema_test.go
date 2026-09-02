package agent

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestQimingSchemaContract(t *testing.T) {
	reg := NewRPCRegistry()
	var document struct {
		Methods []struct {
			Name   string `json:"name"`
			Params struct {
				AdditionalProperties *bool    `json:"additionalProperties"`
				Required             []string `json:"required"`
			} `json:"params"`
			Result struct {
				Properties struct {
					Data json.RawMessage `json:"data"`
				} `json:"properties"`
			} `json:"result"`
		} `json:"methods"`
	}
	if err := json.Unmarshal(reg.OpenRPCDocument(), &document); err != nil {
		t.Fatal(err)
	}

	methods := make(map[string]int, len(document.Methods))
	for i, method := range document.Methods {
		methods[method.Name] = i
	}
	for _, name := range []string{"qiming.char", "qiming.pick", "qiming.compose", "qiming.check"} {
		index, ok := methods[name]
		if !ok {
			t.Fatalf("OpenRPC document missing %s", name)
		}
		if additional := document.Methods[index].Params.AdditionalProperties; additional == nil || *additional {
			t.Errorf("%s params must disable additional properties", name)
		}
	}

	tests := []struct {
		name string
		got  []string
		want []string
	}{
		{"qiming.pick params", document.Methods[methods["qiming.pick"]].Params.Required, []string{"wuxing1"}},
		{"qiming.compose params", document.Methods[methods["qiming.compose"]].Params.Required, []string{"first"}},
		{"qiming.check params", document.Methods[methods["qiming.check"]].Params.Required, []string{"given_names"}},
	}
	for _, test := range tests {
		if !reflect.DeepEqual(test.got, test.want) {
			t.Errorf("%s required = %v, want %v", test.name, test.got, test.want)
		}
	}

	objectSchema := func(name string) map[string]any {
		t.Helper()
		var schema struct {
			AdditionalProperties *bool          `json:"additionalProperties"`
			Properties           map[string]any `json:"properties"`
		}
		data := document.Methods[methods[name]].Result.Properties.Data
		if err := json.Unmarshal(data, &schema); err != nil {
			t.Fatalf("unmarshal %s data schema: %v", name, err)
		}
		if schema.AdditionalProperties == nil || *schema.AdditionalProperties {
			t.Errorf("%s data must disable additional properties", name)
		}
		return schema.Properties
	}

	character := objectSchema("qiming.char")
	if _, exists := character["traditional"]; exists {
		t.Error("qiming.char data must not declare traditional")
	}
	objectSchema("qiming.pick")
	objectSchema("qiming.compose")

	var checkSchema struct {
		Type  string `json:"type"`
		Items struct {
			AdditionalProperties *bool `json:"additionalProperties"`
		} `json:"items"`
	}
	data := document.Methods[methods["qiming.check"]].Result.Properties.Data
	if err := json.Unmarshal(data, &checkSchema); err != nil {
		t.Fatal(err)
	}
	if checkSchema.Type != "array" {
		t.Fatalf("qiming.check data type = %q, want array", checkSchema.Type)
	}
	if checkSchema.Items.AdditionalProperties == nil || *checkSchema.Items.AdditionalProperties {
		t.Error("qiming.check item must disable additional properties")
	}
}
