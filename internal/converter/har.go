package converter

// HAR is the root object of a HAR file.
type HAR struct {
	Log Log `json:"log"`
}

// Log contains the list of entries.
type Log struct {
	Entries []Entry `json:"entries"`
}

// Entry represents a single request-response pair.
type Entry struct {
	Request  Request  `json:"request"`
	Response Response `json:"response"`
}

// Request contains details about the HTTP request.
type Request struct {
	Method      string        `json:"method"`
	URL         string        `json:"url"`
	Headers     []Header      `json:"headers"`
	QueryString []QueryString `json:"queryString"`
	PostData    PostData      `json:"postData"`
}

// Response contains details about the HTTP response.
type Response struct {
	Status  int     `json:"status"`
	Content Content `json:"content"`
}

// Content holds the body of the request or response.
type Content struct {
	Text     string `json:"text"`
	MimeType string `json:"mimeType"`
}

// Header represents a single HTTP header.
type Header struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// QueryString represents a single URL query parameter.
type QueryString struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// PostData represents the request body.
type PostData struct {
	MimeType string  `json:"mimeType"`
	Text     string  `json:"text"`
	Params   []Param `json:"params"`
}

// Param represents a single parameter in a form post.
type Param struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}