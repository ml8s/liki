package mcpcontract

import (
	"encoding/json"
	"fmt"
)

func withSchemaConst(raw json.RawMessage, field, value string) (json.RawMessage, error) {
	return withSchemaPropertyConstraints(raw, field, &value, nil)
}

func withSchemaEnum(raw json.RawMessage, field string, values ...string) (json.RawMessage, error) {
	return withSchemaPropertyConstraints(raw, field, nil, values)
}

func withSchemaOptionalDefault(raw json.RawMessage, field, value string) (json.RawMessage, error) {
	var document map[string]any
	if err := json.Unmarshal(raw, &document); err != nil {
		return nil, err
	}
	properties, ok := document["properties"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("schema has no properties")
	}
	property, ok := properties[field].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("schema has no property %q", field)
	}
	property["default"] = value
	required := make([]string, 0, len(document["required"].([]any)))
	for _, name := range document["required"].([]any) {
		if name.(string) != field {
			required = append(required, name.(string))
		}
	}
	if len(required) == 0 {
		delete(document, "required")
	} else {
		document["required"] = required
	}
	return json.Marshal(document)
}

func withSchemaFormMeta(raw json.RawMessage, title, description string) (json.RawMessage, error) {
	var document map[string]any
	if err := json.Unmarshal(raw, &document); err != nil {
		return nil, err
	}
	document["title"] = title
	document["description"] = description
	return json.Marshal(document)
}

func withSchemaNumberBounds(raw json.RawMessage, field string, minimum, maximum float64) (json.RawMessage, error) {
	var document map[string]any
	if err := json.Unmarshal(raw, &document); err != nil {
		return nil, err
	}
	properties, ok := document["properties"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("schema has no properties")
	}
	property, ok := properties[field].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("schema has no property %q", field)
	}
	property["minimum"] = minimum
	property["maximum"] = maximum
	return json.Marshal(document)
}

func withSchemaPropertyConstraints(
	raw json.RawMessage, field string, constant *string, enum []string,
) (json.RawMessage, error) {
	var document map[string]any
	if err := json.Unmarshal(raw, &document); err != nil {
		return nil, err
	}
	properties, ok := document["properties"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("schema has no properties")
	}
	property, ok := properties[field].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("schema has no property %q", field)
	}
	if constant != nil {
		delete(property, "enum")
		property["const"] = *constant
	}
	if enum != nil {
		delete(property, "const")
		values := make([]any, 0, len(enum))
		for _, value := range enum {
			values = append(values, value)
		}
		property["enum"] = values
	}
	return json.Marshal(document)
}

func withSchemaProperty(raw json.RawMessage, field string, replacement json.RawMessage) (json.RawMessage, error) {
	var document map[string]any
	if err := json.Unmarshal(raw, &document); err != nil {
		return nil, err
	}
	properties, ok := document["properties"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("schema has no properties")
	}
	if _, ok := properties[field]; !ok {
		return nil, fmt.Errorf("schema has no property %q", field)
	}
	var property any
	if err := json.Unmarshal(replacement, &property); err != nil {
		return nil, err
	}
	properties[field] = property
	return json.Marshal(document)
}

func withoutSchemaProperty(raw json.RawMessage, field string) (json.RawMessage, error) {
	var document map[string]any
	if err := json.Unmarshal(raw, &document); err != nil {
		return nil, err
	}
	if properties, ok := document["properties"].(map[string]any); ok {
		delete(properties, field)
	}
	if required, ok := document["required"].([]any); ok {
		kept := make([]any, 0, len(required))
		for _, name := range required {
			if text, ok := name.(string); !ok || text != field {
				kept = append(kept, name)
			}
		}
		if len(kept) == 0 {
			delete(document, "required")
		} else {
			document["required"] = kept
		}
	}
	return json.Marshal(document)
}

func withSchemaRequiredFields(raw json.RawMessage, fields ...string) (json.RawMessage, error) {
	var document map[string]any
	if err := json.Unmarshal(raw, &document); err != nil {
		return nil, err
	}
	existing := make([]string, 0, len(fields))
	seen := map[string]struct{}{}
	if required, ok := document["required"].([]any); ok {
		for _, name := range required {
			text, ok := name.(string)
			if !ok {
				continue
			}
			if _, duplicate := seen[text]; duplicate {
				continue
			}
			seen[text] = struct{}{}
			existing = append(existing, text)
		}
	}
	for _, field := range fields {
		if _, duplicate := seen[field]; duplicate {
			continue
		}
		seen[field] = struct{}{}
		existing = append(existing, field)
	}
	document["required"] = existing
	return json.Marshal(document)
}

// withSchemaEitherRequired requires exactly one of two optional properties.
func withSchemaEitherRequired(raw json.RawMessage, first, second string) (json.RawMessage, error) {
	var document map[string]any
	if err := json.Unmarshal(raw, &document); err != nil {
		return nil, err
	}
	document["oneOf"] = []any{
		map[string]any{
			"required":   []any{first},
			"properties": map[string]any{first: map[string]any{"type": "string"}},
		},
		map[string]any{
			"required":   []any{second},
			"properties": map[string]any{second: map[string]any{"type": "object"}},
		},
	}
	return json.Marshal(document)
}
