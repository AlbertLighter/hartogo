package converter

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"go/format"
	"io/ioutil"
	"net/url"
	"os"
	"strconv"
	"strings"
	"text/template"
)

//go:embed templates/resty.tmpl
var restyTemplate string

// TemplateData holds all the necessary data for generating a single request function.
type TemplateData struct {
	Request
	Response
	FunctionName       string
	RequestStructName  string
	RequestStructDef   string
	ResponseStructName string
	ResponseStructDef  string
}

// UTF8BOM defines the byte order mark for UTF-8
var UTF8BOM = []byte{0xEF, 0xBB, 0xBF}

// escapeString safely escapes a string for use inside a Go string literal.
func escapeString(s string) string {
	quoted := strconv.Quote(s)
	return quoted[1 : len(quoted)-1]
}

// ReadHARFromFile reads and parses a HAR file from the given path.
func ReadHARFromFile(filePath string) (*HAR, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	content, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	if bytes.HasPrefix(content, UTF8BOM) {
		content = content[len(UTF8BOM):]
	}

	var har HAR
	if err := json.Unmarshal(content, &har); err != nil {
		return nil, err
	}

	return &har, nil
}

// GenerateCode generates Go code from a TemplateData object.
// GenerateCode generates Go code from a TemplateData object.
func GenerateCode(data TemplateData) (string, error) {
	parsedURL, err := url.Parse(data.Request.URL)
	if err != nil {
		return "", fmt.Errorf("failed to parse request URL: %w", err)
	}
	baseURL := parsedURL.Scheme + "://" + parsedURL.Host + parsedURL.Path

	finalData := data
	finalData.Request.URL = baseURL

	// Generate request struct if applicable
	if strings.Contains(data.Request.PostData.MimeType, "application/json") && data.Request.PostData.Text != "" {
		structName := data.FunctionName + "Request"
		var structDef string

		// The text from HAR postData can be a string that itself contains escaped JSON.
		// We try to unescape it first.
		var unescapedText string
		if json.Unmarshal([]byte(`"`+data.Request.PostData.Text+`"`), &unescapedText) == nil {
			// If unescaping succeeds, try generating the struct from the unescaped text.
			var err error
			structDef, err = GenerateStruct(unescapedText, structName)
			if err == nil {
				// Success! Update PostData.Text to the unescaped version for the template.
				data.Request.PostData.Text = unescapedText
				finalData.RequestStructName = structName
				finalData.RequestStructDef = structDef
			}
		}

		// If struct generation hasn't succeeded yet, try with the original text.
		if finalData.RequestStructName == "" {
			structDef, err := GenerateStruct(data.Request.PostData.Text, structName)
			if err == nil {
				finalData.RequestStructName = structName
				finalData.RequestStructDef = structDef
			} else {
				fmt.Printf("Could not generate request struct for %s: %v\n", data.FunctionName, err)
			}
		}
	}

	// Generate response struct if applicable
	if strings.Contains(data.Response.Content.MimeType, "application/json") && data.Response.Content.Text != "" {
		structName := data.FunctionName + "Response"
		structDef, err := GenerateStruct(data.Response.Content.Text, structName)
		if err == nil {
			finalData.ResponseStructName = structName
			finalData.ResponseStructDef = structDef
		} else {
			fmt.Printf("Could not generate response struct for %s: %v\n", data.FunctionName, err)
		}
	}

	// Sanitize all string fields for template injection
	finalData.Request.PostData.Text = escapeString(data.Request.PostData.Text)
	finalData.Request.Headers = make([]Header, len(data.Request.Headers))
	for i, h := range data.Request.Headers {
		finalData.Request.Headers[i] = Header{Name: escapeString(h.Name), Value: escapeString(h.Value)}
	}
	finalData.Request.QueryString = make([]QueryString, len(data.Request.QueryString))
	for i, q := range data.Request.QueryString {
		finalData.Request.QueryString[i] = QueryString{Name: escapeString(q.Name), Value: escapeString(q.Value)}
	}
	finalData.Request.PostData.Params = make([]Param, len(data.Request.PostData.Params))
	for i, p := range data.Request.PostData.Params {
		finalData.Request.PostData.Params[i] = Param{Name: escapeString(p.Name), Value: escapeString(p.Value)}
	}

	tmpl, err := template.New("resty").Parse(restyTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, finalData); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	formattedCode, err := format.Source(buf.Bytes())
	if err != nil {
		// If formatting fails, print the unformatted code for debugging.
		fmt.Println("--BEGIN UNFORMATTED CODE--")
		fmt.Println(buf.String())
		fmt.Println("--END UNFORMATTED CODE--")
		return "", fmt.Errorf("failed to format generated code: %w", err)
	}

	return string(formattedCode), nil
}

