package converter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/format"
	"regexp"
	"sort"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// ToCamelCase converts a snake_case or kebab-case string to CamelCase.
func ToCamelCase(s string) string {
	// If the string has no separators, assume it's already in a desirable format.
	if !strings.ContainsAny(s, "._-لیت") {
		if len(s) == 0 {
			return ""
		}
		// Just ensure the first letter is capitalized.
		return strings.ToUpper(string(s[0])) + s[1:]
	}
	re := regexp.MustCompile(`[._-]+`)
	s = re.ReplaceAllString(s, " ")
	titleCaser := cases.Title(language.English)
	s = titleCaser.String(s)
	return strings.ReplaceAll(s, " ", "")
}

type generator struct {
	structs map[string]string
	imports map[string]struct{}
}

func newGenerator() *generator {
	return &generator{
		structs: make(map[string]string),
		imports: map[string]struct{}{
			`"encoding/json"`: {},
		},
	}
}

// GenerateStruct converts a JSON string into a Go struct definition.
// It handles nested JSON strings by creating named types with custom UnmarshalJSON methods.
func GenerateStruct(jsonString, topLevelStructName string) (string, []string, error) {
	var data interface{}
	// Use a decoder to handle large numbers and other JSON nuances
	dec := json.NewDecoder(strings.NewReader(jsonString))
	dec.UseNumber()
	if err := dec.Decode(&data); err != nil {
		return "", nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	g := newGenerator()

	typeName, err := g.generateType(data, topLevelStructName, 0)
	if err != nil {
		return "", nil, err
	}

	var sb strings.Builder

	// Write helper structs
	sortedStructNames := make([]string, 0, len(g.structs))
	for name := range g.structs {
		sortedStructNames = append(sortedStructNames, name)
	}
	sort.Strings(sortedStructNames)

	for _, name := range sortedStructNames {
		sb.WriteString(g.structs[name])
		sb.WriteString("\n\n")
	}

	// Write main struct
	sb.WriteString(fmt.Sprintf("type %s %s\n", topLevelStructName, typeName))

	// Format the generated code
	formatted, err := format.Source([]byte(sb.String()))
	if err != nil {
		// Return unformatted source on error, it's better than nothing
		return sb.String(), nil, fmt.Errorf("failed to format generated code: %w", err)
	}

	// Collect and return imports
	sortedImports := make([]string, 0, len(g.imports))
	for imp := range g.imports {
		sortedImports = append(sortedImports, imp)
	}
	sort.Strings(sortedImports)

	return string(formatted), sortedImports, nil
}

func (g *generator) generateType(data interface{}, nameHint string, indentLevel int) (string, error) {
	switch v := data.(type) {
	case map[string]interface{}:
		return g.generateStruct(v, nameHint, indentLevel)
	case []interface{}:
		if len(v) == 0 {
			return "[]interface{}", nil
		}
		// Infer type from the first element
		elemType, err := g.generateType(v[0], strings.TrimSuffix(nameHint, "s"), indentLevel)
		if err != nil {
			return "", err
		}
		return "[]" + elemType, nil
	case string:
		var nestedData interface{}
		dec := json.NewDecoder(bytes.NewReader([]byte(v)))
		dec.UseNumber()
		if err := dec.Decode(&nestedData); err == nil {
			if _, isMap := nestedData.(map[string]interface{}); isMap {
				return g.addCustomUnmarshalerStruct(nestedData, nameHint, indentLevel)
			}
			if _, isSlice := nestedData.([]interface{}); isSlice {
				return g.addCustomUnmarshalerStruct(nestedData, nameHint, indentLevel)
			}
		}
		return "string", nil
	case json.Number:
		if _, err := v.Int64(); err == nil {
			return "int64", nil
		}
		// g.imports[`"strconv"`] = struct{}{}
		return "float64", nil
	case float64: // This case might be hit if UseNumber() is not used
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

func (g *generator) generateStruct(data map[string]interface{}, nameHint string, indentLevel int) (string, error) {
	var sb strings.Builder
	sb.WriteString("struct {\n")

	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	indent := strings.Repeat("\t", indentLevel+1)

	for _, key := range keys {
		value := data[key]
		fieldName := ToCamelCase(key)
		newHint := nameHint + fieldName

		goType, err := g.generateType(value, newHint, indentLevel+1)
		if err != nil {
			return "", err
		}

		sb.WriteString(fmt.Sprintf("%s%s %s `json:\"%s,omitempty\"`\n", indent, fieldName, goType, key))
	}

	sb.WriteString(strings.Repeat("\t", indentLevel) + "}")
	return sb.String(), nil
}

func (g *generator) addCustomUnmarshalerStruct(data interface{}, nameHint string, indentLevel int) (string, error) {
	structName := nameHint
	if _, exists := g.structs[structName]; exists {
		structName += "Wrapper"
	}

	structBody, err := g.generateType(data, structName, 0) // indent 0 for top-level generation
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("type %s %s\n", structName, structBody))

	receiver := strings.ToLower(string(structName[0]))
	sb.WriteString(fmt.Sprintf("func (%s *%s) UnmarshalJSON(data []byte) error {\n", receiver, structName))
	sb.WriteString(fmt.Sprintf("\ttype alias %s\n", structName))
	sb.WriteString("\tvar s string\n")
	sb.WriteString("\tif err := json.Unmarshal(data, &s); err == nil {\n")
	sb.WriteString(fmt.Sprintf("\t\treturn json.Unmarshal([]byte(s), (*alias)(%s))\n", receiver))
	sb.WriteString("\t}\n")
	sb.WriteString(fmt.Sprintf("\treturn json.Unmarshal(data, &%s)\n", receiver))
	sb.WriteString("}\n\n")

	sb.WriteString(fmt.Sprintf("func (%s %s) MarshalJSON() ([]byte, error) {\n", receiver, structName))
	sb.WriteString(fmt.Sprintf("\ttype alias %s\n", structName))
	sb.WriteString(fmt.Sprintf("\tdata, err := json.Marshal(alias(%s))\n", receiver))
	sb.WriteString("\tif err != nil {\n")
	sb.WriteString("\t	return nil, err\n")
	sb.WriteString("\t}\n")
	sb.WriteString("\treturn json.Marshal(string(data))\n")
	sb.WriteString("}\n")

	g.structs[structName] = sb.String()

	return structName, nil
}
