package web

// OpenAPISpec represents the root OpenAPI specification object
type OpenAPISpec struct {
	OpenAPI    string     `json:"openapi"`
	Info       Info       `json:"info"`
	Servers    []Server   `json:"servers"`
	Paths      Paths      `json:"paths"`
	Components Components `json:"components"`
}

// Info represents the OpenAPI info object
type Info struct {
	Title       string `json:"title"`
	Version     string `json:"version"`
	Description string `json:"description"`
}

// Server represents an OpenAPI server object
type Server struct {
	URL         string `json:"url"`
	Description string `json:"description"`
}

// Paths represents the OpenAPI paths object
type Paths map[string]PathItem

// PathItem represents an OpenAPI path item object
type PathItem struct {
	Get    Operation `json:"get,omitempty"`
	Post   Operation `json:"post,omitempty"`
	Put    Operation `json:"put,omitempty"`
	Delete Operation `json:"delete,omitempty"`
}

// Operation represents an OpenAPI operation object
type Operation struct {
	Summary     string       `json:"summary"`
	Description string       `json:"description"`
	Tags        []string     `json:"tags"`
	Parameters  []Parameter  `json:"parameters,omitempty"`
	RequestBody *RequestBody `json:"requestBody,omitempty"`
	Responses   Responses    `json:"responses"`
}

// Parameter represents an OpenAPI parameter object
type Parameter struct {
	Name        string `json:"name"`
	In          string `json:"in"`
	Description string `json:"description,omitempty"`
	Required    bool   `json:"required,omitempty"`
	Schema      Schema `json:"schema"`
}

// RequestBody represents an OpenAPI request body object
type RequestBody struct {
	Description string               `json:"description,omitempty"`
	Required    bool                 `json:"required,omitempty"`
	Content     map[string]MediaType `json:"content"`
}

// Responses represents an OpenAPI responses object
type Responses map[string]Response

// Response represents an OpenAPI response object
type Response struct {
	Description string               `json:"description"`
	Content     map[string]MediaType `json:"content,omitempty"`
}

// MediaType represents an OpenAPI media type object
type MediaType struct {
	Schema Schema `json:"schema"`
}

// Components represents the OpenAPI components object
type Components struct {
	Schemas map[string]Schema `json:"schemas"`
}

// Schema represents an OpenAPI schema object
type Schema struct {
	Type                 string            `json:"type,omitempty"`
	Format               string            `json:"format,omitempty"`
	Items                *Schema           `json:"items,omitempty"`
	Properties           map[string]Schema `json:"properties,omitempty"`
	AdditionalProperties *Schema           `json:"additionalProperties,omitempty"`
	Required             []string          `json:"required,omitempty"`
	Default              interface{}       `json:"default,omitempty"`
	Example              interface{}       `json:"example,omitempty"`
	Ref                  string            `json:"$ref,omitempty"`
}
