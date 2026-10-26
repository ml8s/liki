package mcpcontract

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/google/jsonschema-go/jsonschema"
)

var errFastInputFallback = errors.New("fast input validator fallback")

func decodeInputForSchema[In any](schemas derivedSchemas, raw json.RawMessage) (In, error) {
	var input In
	if len(raw) == 0 || string(raw) == "null" {
		raw = json.RawMessage(`{}`)
	}
	if schemas.fastInput {
		if err := validateFastInput(schemas.inputTree, raw, schemas.fastClosed); err != nil {
			if errors.Is(err, errFastInputFallback) {
				return decodeInput[In](schemas.inputResolved, raw)
			}
			return input, err
		}
		if err := json.Unmarshal(raw, &input); err != nil {
			return input, err
		}
		return input, nil
	}
	return decodeInput[In](schemas.inputResolved, raw)
}

// supportsFastInputSchema reports whether the generated schema uses only the
// closed keyword subset handled by validateFastInput. Complex composition,
// references, regexes, and uniqueness checks stay on the generic validator.
func supportsFastInputSchema(schema *jsonschema.Schema) bool {
	if schema == nil {
		return true
	}
	if schema.Ref != "" || schema.DynamicRef != "" ||
		schema.Const != nil || schema.MultipleOf != nil || schema.Pattern != "" ||
		schema.ExclusiveMinimum != nil || schema.ExclusiveMaximum != nil ||
		len(schema.PrefixItems) > 0 || len(schema.ItemsArray) > 0 ||
		schema.AdditionalItems != nil || schema.UniqueItems ||
		schema.Contains != nil || schema.MinContains != nil || schema.MaxContains != nil ||
		schema.UnevaluatedItems != nil ||
		schema.MinProperties != nil || schema.MaxProperties != nil ||
		len(schema.PatternProperties) > 0 || schema.PropertyNames != nil ||
		schema.UnevaluatedProperties != nil ||
		len(schema.AllOf) > 0 || len(schema.AnyOf) > 0 || len(schema.OneOf) > 0 ||
		schema.Not != nil || schema.If != nil || schema.Then != nil || schema.Else != nil ||
		len(schema.DependentRequired) > 0 || len(schema.DependentSchemas) > 0 ||
		schema.ContentEncoding != "" || schema.ContentMediaType != "" || schema.ContentSchema != nil ||
		len(schema.Extra) > 0 {
		return false
	}
	if schema.AdditionalProperties != nil && !isFalseSchema(schema.AdditionalProperties) {
		return false
	}
	for _, property := range schema.Properties {
		if !supportsFastInputSchema(property) {
			return false
		}
	}
	return schema.Items == nil || supportsFastInputSchema(schema.Items)
}

func isFalseSchema(schema *jsonschema.Schema) bool {
	return schema != nil && schema.Not != nil && reflect.ValueOf(*schema.Not).IsZero()
}

func collectClosedInputSchemas(schema *jsonschema.Schema, closed map[*jsonschema.Schema]bool) {
	if schema == nil {
		return
	}
	if isFalseSchema(schema.AdditionalProperties) {
		closed[schema] = true
	}
	for _, property := range schema.Properties {
		collectClosedInputSchemas(property, closed)
	}
	collectClosedInputSchemas(schema.Items, closed)
}

func validateFastInput(schema *jsonschema.Schema, raw json.RawMessage, closed map[*jsonschema.Schema]bool) error {
	// The scanner validates the generated schema subset without materializing
	// an any-tree. Malformed JSON that this permissive scanner accepts is
	// still rejected by the typed json.Unmarshal below; JSON constructs with
	// ambiguous duplicate names return errFastInputFallback to the canonical
	// generic validator.
	scanner := fastJSONScanner{data: raw, closed: closed}
	if err := scanner.value(schema, "$"); err != nil {
		return err
	}
	if err := scanner.whitespace(); err != nil {
		return err
	}
	if scanner.pos != len(scanner.data) {
		return fmt.Errorf("input contains more than one JSON value")
	}
	return nil
}

type fastJSONScanner struct {
	data   []byte
	pos    int
	closed map[*jsonschema.Schema]bool
}

func (scanner *fastJSONScanner) value(schema *jsonschema.Schema, path string) error {
	if err := scanner.whitespace(); err != nil {
		return err
	}
	if scanner.pos >= len(scanner.data) {
		return fmt.Errorf("%s: unexpected end of JSON", path)
	}
	if schema == nil {
		return scanner.skipValue(path)
	}
	switch scanner.data[scanner.pos] {
	case '{':
		if !schemaAllowsType(schema, "object") {
			return fmt.Errorf("%s: type: value has type %q, want %q", path, "object", strings.Join(allowedTypes(schema), ", "))
		}
		return scanner.object(schema, path)
	case '[':
		if !schemaAllowsType(schema, "array") {
			return fmt.Errorf("%s: type: value has type %q, want %q", path, "array", strings.Join(allowedTypes(schema), ", "))
		}
		return scanner.array(schema, path)
	default:
		value, err := scanner.scalar()
		if err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
		return validateFastScalar(value, schema, path)
	}
}

func (scanner *fastJSONScanner) isClosed(schema *jsonschema.Schema) bool {
	if scanner.closed != nil {
		return scanner.closed[schema]
	}
	return isFalseSchema(schema.AdditionalProperties)
}

func (scanner *fastJSONScanner) skipValue(path string) error {
	if err := scanner.whitespace(); err != nil {
		return err
	}
	if scanner.pos >= len(scanner.data) {
		return fmt.Errorf("%s: unexpected end of JSON", path)
	}
	switch scanner.data[scanner.pos] {
	case '{':
		scanner.pos++
		if err := scanner.whitespace(); err != nil {
			return err
		}
		if scanner.pos < len(scanner.data) && scanner.data[scanner.pos] == '}' {
			scanner.pos++
			return nil
		}
		for {
			if _, err := scanner.string(); err != nil {
				return fmt.Errorf("%s: object key: %w", path, err)
			}
			if err := scanner.colon(); err != nil {
				return fmt.Errorf("%s: %w", path, err)
			}
			if err := scanner.skipValue(path); err != nil {
				return err
			}
			if err := scanner.whitespace(); err != nil {
				return err
			}
			if scanner.pos >= len(scanner.data) {
				return fmt.Errorf("%s: unexpected end of object", path)
			}
			if scanner.data[scanner.pos] == ',' {
				scanner.pos++
				continue
			}
			if scanner.data[scanner.pos] == '}' {
				scanner.pos++
				return nil
			}
			return fmt.Errorf("%s: malformed object", path)
		}
	case '[':
		scanner.pos++
		if err := scanner.whitespace(); err != nil {
			return err
		}
		if scanner.pos < len(scanner.data) && scanner.data[scanner.pos] == ']' {
			scanner.pos++
			return nil
		}
		for {
			if err := scanner.skipValue(path); err != nil {
				return err
			}
			if err := scanner.whitespace(); err != nil {
				return err
			}
			if scanner.pos >= len(scanner.data) {
				return fmt.Errorf("%s: unexpected end of array", path)
			}
			if scanner.data[scanner.pos] == ',' {
				scanner.pos++
				continue
			}
			if scanner.data[scanner.pos] == ']' {
				scanner.pos++
				return nil
			}
			return fmt.Errorf("%s: malformed array", path)
		}
	default:
		_, err := scanner.scalar()
		if err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
		return nil
	}
}

func schemaAllowsType(schema *jsonschema.Schema, actual string) bool {
	allowed := allowedTypes(schema)
	if len(allowed) == 0 {
		return true
	}
	for _, want := range allowed {
		if actual == want || actual == "integer" && want == "number" {
			return true
		}
	}
	return false
}

func allowedTypes(schema *jsonschema.Schema) []string {
	if schema.Type != "" {
		return []string{schema.Type}
	}
	return schema.Types
}

func (scanner *fastJSONScanner) object(schema *jsonschema.Schema, path string) error {
	scanner.pos++ // consume {
	if err := scanner.whitespace(); err != nil {
		return err
	}
	if scanner.pos >= len(scanner.data) {
		return fmt.Errorf("%s: unexpected end of object", path)
	}
	if scanner.data[scanner.pos] == '}' {
		scanner.pos++
		return requiredPresent(schema, [32]bool{}, nil, path)
	}

	var requiredSeen [32]bool
	var requiredSeenLarge map[int]bool
	if len(schema.Required) > len(requiredSeen) {
		requiredSeenLarge = make(map[int]bool, len(schema.Required))
	}
	var seenNames [8]string
	seenNameCount := 0
	var seenNamesLarge []string
	for {
		if err := scanner.whitespace(); err != nil {
			return err
		}
		key, err := scanner.string()
		if err != nil {
			return fmt.Errorf("%s: object key: %w", path, err)
		}
		if err := scanner.colon(); err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
		for _, seen := range seenNames[:seenNameCount] {
			if seen == key {
				return errFastInputFallback
			}
		}
		for _, seen := range seenNamesLarge {
			if seen == key {
				return errFastInputFallback
			}
		}
		if seenNameCount < len(seenNames) {
			seenNames[seenNameCount] = key
			seenNameCount++
		} else {
			seenNamesLarge = append(seenNamesLarge, key)
		}
		propertySchema := schema.Properties[key]
		if err := scanner.value(propertySchema, path); err != nil {
			return fmt.Errorf("%s.%s: %w", path, key, err)
		}
		for index, required := range schema.Required {
			if required == key {
				if requiredSeenLarge != nil {
					requiredSeenLarge[index] = true
				} else if index < len(requiredSeen) {
					requiredSeen[index] = true
				}
			}
		}
		if err := scanner.whitespace(); err != nil {
			return err
		}
		if scanner.pos >= len(scanner.data) {
			return fmt.Errorf("%s: unexpected end of object", path)
		}
		switch scanner.data[scanner.pos] {
		case ',':
			scanner.pos++
			continue
		case '}':
			scanner.pos++
			if err := requiredPresent(schema, requiredSeen, requiredSeenLarge, path); err != nil {
				return err
			}
			return scanner.noUnknownProperties(schema, seenNames, seenNameCount, seenNamesLarge, path)
		default:
			return fmt.Errorf("%s: malformed object", path)
		}
	}
}

func (scanner *fastJSONScanner) noUnknownProperties(
	schema *jsonschema.Schema,
	seenNames [8]string,
	seenNameCount int,
	seenNamesLarge []string,
	path string,
) error {
	if !scanner.isClosed(schema) {
		return nil
	}
	for _, key := range seenNames[:seenNameCount] {
		if _, exists := schema.Properties[key]; !exists {
			return fmt.Errorf("%s: unexpected additional property %q", path, key)
		}
	}
	for _, key := range seenNamesLarge {
		if _, exists := schema.Properties[key]; !exists {
			return fmt.Errorf("%s: unexpected additional property %q", path, key)
		}
	}
	return nil
}

func requiredPresent(
	schema *jsonschema.Schema,
	seen [32]bool,
	seenLarge map[int]bool,
	path string,
) error {
	var missing []string
	for index, required := range schema.Required {
		present := index < len(seen) && seen[index]
		if seenLarge != nil {
			present = seenLarge[index]
		}
		if !present {
			missing = append(missing, required)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("%s: required: missing properties: %q", path, missing)
	}
	return nil
}

func (scanner *fastJSONScanner) array(schema *jsonschema.Schema, path string) error {
	scanner.pos++ // consume [
	length := 0
	if err := scanner.whitespace(); err != nil {
		return err
	}
	if scanner.pos < len(scanner.data) && scanner.data[scanner.pos] == ']' {
		scanner.pos++
	} else {
		for {
			if err := scanner.value(schema.Items, path); err != nil {
				return fmt.Errorf("%s[%d]: %w", path, length, err)
			}
			length++
			if err := scanner.whitespace(); err != nil {
				return err
			}
			if scanner.pos >= len(scanner.data) {
				return fmt.Errorf("%s: unexpected end of array", path)
			}
			if scanner.data[scanner.pos] == ',' {
				scanner.pos++
				continue
			}
			if scanner.data[scanner.pos] == ']' {
				scanner.pos++
				break
			}
			return fmt.Errorf("%s: malformed array", path)
		}
	}
	if schema.MinItems != nil && length < *schema.MinItems {
		return fmt.Errorf("%s: minItems: array has %d items, want at least %d", path, length, *schema.MinItems)
	}
	if schema.MaxItems != nil && length > *schema.MaxItems {
		return fmt.Errorf("%s: maxItems: array has %d items, want at most %d", path, length, *schema.MaxItems)
	}
	return nil
}

func (scanner *fastJSONScanner) scalar() (any, error) {
	if scanner.pos >= len(scanner.data) {
		return nil, fmt.Errorf("unexpected end of JSON")
	}
	switch scanner.data[scanner.pos] {
	case '"':
		return scanner.string()
	case 'n':
		if err := scanner.literal("null"); err != nil {
			return nil, err
		}
		return nil, nil
	case 't':
		if err := scanner.literal("true"); err != nil {
			return nil, err
		}
		return true, nil
	case 'f':
		if err := scanner.literal("false"); err != nil {
			return nil, err
		}
		return false, nil
	default:
		start := scanner.pos
		for scanner.pos < len(scanner.data) {
			char := scanner.data[scanner.pos]
			if char == '-' || char == '+' || char == '.' || char == 'e' || char == 'E' ||
				(char >= '0' && char <= '9') {
				scanner.pos++
				continue
			}
			break
		}
		if scanner.pos == start {
			return nil, fmt.Errorf("invalid JSON value")
		}
		value, err := strconv.ParseFloat(string(scanner.data[start:scanner.pos]), 64)
		if err != nil {
			return nil, fmt.Errorf("invalid number %q", scanner.data[start:scanner.pos])
		}
		return value, nil
	}
}

func (scanner *fastJSONScanner) string() (string, error) {
	raw, err := scanner.stringBytes()
	if err != nil {
		return "", err
	}
	if strings.ContainsRune(string(raw), '\\') {
		var decoded string
		if err := json.Unmarshal(raw, &decoded); err != nil {
			return "", err
		}
		return decoded, nil
	}
	return string(raw[1 : len(raw)-1]), nil
}

func (scanner *fastJSONScanner) stringBytes() ([]byte, error) {
	if scanner.pos >= len(scanner.data) || scanner.data[scanner.pos] != '"' {
		return nil, fmt.Errorf("expected string")
	}
	start := scanner.pos
	scanner.pos++
	for scanner.pos < len(scanner.data) {
		char := scanner.data[scanner.pos]
		if char == '"' {
			scanner.pos++
			return scanner.data[start:scanner.pos], nil
		}
		if char == '\\' {
			scanner.pos++
			if scanner.pos >= len(scanner.data) {
				return nil, fmt.Errorf("unterminated escaped string")
			}
		} else if char < 0x20 {
			return nil, fmt.Errorf("control character in string")
		}
		scanner.pos++
	}
	return nil, fmt.Errorf("unterminated string")
}

func (scanner *fastJSONScanner) literal(value string) error {
	if len(scanner.data)-scanner.pos < len(value) || string(scanner.data[scanner.pos:scanner.pos+len(value)]) != value {
		return fmt.Errorf("invalid JSON literal")
	}
	scanner.pos += len(value)
	return nil
}

func (scanner *fastJSONScanner) colon() error {
	if err := scanner.whitespace(); err != nil {
		return err
	}
	if scanner.pos >= len(scanner.data) || scanner.data[scanner.pos] != ':' {
		return fmt.Errorf("expected ':'")
	}
	scanner.pos++
	return nil
}

func (scanner *fastJSONScanner) whitespace() error {
	for scanner.pos < len(scanner.data) {
		switch scanner.data[scanner.pos] {
		case ' ', '\t', '\r', '\n':
			scanner.pos++
		default:
			return nil
		}
	}
	return nil
}

func validateFastScalar(value any, schema *jsonschema.Schema, path string) error {
	if schema == nil {
		return nil
	}
	if err := validateFastType(value, schema, path); err != nil {
		return err
	}
	if len(schema.Enum) > 0 && !enumContains(value, schema.Enum) {
		return fmt.Errorf("%s: enum: %v does not equal any allowed value", path, value)
	}
	if text, ok := value.(string); ok {
		length := utf8.RuneCountInString(text)
		if schema.MinLength != nil && length < *schema.MinLength {
			return fmt.Errorf("%s: minLength: string has %d runes, want at least %d", path, length, *schema.MinLength)
		}
		if schema.MaxLength != nil && length > *schema.MaxLength {
			return fmt.Errorf("%s: maxLength: string has %d runes, want at most %d", path, length, *schema.MaxLength)
		}
	}
	if number, ok := value.(float64); ok {
		if schema.Minimum != nil && number < *schema.Minimum {
			return fmt.Errorf("%s: minimum: %v is less than %v", path, number, *schema.Minimum)
		}
		if schema.Maximum != nil && number > *schema.Maximum {
			return fmt.Errorf("%s: maximum: %v is greater than %v", path, number, *schema.Maximum)
		}
	}
	return nil
}

func validateFastType(value any, schema *jsonschema.Schema, path string) error {
	actual := scalarJSONType(value)
	allowed := schema.Types
	if schema.Type != "" {
		allowed = []string{schema.Type}
	}
	if len(allowed) == 0 {
		return nil
	}
	for _, want := range allowed {
		if actual == want || actual == "integer" && want == "number" {
			return nil
		}
	}
	return fmt.Errorf("%s: type: %v has type %q, want %q", path, value, actual, strings.Join(allowed, ", "))
}

func scalarJSONType(value any) string {
	switch typed := value.(type) {
	case nil:
		return "null"
	case bool:
		return "boolean"
	case string:
		return "string"
	case float64:
		if math.Trunc(typed) == typed {
			return "integer"
		}
		return "number"
	case []any:
		return "array"
	case map[string]any:
		return "object"
	default:
		return fmt.Sprintf("%T", value)
	}
}

func enumContains(value any, values []any) bool {
	for _, allowed := range values {
		switch typed := value.(type) {
		case string:
			text, ok := allowed.(string)
			if ok && typed == text {
				return true
			}
		case float64:
			number, ok := allowed.(float64)
			if ok && typed == number {
				return true
			}
		case bool:
			boolean, ok := allowed.(bool)
			if ok && typed == boolean {
				return true
			}
		case nil:
			if allowed == nil {
				return true
			}
		}
	}
	return false
}
