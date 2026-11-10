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
	hoistRepeatedObjects(schema)
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

// hoistRepeatedObjects shares exact object schemas at root $defs. Only
// property/item positions are candidates: tool roots and generated enum
// definitions stay untouched. A size floor prevents many tiny $ref wrappers
// from costing more wire bytes than they save.
func hoistRepeatedObjects(schema map[string]any) {
	counts := make(map[string]int)
	for key, value := range schema {
		if key == "$defs" {
			continue
		}
		collectObjectSchemas(value, counts)
	}
	definitions, _ := schema["$defs"].(map[string]any)
	if definitions == nil {
		definitions = make(map[string]any)
	}
	for key, value := range schema {
		if key == "$defs" {
			continue
		}
		hoistObjectSchemas(value, definitions, counts)
	}
	if len(definitions) > 0 {
		schema["$defs"] = definitions
	}
}

func collectObjectSchemas(node any, counts map[string]int) {
	switch value := node.(type) {
	case map[string]any:
		if kind, ok := value["type"].(string); ok && kind == "object" {
			if objectSchemaBytes(value) >= minObjectSchemaBytes {
				counts[objectSchemaKey(value)]++
			}
			// A repeated object is published as one unit. Do not also count
			// its descendants for replacement inside this unit; that keeps
			// generated definitions self-contained and easier for clients to
			// resolve conservatively.
			return
		}
		for _, field := range value {
			collectObjectSchemas(field, counts)
		}
	case []any:
		for _, field := range value {
			collectObjectSchemas(field, counts)
		}
	}
}

func hoistObjectSchemas(node any, definitions map[string]any, counts map[string]int) {
	switch value := node.(type) {
	case map[string]any:
		for key, field := range value {
			object, ok := field.(map[string]any)
			if !ok {
				hoistObjectSchemas(field, definitions, counts)
				continue
			}
			kind, ok := object["type"].(string)
			if !ok || kind != "object" {
				hoistObjectSchemas(field, definitions, counts)
				continue
			}
			id := objectSchemaKey(object)
			if counts[id] < 2 || objectSchemaBytes(object) < minObjectSchemaBytes {
				hoistObjectSchemas(field, definitions, counts)
				continue
			}
			name := "Object_" + id
			if _, exists := definitions[name]; !exists {
				definition := make(map[string]any, len(object))
				for objectKey, objectValue := range object {
					switch objectKey {
					case "description", "title", "default", "examples":
						continue
					default:
						definition[objectKey] = objectValue
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
		for key, field := range value {
			if object, ok := field.(map[string]any); ok {
				if kind, exists := object["type"].(string); exists && kind == "object" {
					id := objectSchemaKey(object)
					if counts[id] >= 2 && objectSchemaBytes(object) >= minObjectSchemaBytes {
						name := "Object_" + id
						if _, exists := definitions[name]; !exists {
							definition := make(map[string]any, len(object))
							for objectKey, objectValue := range object {
								switch objectKey {
								case "description", "title", "default", "examples":
									continue
								default:
									definition[objectKey] = objectValue
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
						continue
					}
				}
			}
			hoistObjectSchemas(field, definitions, counts)
		}
	}
}

func objectSchemaKey(schema map[string]any) string {
	key := make(map[string]any, len(schema))
	for field, value := range schema {
		switch field {
		case "description", "title", "default", "examples":
			continue
		default:
			key[field] = value
		}
	}
	raw, err := json.Marshal(key)
	if err != nil {
		panic(fmt.Sprintf("marshal object schema: %v", err))
	}
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:8])
}

const minObjectSchemaBytes = 120

func objectSchemaBytes(schema map[string]any) int {
	key := canonicalObjectSchema(schema)
	raw, err := json.Marshal(key)
	if err != nil {
		panic(fmt.Sprintf("marshal object schema: %v", err))
	}
	return len(raw)
}

func canonicalObjectSchema(schema map[string]any) map[string]any {
	key := make(map[string]any, len(schema))
	for field, value := range schema {
		switch field {
		case "description", "title", "default", "examples":
			continue
		default:
			key[field] = value
		}
	}
	return key
}
