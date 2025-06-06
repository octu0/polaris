package polaris

import "google.golang.org/genai"

type NullableType string

const (
	NullableYes NullableType = "yes"
	NullableNo  NullableType = "no"
)

func (n NullableType) Nullable() *bool {
	switch n {
	case NullableYes:
		return genai.Ptr(true)
	case NullableNo:
		return genai.Ptr(false)
	default:
		return genai.Ptr(true)
	}
}

type (
	ToolHandler  func(*ReqCtx) (Resp, error)
	ErrorHandler func(error)
)

type Tool struct {
	Name         string
	Description  string
	Parameters   Object
	Response     Object
	Handler      ToolHandler
	ErrorHandler ErrorHandler
}

func (t Tool) FunctionDeclaration() WrapFunctionDeclaration {
	return WrapFunctionDeclaration{
		Name:        t.Name,
		Description: t.Description,
		Parameters:  t.Parameters.Schema(),
		Response:    t.Response.Schema(),
	}
}

type TypeDef interface {
	Schema() *WrapSchema
	IsRequired() bool
}

var (
	_ TypeDef = Object{}
	_ TypeDef = Array{}
	_ TypeDef = IntArray{}
	_ TypeDef = FloatArray{}
	_ TypeDef = StringArray{}
	_ TypeDef = BoolArray{}
	_ TypeDef = ObjectArray{}
	_ TypeDef = IntEnum{}
	_ TypeDef = StringEnum{}
	_ TypeDef = Int{}
	_ TypeDef = Float{}
	_ TypeDef = String{}
	_ TypeDef = Bool{}
)

type Properties map[string]TypeDef

//	*genai.Schema{
//	  Type:        genai.TypeObject,
//	  Description: "...",
//	  Properties:  map[string]*genai.Schema{...},
//	}
type Object struct {
	Description string
	Properties  Properties
	Required    bool
	Nullable    NullableType
}

func (o Object) Schema() *WrapSchema {
	properties := make(map[string]*WrapSchema, len(o.Properties))
	requiredKeys := make([]string, 0, len(o.Properties))
	for k, v := range o.Properties {
		properties[k] = v.Schema()
		if v.IsRequired() {
			requiredKeys = append(requiredKeys, k)
		}
	}
	return &WrapSchema{
		Type:        string(genai.TypeObject),
		Description: o.Description,
		Properties:  properties,
		Required:    requiredKeys,
		Nullable:    o.Nullable.Nullable(),
	}
}

func (o Object) IsRequired() bool {
	return o.Required
}

//	*genai.Schema{
//	  Type:        genai.TypeArray,
//	  Description: "...",
//	  Items:       &genai.Schema{...},
//	}
type (
	Array struct {
		Description string
		Items       TypeDef
		Required    bool
		Nullable    NullableType
	}

	//	*genai.Schema{
	//	  Type:        genai.TypeArray,
	//	  Description: "...",
	//	  Items:       &genai.Schema{
	//      Type:        genai.TypeInteger,
	//      Description: "...",
	//    },
	//	}
	IntArray struct {
		Description     string
		ItemDescription string
		Required        bool
		Nullable        NullableType
	}

	//	*genai.Schema{
	//	  Type:        genai.TypeArray,
	//	  Description: "...",
	//	  Items:       &genai.Schema{
	//      Type:        genai.TypeNumber,
	//      Description: "...",
	//    },
	//	}
	FloatArray struct {
		Description     string
		ItemDescription string
		Required        bool
		Nullable        NullableType
	}

	//	*genai.Schema{
	//	  Type:        genai.TypeArray,
	//	  Description: "...",
	//	  Items:       &genai.Schema{
	//      Type:        genai.TypeString,
	//      Description: "...",
	//    },
	//	}
	StringArray struct {
		Description     string
		ItemDescription string
		Required        bool
		Nullable        NullableType
	}

	//	*genai.Schema{
	//	  Type:        genai.TypeArray,
	//	  Description: "...",
	//	  Items:       &genai.Schema{
	//      Type:        genai.TypeBoolean,
	//      Description: "...",
	//    },
	//	}
	BoolArray struct {
		Description     string
		ItemDescription string
		Required        bool
		Nullable        NullableType
	}

	//	*genai.Schema{
	//	  Type:        genai.TypeArray,
	//	  Description: "...",
	//	  Items:       &genai.Schema{
	//      Type:        genai.TypeObject,
	//      Properties:  map[string]*genai.Schema{
	//        ...
	//      },
	//    },
	//	}
	ObjectArray struct {
		Description     string
		ItemDescription string
		Required        bool
		Nullable        NullableType
		Items           Properties
	}
)

func (a Array) Schema() *WrapSchema {
	return &WrapSchema{
		Type:        string(genai.TypeArray),
		Description: a.Description,
		Items:       a.Items.Schema(),
		Nullable:    a.Nullable.Nullable(),
	}
}

func (a Array) IsRequired() bool {
	return a.Required
}

func (ia IntArray) Schema() *WrapSchema {
	return &WrapSchema{
		Type:        string(genai.TypeArray),
		Description: ia.Description,
		Items: &WrapSchema{
			Type:        string(genai.TypeInteger),
			Description: ia.ItemDescription,
		},
		Nullable: ia.Nullable.Nullable(),
	}
}

func (ia IntArray) IsRequired() bool {
	return ia.Required
}

func (fa FloatArray) Schema() *WrapSchema {
	return &WrapSchema{
		Type:        string(genai.TypeArray),
		Description: fa.Description,
		Items: &WrapSchema{
			Type:        string(genai.TypeNumber),
			Description: fa.ItemDescription,
		},
		Nullable: fa.Nullable.Nullable(),
	}
}

func (fa FloatArray) IsRequired() bool {
	return fa.Required
}

func (sa StringArray) Schema() *WrapSchema {
	return &WrapSchema{
		Type:        string(genai.TypeArray),
		Description: sa.Description,
		Items: &WrapSchema{
			Type:        string(genai.TypeString),
			Description: sa.ItemDescription,
		},
		Nullable: sa.Nullable.Nullable(),
	}
}

func (sa StringArray) IsRequired() bool {
	return sa.Required
}

func (ba BoolArray) Schema() *WrapSchema {
	return &WrapSchema{
		Type:        string(genai.TypeArray),
		Description: ba.Description,
		Items: &WrapSchema{
			Type:        string(genai.TypeBoolean),
			Description: ba.ItemDescription,
		},
		Nullable: ba.Nullable.Nullable(),
	}
}

func (ba BoolArray) IsRequired() bool {
	return ba.Required
}

func (oa ObjectArray) Schema() *WrapSchema {
	properties := make(map[string]*WrapSchema, len(oa.Items))
	requiredKeys := make([]string, 0, len(oa.Items))
	for k, v := range oa.Items {
		properties[k] = v.Schema()
		if v.IsRequired() {
			requiredKeys = append(requiredKeys, k)
		}
	}
	return &WrapSchema{
		Type:        string(genai.TypeArray),
		Description: oa.Description,
		Items: &WrapSchema{
			Type:        string(genai.TypeObject),
			Description: oa.ItemDescription,
			Properties:  properties,
			Required:    requiredKeys,
		},
		Nullable: oa.Nullable.Nullable(),
	}
}

func (oa ObjectArray) IsRequired() bool {
	return oa.Required
}

type (
	//	*genai.Schema{
	//	  Type:        genai.TypeInteger,
	//	  Description: "...",
	//    Enum:        []string{"100", "200", "300"},
	//    Format:      "enum",
	//	}
	IntEnum struct {
		Description string
		Values      []string
		Required    bool
		Nullable    NullableType
	}

	//	*genai.Schema{
	//	  Type:        genai.TypeString,
	//	  Description: "...",
	//    Enum:        []string{"north", "east", "south", "west"},
	//    Format:      "enum",
	//	}
	StringEnum struct {
		Description string
		Values      []string
		Required    bool
		Nullable    NullableType
	}
)

func (ie IntEnum) Schema() *WrapSchema {
	return &WrapSchema{
		Type:        string(genai.TypeInteger),
		Description: ie.Description,
		Enum:        ie.Values,
		Format:      "enum",
		Nullable:    ie.Nullable.Nullable(),
	}
}

func (ie IntEnum) IsRequired() bool {
	return ie.Required
}

func (se StringEnum) Schema() *WrapSchema {
	return &WrapSchema{
		Type:        string(genai.TypeString),
		Description: se.Description,
		Enum:        se.Values,
		Format:      "enum",
		Nullable:    se.Nullable.Nullable(),
	}
}

func (se StringEnum) IsRequired() bool {
	return se.Required
}

//	*genai.Schema{
//	  Type:        genai.TypeInteger,
//	  Description: "...",
//	}
type Int struct {
	Description string
	Default     int
	Required    bool
	Nullable    NullableType
}

func (i Int) Schema() *WrapSchema {
	return &WrapSchema{
		Type:        string(genai.TypeInteger),
		Description: i.Description,
		//Default:     i.Default,
		Nullable: i.Nullable.Nullable(),
	}
}

func (i Int) IsRequired() bool {
	return i.Required
}

//	*genai.Schema{
//	  Type:        genai.TypeNumber,
//	  Description: "...",
//	}
type Float struct {
	Description string
	Default     float64
	Required    bool
	Nullable    NullableType
}

func (f Float) Schema() *WrapSchema {
	return &WrapSchema{
		Type:        string(genai.TypeNumber),
		Description: f.Description,
		//Default:     f.Default,
		Nullable: f.Nullable.Nullable(),
	}
}

func (f Float) IsRequired() bool {
	return f.Required
}

//	*genai.Schema{
//	  Type:       genai.TypeString,
//	  Description "...",
//	}
type String struct {
	Description string
	Default     string
	Required    bool
	Nullable    NullableType
}

func (s String) Schema() *WrapSchema {
	return &WrapSchema{
		Type:        string(genai.TypeString),
		Description: s.Description,
		//Default:     s.Default,
		Nullable: s.Nullable.Nullable(),
	}
}

func (s String) IsRequired() bool {
	return s.Required
}

//	*genai.Schema{
//	  Type:        genai.TypeBoolean,
//	  Description: "...",
//	}
type Bool struct {
	Description string
	Default     bool
	Required    bool
	Nullable    NullableType
}

func (b Bool) Schema() *WrapSchema {
	return &WrapSchema{
		Type:        string(genai.TypeBoolean),
		Description: b.Description,
		//Default:     b.Default,
		Nullable: b.Nullable.Nullable(),
	}
}

func (b Bool) IsRequired() bool {
	return b.Required
}
