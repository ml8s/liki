package mcpcontract

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
)

// canonicalJSON encodes JSON with sorted object keys and no insignificant
// whitespace. Array order is preserved. Numbers remain their decoded JSON
// scalar form.
// marshalCompactJSON emits one compact JSON document without the HTML
// escaping that Go enables by default. Canonical output must match Python's
// UTF-8 JSON encoder byte for byte.
func marshalCompactJSON(value any) ([]byte, error) {
	var raw bytes.Buffer
	encoder := json.NewEncoder(&raw)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(value); err != nil {
		return nil, err
	}
	return bytes.TrimRight(raw.Bytes(), "\n"), nil
}

func canonicalJSON(value any) ([]byte, error) {
	raw, err := marshalCompactJSON(value)
	if err != nil {
		return nil, err
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	var decoded any
	if err := decoder.Decode(&decoded); err != nil {
		return nil, err
	}

	var out bytes.Buffer
	if err := encodeCanonicalJSON(&out, decoded); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

func encodeCanonicalJSON(out *bytes.Buffer, value any) error {
	switch item := value.(type) {
	case nil:
		out.WriteString("null")
	case bool:
		if item {
			out.WriteString("true")
		} else {
			out.WriteString("false")
		}
	case string:
		encoded, err := marshalCompactJSON(item)
		if err != nil {
			return err
		}
		out.Write(encoded)
	case json.Number:
		out.WriteString(item.String())
	case []any:
		out.WriteByte('[')
		for index, child := range item {
			if index > 0 {
				out.WriteByte(',')
			}
			if err := encodeCanonicalJSON(out, child); err != nil {
				return err
			}
		}
		out.WriteByte(']')
	case map[string]any:
		keys := make([]string, 0, len(item))
		for key := range item {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		out.WriteByte('{')
		for index, key := range keys {
			if index > 0 {
				out.WriteByte(',')
			}
			encodedKey, err := marshalCompactJSON(key)
			if err != nil {
				return err
			}
			out.Write(encodedKey)
			out.WriteByte(':')
			if err := encodeCanonicalJSON(out, item[key]); err != nil {
				return err
			}
		}
		out.WriteByte('}')
	default:
		return fmt.Errorf("unsupported canonical JSON value %T", value)
	}
	return nil
}
