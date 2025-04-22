package polaris

import (
	"cloud.google.com/go/vertexai/genai"
	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/client/transport"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/pkg/errors"
)

func (c *Conn) RegisterSSEMCPTools(baseURL string, initReq mcp.InitializeRequest, options ...transport.ClientOption) error {
	mcpClient, err := client.NewSSEMCPClient(baseURL, options...)
	if err != nil {
		return errors.WithStack(err)
	}

	if err := mcpClient.Start(c.ctx); err != nil {
		return errors.WithStack(err)
	}

	initResp, err := mcpClient.Initialize(c.ctx, initReq)
	if err != nil {
		return errors.WithStack(err)
	}
	c.logger.Infof("sseMCPClient init=%v", initResp.ServerInfo)

	r, err := mcpClient.ListTools(c.ctx, mcp.ListToolsRequest{})
	if err != nil {
		return errors.WithStack(err)
	}

	tools := make([]Tool, 0)
	for _, t := range r.Tools {
		tools = append(tools, Tool{
			Name:        t.Name,
			Description: t.Description,
			Parameters:  convertInputSchema(t.InputSchema),
			Response:    Object{},
		})
	}

	for _, t := range tools {
		if err := subscribeReqResp(
			c,
			tooltopic(t.Name),
			JSONEncoder[map[string]any](),
			JSONEncoder[map[string]any](),
			handleMCPToolCall(c.ctx, mcpClient, t),
		); err != nil {
			return errors.WithStack(err)
		}
	}

	for _, t := range tools {
		resp, err := requestWithData(
			c,
			TopicRegisterTool,
			GobEncoder[genai.FunctionDeclaration](),
			GobEncoder[RespError](),
			t.FunctionDeclaration(),
		)
		if err != nil {
			return errors.WithStack(err)
		}
		if err := resp.Err(); err != nil {
			return errors.WithStack(err)
		}
	}

	c.tools = append(c.tools, tools...)
	c.mcpClients = append(c.mcpClients, mcpClient)
	return nil
}

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
