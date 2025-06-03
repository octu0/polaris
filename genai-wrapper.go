package polaris

import (
	"google.golang.org/genai"
)

//
// `genai.Type` cannot be JSON Marshal (Schema.Type)
// For this reason, mutual conversion here
//

type WrapFunctionDeclaration struct {
	Name        string      `json:"name,omitempty"`
	Description string      `json:"description,omitempty"`
	Parameters  *WrapSchema `json:"parameters,omitempty"`
	Response    *WrapSchema `json:"response,omitempty"`
}

func (w WrapFunctionDeclaration) ToGenAI() genai.FunctionDeclaration {
	return genai.FunctionDeclaration{
		Name:        w.Name,
		Description: w.Description,
		Parameters:  w.Parameters.ToGenAI(),
		Response:    w.Response.ToGenAI(),
	}
}

type WrapSchema struct {
	Type        string                 `json:"type,omitempty"`
	Description string                 `json:"description,omitempty"`
	Default     any                    `json:"default,omitempty"`
	Format      string                 `json:"format,omitempty"`
	Enum        []string               `json:"enum,omitempty"`
	Items       *WrapSchema            `json:"items,omitempty"`
	Nullable    *bool                  `json:"nullable,omitempty"`
	Properties  map[string]*WrapSchema `json:"properties,omitempty"`
	Required    []string               `json:"required,omitempty"`
}

func (w WrapSchema) ToGenAI() *genai.Schema {
	properties := make(map[string]*genai.Schema, len(w.Properties))
	for k, v := range w.Properties {
		properties[k] = v.ToGenAI()
	}
	items := (*genai.Schema)(nil)
	if w.Items != nil {
		items = w.Items.ToGenAI()
	}
	return &genai.Schema{
		Type:        ToGenAIType(w.Type),
		Description: w.Description,
		Default:     w.Default,
		Format:      w.Format,
		Enum:        w.Enum,
		Items:       items,
		Nullable:    w.Nullable,
		Properties:  properties,
		Required:    w.Required,
	}
}

func ToGenAIType(t string) genai.Type {
	switch genai.Type(t) {
	case genai.TypeString:
		return genai.TypeString
	case genai.TypeNumber:
		return genai.TypeNumber
	case genai.TypeInteger:
		return genai.TypeInteger
	case genai.TypeBoolean:
		return genai.TypeBoolean
	case genai.TypeArray:
		return genai.TypeArray
	case genai.TypeObject:
		return genai.TypeObject
	case genai.TypeNULL:
		return genai.TypeNULL
	default:
		return genai.TypeUnspecified
	}
}
