package converter

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// ToCamelCase converts a snake_case or kebab-case string to CamelCase.
func ToCamelCase(s string) string {
	re := regexp.MustCompile(`[._-]+`)
	s = re.ReplaceAllString(s, " ")
	titleCaser := cases.Title(language.English)
	s = titleCaser.String(s)
	return strings.ReplaceAll(s, " ", "")
}

// GenerateStruct converts a JSON string into a Go struct definition with inlined nested structs.
func GenerateStruct(jsonString, topLevelStructName string) (string, error) {
	var data interface{}
	if err := json.Unmarshal([]byte(jsonString), &data); err != nil {
		return "", fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("type %s ", topLevelStructName))

	generated, err := generateType(data, 0)
	if err != nil {
		return "", err
	}
	sb.WriteString(generated)
	sb.WriteString("\n")

	return sb.String(), nil
}

func generateType(data interface{}, indentLevel int) (string, error) {
	switch v := data.(type) {
	case map[string]interface{}:
		return generateStruct(v, indentLevel)
	case []interface{}:
		if len(v) == 0 {
			return "[]interface{}", nil
		}
		// Infer type from the first element of the slice
		firstElType, err := generateType(v[0], indentLevel)
		if err != nil {
			return "", err
		}
		return "[]" + firstElType, nil
	case string:
		return "string", nil
	case float64:
		if v == float64(int(v)) {
			return "int", nil
		}
		return "float64", nil
	case bool:
		return "bool", nil
	case nil:
		return "interface{}", nil
	default:
		return "", fmt.Errorf("unsupported type: %T", v)
	}
}

func generateStruct(data map[string]interface{}, indentLevel int) (string, error) {
	var sb strings.Builder
	sb.WriteString("struct {")

	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	indent := strings.Repeat("\t", indentLevel+1)

	for _, key := range keys {
		value := data[key]
		fieldName := ToCamelCase(key)
		goType, err := generateType(value, indentLevel+1)
		if err != nil {
			return "", err
		}

		sb.WriteString(fmt.Sprintf("%s%s %s `json:\"%s,omitempty\"`\n", indent, fieldName, goType, key))
	}

	sb.WriteString(strings.Repeat("\t", indentLevel) + "}")
	return sb.String(), nil
}