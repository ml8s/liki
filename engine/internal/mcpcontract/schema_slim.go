package mcpcontract

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"slices"
)

// slimSchemaRaw applies semantics-preserving, publication-only compaction.
// Runtime validation uses the derived schema tree and typed DTOs; the raw wire
// schema is free to use nullable type arrays and shared enum definitions.
func slimSchemaRaw(raw json.RawMessage) (json.RawMessage, error) {
	if len(raw) == 0 {
		return raw, nil
	}
	var schema map[string]any
	if err := json.Unmarshal(raw, &schema); err != nil {
		return nil, fmt.Errorf("decode schema for slimming: %w", err)
	}
	normalizeNullableAnyOf(schema)
	if err := hoistRepeatedEnums(schema); err != nil {
		return nil, err
	}
	slimmed, err := json.Marshal(schema)
	if err != nil {
		return nil, fmt.Errorf("encode slimmed schema: %w", err)
	}
	return json.RawMessage(slimmed), nil
}

// normalizeNullableAnyOf rewrites Pydantic's two-branch null unions into the
// equivalent JSON Schema type array. The null branch must carry no constraints;
// any richer union is left untouched.
func normalizeNullableAnyOf(node any) {
	switch value := node.(type) {
	case map[string]any:
		if branches, ok := value["anyOf"].([]any); ok && len(branches) == 2 {
			var nonNull map[string]any
			nullCount := 0
			for _, branch := range branches {
				object, ok := branch.(map[string]any)
				if !ok {
					continue
				}
				if len(object) == 1 {
					if kind, ok := object["type"].(string); ok && kind == "null" {
						nullCount++
						continue
					}
				}
				nonNull = object
			}
			if nullCount == 1 && nonNull != nil {
				if kind, ok := schemaTypeValue(nonNull["type"]); ok {
					types := []any{"null", kind}
					replacement := make(map[string]any, len(nonNull))
					for key, field := range nonNull {
						if key != "type" {
							replacement[key] = field
						}
					}
					replacement["type"] = types
					for key, field := range value {
						if key != "anyOf" {
							replacement[key] = field
						}
					}
					clear(value)
					for key, field := range replacement {
						value[key] = field
					}
				}
			}
		}
		for _, field := range value {
			normalizeNullableAnyOf(field)
		}
	case []any:
		for _, field := range value {
			normalizeNullableAnyOf(field)
		}
	}
}

func schemaTypeValue(value any) (string, bool) {
	switch kind := value.(type) {
	case string:
		if kind != "null" {
			return kind, true
		}
	case []any:
		var nonNull []any
		for _, item := range kind {
			if item != "null" {
				nonNull = append(nonNull, item)
			}
		}
		if len(nonNull) == 1 {
			if item, ok := nonNull[0].(string); ok && item != "null" {
				return item, true
			}
		}
	}
	return "", false
}

func hoistRepeatedEnums(schema map[string]any) error {
	counts := make(map[string]int)
	collectEnumSchemas(schema, counts)
	definitions, _ := schema["$defs"].(map[string]any)
	if definitions == nil {
		definitions = make(map[string]any)
	}
	hoistEnumSchemas(schema, definitions, counts)
	if len(definitions) > 0 {
		schema["$defs"] = definitions
	}
	return nil
}

func collectEnumSchemas(node any, counts map[string]int) {
	switch value := node.(type) {
	case map[string]any:
		if _, ok := value["enum"]; ok {
			counts[enumSchemaKey(value)]++
		}
		// JSON object iteration order is irrelevant here: only exact, direct
		// enum schemas are candidates, so no ordering decision is needed.
		for _, field := range value {
			collectEnumSchemas(field, counts)
		}
	case []any:
		for _, field := range value {
			collectEnumSchemas(field, counts)
		}
	}
}

func hoistEnumSchemas(node any, definitions map[string]any, counts map[string]int) {
	switch value := node.(type) {
	case map[string]any:
		for key, field := range value {
			object, ok := field.(map[string]any)
			if !ok {
				hoistEnumSchemas(field, definitions, counts)
				continue
			}
			_, hasEnum := object["enum"]
			id := enumSchemaKey(object)
			if !hasEnum || counts[id] < 2 {
				hoistEnumSchemas(object, definitions, counts)
				continue
			}
			name := "Enum_" + id
			if _, exists := definitions[name]; !exists {
				definition := make(map[string]any, len(object))
				for fieldKey, fieldValue := range object {
					if fieldKey != "description" {
						definition[fieldKey] = fieldValue
					}
				}
				definitions[name] = definition
			}
			ref := map[string]any{"$ref": "#/$defs/" + name}
			for _, metadata := range []string{"description", "title", "default", "examples"} {
				if metadataValue, exists := object[metadata]; exists {
					ref[metadata] = metadataValue
				}
			}
			value[key] = ref
		}
	case []any:
		for _, field := range value {
			hoistEnumSchemas(field, definitions, counts)
		}
	}
}

func enumSchemaKey(schema map[string]any) string {
	key := canonicalEnumSchema(schema)
	raw, err := json.Marshal(key)
	if err != nil {
		panic(fmt.Sprintf("marshal enum schema: %v", err))
	}
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:8])
}

func canonicalEnumSchema(schema map[string]any) map[string]any {
	key := make(map[string]any, len(schema))
	for field, value := range schema {
		switch field {
		case "description", "title", "default", "examples":
			continue
		default:
			key[field] = value
		}
	}
	if values, ok := key["enum"].([]any); ok {
		texts := make([]string, 0, len(values))
		valid := true
		for _, value := range values {
			text, ok := value.(string)
			if !ok {
				valid = false
				break
			}
			texts = append(texts, text)
		}
		if valid {
			slices.Sort(texts)
			sorted := make([]any, len(texts))
			for index, text := range texts {
				sorted[index] = text
			}
			key["enum"] = sorted
		}
	}
	return key
}
