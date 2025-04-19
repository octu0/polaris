package polaris

import (
	"github.com/mark3labs/mcp-go/mcp"
)

func primitiveToTypeDef(schema map[string]any, isRequired bool) TypeDef {
	propertyDesc, ok := schema["description"].(string)
	if ok != true {
		propertyDesc = ""
	}

	switch schema["type"] {
	case "number":
		f := Float{
			Description: propertyDesc,
			Required:    isRequired,
		}
		if defaultValue, hasDefault := schema["default"].(float64); hasDefault {
			f.Default = defaultValue
		}
		return f
	case "string":
		s := String{
			Description: propertyDesc,
			Required:    isRequired,
		}
		if defaultValue, hasDefault := schema["default"].(string); hasDefault {
			s.Default = defaultValue
		}
		return s
	case "boolean":
		b := Bool{
			Description: propertyDesc,
			Required:    isRequired,
		}
		if defaultValue, hasDefault := schema["default"].(bool); hasDefault {
			b.Default = defaultValue
		}
		return b
	}
	return nil
}

func arrayToTypedef(schema map[string]any, isRequired bool) TypeDef {
	propertyDesc, ok := schema["description"].(string)
	if ok != true {
		propertyDesc = ""
	}

	items := schema["items"].(map[string]any)
	switch items["type"] {
	case "number":
		return FloatArray{
			Description: propertyDesc,
			Required:    isRequired,
		}
	case "string":
		return StringArray{
			Description: propertyDesc,
			Required:    isRequired,
		}
	case "boolean":
		return BoolArray{
			Description: propertyDesc,
			Required:    isRequired,
		}
	case "object":
		itemProperties := items["properties"].(map[string]any)
		return ObjectArray{
			Description: propertyDesc,
			Items:       objectToProperties(itemProperties, isRequired),
			Required:    isRequired,
		}
	}
	return nil
}

func objectToProperties(schema map[string]any, isRequired bool) Properties {
	propertyDesc, ok := schema["description"].(string)
	if ok != true {
		propertyDesc = ""
	}
	requires, ok := schema["required"].([]string)
	if ok != true {
		requires = []string{}
	}
	requiredMap := make(map[string]struct{})
	for _, key := range requires {
		requiredMap[key] = struct{}{}
	}

	objectProp := Properties{}
	for itemPropKey, itemPropSchema := range schema {
		_, subRequired := requiredMap[itemPropKey]

		subSchema, subSchemaOK := itemPropSchema.(map[string]any)
		if subSchemaOK != true {
			continue
		}
		switch subSchema["type"] {
		case "number", "string", "boolean":
			objectProp[itemPropKey] = primitiveToTypeDef(subSchema, subRequired)
		case "array":
			objectProp[itemPropKey] = arrayToTypedef(subSchema, subRequired)
		case "object":
			objectProp[itemPropKey] = Object{
				Description: propertyDesc,
				Properties:  objectToProperties(subSchema, subRequired),
				Required:    isRequired,
			}
		}
	}
	return objectProp
}

func convertInputSchema(schema mcp.ToolInputSchema) Object {
	prop := Properties{}
	requiredMap := make(map[string]struct{}, len(schema.Required))
	for _, key := range schema.Required {
		requiredMap[key] = struct{}{}
	}

	for key, properySchema := range schema.Properties {
		schema, ok := properySchema.(map[string]any)
		if ok != true {
			continue
		}

		_, isRequired := requiredMap[key]

		switch schema["type"] {
		case "number", "string", "boolean":
			prop[key] = primitiveToTypeDef(schema, isRequired)
		case "array":
			prop[key] = arrayToTypedef(schema, isRequired)
		case "object":
			propertyDesc, ok := schema["description"].(string)
			if ok != true {
				propertyDesc = ""
			}
			objProperties := schema["properties"].(map[string]any)
			prop[key] = Object{
				Description: propertyDesc,
				Properties:  objectToProperties(objProperties, isRequired),
				Required:    isRequired,
			}
		}
	}
	return Object{
		Properties: prop,
		Required:   true,
	}
}
