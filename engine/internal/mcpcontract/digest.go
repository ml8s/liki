package mcpcontract

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
)

// canonicalJSON encodes JSON with sorted object keys and no insignificant
// whitespace. Array order is preserved. Numbers remain their decoded JSON
// scalar form.
func canonicalJSON(value any) ([]byte, error) {
	raw, err := json.Marshal(value)
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
		encoded, err := json.Marshal(item)
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
			encodedKey, err := json.Marshal(key)
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

// digestExcluding hashes an object-valued artifact after removing metadata
// fields such as the digest itself.
func digestExcluding(value any, excluded ...string) (string, error) {
	raw, err := json.Marshal(value)
	if err != nil {
		return "", fmt.Errorf("marshal artifact: %w", err)
	}
	var item map[string]any
	if err := json.Unmarshal(raw, &item); err != nil {
		return "", fmt.Errorf("decode artifact object: %w", err)
	}
	for _, field := range excluded {
		delete(item, field)
	}
	canonical, err := canonicalJSON(item)
	if err != nil {
		return "", fmt.Errorf("canonical json: %w", err)
	}
	sum := sha256.Sum256(canonical)
	return hex.EncodeToString(sum[:]), nil
}

// chartDigest computes the canonical digest of a chart while excluding its own
// chart_digest metadata field.
func chartDigest(chart any) (string, error) {
	return digestExcluding(chart, "chart_digest")
}

// sourceChartDigest verifies the digest carried by a reusable source chart and
// returns its canonical value.
func sourceChartDigest(chart any) (string, error) {
	raw, err := json.Marshal(chart)
	if err != nil {
		return "", fmt.Errorf("marshal source chart: %w", err)
	}
	var item map[string]any
	if err := json.Unmarshal(raw, &item); err != nil {
		return "", fmt.Errorf("decode source chart: %w", err)
	}
	provided, _ := item["chart_digest"].(string)
	expected, err := chartDigest(item)
	if err != nil {
		return "", err
	}
	if provided == "" {
		return "", fmt.Errorf("source chart has no chart_digest")
	}
	if provided != expected {
		return "", fmt.Errorf("chart_digest mismatch: got %s, want %s", provided, expected)
	}
	return provided, nil
}

// inputSourceDigest extracts and verifies the chart digest from an input that
// carries a reusable Chart field. A nil optional chart has no source digest.
func inputSourceDigest(input any) (*ChartDigest, error) {
	value := reflect.ValueOf(input)
	if value.Kind() == reflect.Pointer {
		value = value.Elem()
	}
	if value.Kind() != reflect.Struct {
		return nil, fmt.Errorf("input is not a struct")
	}
	field := value.FieldByName("Chart")
	if !field.IsValid() {
		return nil, fmt.Errorf("input has no chart")
	}
	if field.Kind() == reflect.Pointer && field.IsNil() {
		return nil, nil
	}
	digest, err := sourceChartDigest(field.Interface())
	if err != nil {
		return nil, err
	}
	chartDigest := ChartDigest(digest)
	return &chartDigest, nil
}

// withSourceChartDigest attaches the verified source-chart identity to a raw
// downstream artifact without changing its domain projection.
func withSourceChartDigest(raw json.RawMessage, digest string) (json.RawMessage, error) {
	var item map[string]any
	if err := json.Unmarshal(raw, &item); err != nil {
		return nil, fmt.Errorf("decode artifact: %w", err)
	}
	item["source_chart_digest"] = digest
	return json.Marshal(item)
}
