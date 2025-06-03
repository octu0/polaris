package polaris

/*
import (
	"reflect"
	"slices"
	"testing"

	"google.golang.org/genai"
)

func TestNullableType(t *testing.T) {
	tests := []struct {
		name     string
		nullable NullableType
		want     bool
	}{
		{
			name:     "NullableYes",
			nullable: NullableYes,
			want:     true,
		},
		{
			name:     "NullableNo",
			nullable: NullableNo,
			want:     false,
		},
		{
			name:     "Empty string defaults to true",
			nullable: "",
			want:     true,
		},
		{
			name:     "Unknown value defaults to true",
			nullable: "unknown",
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.nullable.Nullable(); *got != tt.want {
				t.Errorf("NullableType.Nullable() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTool(t *testing.T) {
	t.Run("FunctionDeclaration", func(tt *testing.T) {
		params := Object{
			Description: "Test parameters",
			Properties: Properties{
				"name": String{
					Description: "User name",
					Required:    true,
				},
				"age": Int{
					Description: "User age",
					Required:    false,
				},
			},
		}

		response := Object{
			Description: "Test response",
			Properties: Properties{
				"success": Bool{
					Description: "Operation success",
					Required:    true,
				},
				"message": String{
					Description: "Response message",
					Required:    true,
				},
			},
		}

		tool := Tool{
			Name:        "test_tool",
			Description: "A test tool",
			Parameters:  params,
			Response:    response,
			Handler:     func(ctx *ReqCtx) (Resp, error) { return nil, nil },
		}

		want := genai.FunctionDeclaration{
			Name:        "test_tool",
			Description: "A test tool",
			Parameters:  params.Schema(),
			Response:    response.Schema(),
		}

		got := tool.FunctionDeclaration()
		if got.Name != want.Name {
			tt.Errorf("Tool.FunctionDeclaration().Name = %v, want %v", got.Name, want.Name)
		}
		if got.Description != want.Description {
			tt.Errorf("Tool.FunctionDeclaration().Description = %v, want %v", got.Description, want.Description)
		}

		if got.Parameters.Type != want.Parameters.Type {
			tt.Errorf("Tool.FunctionDeclaration().Parameters.Type = %v, want %v", got.Parameters.Type, want.Parameters.Type)
		}
		if got.Parameters.Description != want.Parameters.Description {
			tt.Errorf("Tool.FunctionDeclaration().Parameters.Description = %v, want %v", got.Parameters.Description, want.Parameters.Description)
		}

		if got.Response.Type != want.Response.Type {
			tt.Errorf("Tool.FunctionDeclaration().Response.Type = %v, want %v", got.Response.Type, want.Response.Type)
		}
		if got.Response.Description != want.Response.Description {
			tt.Errorf("Tool.FunctionDeclaration().Response.Description = %v, want %v", got.Response.Description, want.Response.Description)
		}
	})
}

func TestObject(t *testing.T) {
	t.Run("Schema", func(tt *testing.T) {
		obj := Object{
			Description: "Test object",
			Properties: Properties{
				"name": String{
					Description: "User name",
					Required:    true,
				},
				"age": Int{
					Description: "User age",
					Required:    false,
				},
				"active": Bool{
					Description: "Is user active",
					Required:    true,
				},
			},
			Required: true,
			Nullable: NullableNo,
		}

		schema := obj.Schema()
		if ToGenAIType(schema.Type) != genai.TypeObject {
			tt.Errorf("Object.Schema().Type = %v, want %v", schema.Type, genai.TypeObject)
		}
		if schema.Description != "Test object" {
			tt.Errorf("Object.Schema().Description = %v, want %v", schema.Description, "Test object")
		}
		if *schema.Nullable != false {
			tt.Errorf("Object.Schema().Nullable = %v, want %v", schema.Nullable, false)
		}

		if len(schema.Properties) != 3 {
			tt.Errorf("len(Object.Schema().Properties) = %v, want %v", len(schema.Properties), 3)
		}

		expectedRequired := []string{"name", "active"}
		requiredCopy := slices.Clone(schema.Required)
		expectedCopy := slices.Clone(expectedRequired)
		slices.Sort(requiredCopy)
		slices.Sort(expectedCopy)
		if reflect.DeepEqual(requiredCopy, expectedCopy) != true {
			tt.Errorf("Object.Schema().Required = %v, want %v", schema.Required, expectedRequired)
		}

		if _, ok := schema.Properties["name"]; !ok {
			tt.Errorf(`Object.Schema().Properties["name"] not found`)
		}
		if _, ok := schema.Properties["age"]; !ok {
			tt.Errorf(`Object.Schema().Properties["age"] not found`)
		}
		if _, ok := schema.Properties["active"]; !ok {
			tt.Errorf(`Object.Schema().Properties["active"] not found`)
		}
		if ToGenAIType(schema.Properties["name"].Type) != genai.TypeString {
			tt.Errorf(`Object.Schema().Properties["name"].Type = %v, want %v`, schema.Properties["name"].Type, genai.TypeString)
		}
		if ToGenAIType(schema.Properties["age"].Type) != genai.TypeInteger {
			tt.Errorf(`Object.Schema().Properties["age"].Type = %v, want %v`, schema.Properties["age"].Type, genai.TypeInteger)
		}
		if ToGenAIType(schema.Properties["active"].Type) != genai.TypeBoolean {
			tt.Errorf(`Object.Schema().Properties["active"].Type = %v, want %v`, schema.Properties["active"].Type, genai.TypeBoolean)
		}
	})
}

func TestArray(t *testing.T) {
	t.Run("Schema", func(tt *testing.T) {
		stringType := String{
			Description: "A string item",
			Required:    true,
		}

		array := Array{
			Description: "Test array",
			Items:       stringType,
			Required:    true,
			Nullable:    NullableNo,
		}

		schema := array.Schema()
		if schema.Type != genai.TypeArray {
			tt.Errorf("Array.Schema().Type = %v, want %v", schema.Type, genai.TypeArray)
		}
		if schema.Description != "Test array" {
			tt.Errorf("Array.Schema().Description = %v, want %v", schema.Description, "Test array")
		}
		if *schema.Nullable != false {
			tt.Errorf("Array.Schema().Nullable = %v, want %v", schema.Nullable, false)
		}
		if schema.Items.Type != genai.TypeString {
			tt.Errorf("Array.Schema().Items.Type = %v, want %v", schema.Items.Type, genai.TypeString)
		}
		if schema.Items.Description != "A string item" {
			tt.Errorf("Array.Schema().Items.Description = %v, want %v", schema.Items.Description, "A string item")
		}
	})
}

func TestIntArray(t *testing.T) {
	t.Run("Schema", func(tt *testing.T) {
		array := IntArray{
			Description:     "Test int array",
			ItemDescription: "An integer item",
			Required:        true,
			Nullable:        NullableNo,
		}

		schema := array.Schema()
		if schema.Type != genai.TypeArray {
			tt.Errorf("IntArray.Schema().Type = %v, want %v", schema.Type, genai.TypeArray)
		}
		if schema.Description != "Test int array" {
			tt.Errorf("IntArray.Schema().Description = %v, want %v", schema.Description, "Test int array")
		}
		if *schema.Nullable != false {
			tt.Errorf("IntArray.Schema().Nullable = %v, want %v", schema.Nullable, false)
		}
		if schema.Items.Type != genai.TypeInteger {
			tt.Errorf("IntArray.Schema().Items.Type = %v, want %v", schema.Items.Type, genai.TypeInteger)
		}
		if schema.Items.Description != "An integer item" {
			tt.Errorf("IntArray.Schema().Items.Description = %v, want %v", schema.Items.Description, "An integer item")
		}
	})
}

func TestFloatArray(t *testing.T) {
	t.Run("Schema", func(tt *testing.T) {
		array := FloatArray{
			Description:     "Test float array",
			ItemDescription: "A float item",
			Required:        true,
			Nullable:        NullableNo,
		}

		schema := array.Schema()
		if schema.Type != genai.TypeArray {
			tt.Errorf("FloatArray.Schema().Type = %v, want %v", schema.Type, genai.TypeArray)
		}
		if schema.Description != "Test float array" {
			tt.Errorf("FloatArray.Schema().Description = %v, want %v", schema.Description, "Test float array")
		}
		if *schema.Nullable != false {
			tt.Errorf("FloatArray.Schema().Nullable = %v, want %v", schema.Nullable, false)
		}
		if schema.Items.Type != genai.TypeNumber {
			tt.Errorf("FloatArray.Schema().Items.Type = %v, want %v", schema.Items.Type, genai.TypeNumber)
		}
		if schema.Items.Description != "A float item" {
			tt.Errorf("FloatArray.Schema().Items.Description = %v, want %v", schema.Items.Description, "A float item")
		}
	})
}

func TestStringArray(t *testing.T) {
	t.Run("Schema", func(tt *testing.T) {
		array := StringArray{
			Description:     "Test string array",
			ItemDescription: "A string item",
			Required:        true,
			Nullable:        NullableNo,
		}

		schema := array.Schema()
		if schema.Type != genai.TypeArray {
			tt.Errorf("StringArray.Schema().Type = %v, want %v", schema.Type, genai.TypeArray)
		}
		if schema.Description != "Test string array" {
			tt.Errorf("StringArray.Schema().Description = %v, want %v", schema.Description, "Test string array")
		}
		if *schema.Nullable != false {
			tt.Errorf("StringArray.Schema().Nullable = %v, want %v", schema.Nullable, false)
		}
		if schema.Items.Type != genai.TypeString {
			tt.Errorf("StringArray.Schema().Items.Type = %v, want %v", schema.Items.Type, genai.TypeString)
		}
		if schema.Items.Description != "A string item" {
			tt.Errorf("StringArray.Schema().Items.Description = %v, want %v", schema.Items.Description, "A string item")
		}
	})
}

func TestBoolArray(t *testing.T) {
	t.Run("Schema", func(tt *testing.T) {
		array := BoolArray{
			Description:     "Test bool array",
			ItemDescription: "A boolean item",
			Required:        true,
			Nullable:        NullableNo,
		}

		schema := array.Schema()
		if schema.Type != genai.TypeArray {
			tt.Errorf("BoolArray.Schema().Type = %v, want %v", schema.Type, genai.TypeArray)
		}
		if schema.Description != "Test bool array" {
			tt.Errorf("BoolArray.Schema().Description = %v, want %v", schema.Description, "Test bool array")
		}
		if *schema.Nullable != false {
			tt.Errorf("BoolArray.Schema().Nullable = %v, want %v", schema.Nullable, false)
		}
		if schema.Items.Type != genai.TypeBoolean {
			tt.Errorf("BoolArray.Schema().Items.Type = %v, want %v", schema.Items.Type, genai.TypeBoolean)
		}
		if schema.Items.Description != "A boolean item" {
			tt.Errorf("BoolArray.Schema().Items.Description = %v, want %v", schema.Items.Description, "A boolean item")
		}
	})
}

func TestObjectArray(t *testing.T) {
	t.Run("Schema", func(tt *testing.T) {
		array := ObjectArray{
			Description:     "Test object array",
			ItemDescription: "An object item",
			Required:        true,
			Nullable:        NullableNo,
			Items: Properties{
				"name": String{
					Description: "User name",
					Required:    true,
				},
				"age": Int{
					Description: "User age",
					Required:    false,
				},
			},
		}

		schema := array.Schema()
		if schema.Type != genai.TypeArray {
			tt.Errorf("ObjectArray.Schema().Type = %v, want %v", schema.Type, genai.TypeArray)
		}
		if schema.Description != "Test object array" {
			tt.Errorf("ObjectArray.Schema().Description = %v, want %v", schema.Description, "Test object array")
		}
		if *schema.Nullable != false {
			tt.Errorf("ObjectArray.Schema().Nullable = %v, want %v", schema.Nullable, false)
		}
		if schema.Items.Type != genai.TypeObject {
			tt.Errorf("ObjectArray.Schema().Items.Type = %v, want %v", schema.Items.Type, genai.TypeObject)
		}
		if schema.Items.Description != "An object item" {
			tt.Errorf("ObjectArray.Schema().Items.Description = %v, want %v", schema.Items.Description, "An object item")
		}

		if len(schema.Items.Properties) != 2 {
			tt.Errorf("len(ObjectArray.Schema().Items.Properties) = %v, want %v", len(schema.Items.Properties), 2)
		}

		expectedRequired := []string{"name"}
		requiredCopy := slices.Clone(schema.Items.Required)
		expectedCopy := slices.Clone(expectedRequired)
		slices.Sort(requiredCopy)
		slices.Sort(expectedCopy)
		if reflect.DeepEqual(requiredCopy, expectedCopy) != true {
			tt.Errorf("ObjectArray.Schema().Items.Required = %v, want %v", schema.Items.Required, expectedRequired)
		}

		if _, ok := schema.Items.Properties["name"]; !ok {
			tt.Errorf(`ObjectArray.Schema().Items.Properties["name"] not found`)
		}
		if _, ok := schema.Items.Properties["age"]; !ok {
			tt.Errorf(`ObjectArray.Schema().Items.Properties["age"] not found`)
		}
	})
}

func TestIntEnum(t *testing.T) {
	t.Run("Schema", func(tt *testing.T) {
		enum := IntEnum{
			Description: "Test int enum",
			Values:      []string{"100", "200", "300"},
			Required:    true,
			Nullable:    NullableNo,
		}

		schema := enum.Schema()
		if schema.Type != genai.TypeInteger {
			tt.Errorf("IntEnum.Schema().Type = %v, want %v", schema.Type, genai.TypeInteger)
		}
		if schema.Description != "Test int enum" {
			tt.Errorf("IntEnum.Schema().Description = %v, want %v", schema.Description, "Test int enum")
		}
		if *schema.Nullable != false {
			tt.Errorf("IntEnum.Schema().Nullable = %v, want %v", schema.Nullable, false)
		}
		if schema.Format != "enum" {
			tt.Errorf("IntEnum.Schema().Format = %v, want %v", schema.Format, "enum")
		}

		expectedValues := []string{"100", "200", "300"}
		if reflect.DeepEqual(schema.Enum, expectedValues) != true {
			tt.Errorf("IntEnum.Schema().Enum = %v, want %v", schema.Enum, expectedValues)
		}
	})
}

func TestStringEnum(t *testing.T) {
	t.Run("Schema", func(tt *testing.T) {
		enum := StringEnum{
			Description: "Test string enum",
			Values:      []string{"north", "east", "south", "west"},
			Required:    true,
			Nullable:    NullableNo,
		}

		schema := enum.Schema()
		if schema.Type != genai.TypeString {
			tt.Errorf("StringEnum.Schema().Type = %v, want %v", schema.Type, genai.TypeString)
		}
		if schema.Description != "Test string enum" {
			tt.Errorf("StringEnum.Schema().Description = %v, want %v", schema.Description, "Test string enum")
		}
		if *schema.Nullable != false {
			tt.Errorf("StringEnum.Schema().Nullable = %v, want %v", schema.Nullable, false)
		}
		if schema.Format != "enum" {
			tt.Errorf("StringEnum.Schema().Format = %v, want %v", schema.Format, "enum")
		}

		expectedValues := []string{"north", "east", "south", "west"}
		if reflect.DeepEqual(schema.Enum, expectedValues) != true {
			tt.Errorf("StringEnum.Schema().Enum = %v, want %v", schema.Enum, expectedValues)
		}
	})
}

func TestInt(t *testing.T) {
	t.Run("Schema", func(tt *testing.T) {
		intType := Int{
			Description: "Test int",
			Default:     42,
			Required:    true,
			Nullable:    NullableNo,
		}

		schema := intType.Schema()
		if schema.Type != genai.TypeInteger {
			tt.Errorf("Int.Schema().Type = %v, want %v", schema.Type, genai.TypeInteger)
		}
		if schema.Description != "Test int" {
			tt.Errorf("Int.Schema().Description = %v, want %v", schema.Description, "Test int")
		}
		if *schema.Nullable != false {
			tt.Errorf("Int.Schema().Nullable = %v, want %v", schema.Nullable, false)
		}
	})
}

func TestFloat(t *testing.T) {
	t.Run("Schema", func(tt *testing.T) {
		floatType := Float{
			Description: "Test float",
			Default:     3.14,
			Required:    true,
			Nullable:    NullableNo,
		}

		schema := floatType.Schema()
		if schema.Type != genai.TypeNumber {
			tt.Errorf("Float.Schema().Type = %v, want %v", schema.Type, genai.TypeNumber)
		}
		if schema.Description != "Test float" {
			tt.Errorf("Float.Schema().Description = %v, want %v", schema.Description, "Test float")
		}
		if *schema.Nullable != false {
			tt.Errorf("Float.Schema().Nullable = %v, want %v", schema.Nullable, false)
		}
	})
}

func TestString(t *testing.T) {
	t.Run("Schema", func(tt *testing.T) {
		stringType := String{
			Description: "Test string",
			Default:     "default value",
			Required:    true,
			Nullable:    NullableNo,
		}

		schema := stringType.Schema()
		if schema.Type != genai.TypeString {
			tt.Errorf("String.Schema().Type = %v, want %v", schema.Type, genai.TypeString)
		}
		if schema.Description != "Test string" {
			tt.Errorf("String.Schema().Description = %v, want %v", schema.Description, "Test string")
		}
		if *schema.Nullable != false {
			tt.Errorf("String.Schema().Nullable = %v, want %v", schema.Nullable, false)
		}
	})
}

func TestBool(t *testing.T) {
	t.Run("Schema", func(tt *testing.T) {
		boolType := Bool{
			Description: "Test bool",
			Default:     true,
			Required:    true,
			Nullable:    NullableNo,
		}

		schema := boolType.Schema()
		if schema.Type != genai.TypeBoolean {
			tt.Errorf("Bool.Schema().Type = %v, want %v", schema.Type, genai.TypeBoolean)
		}
		if schema.Description != "Test bool" {
			tt.Errorf("Bool.Schema().Description = %v, want %v", schema.Description, "Test bool")
		}
		if *schema.Nullable != false {
			tt.Errorf("Bool.Schema().Nullable = %v, want %v", schema.Nullable, false)
		}
	})
}

func TestIsRequired(t *testing.T) {
	tests := []struct {
		name     string
		typeDef  TypeDef
		required bool
	}{
		{
			name: "Object required",
			typeDef: Object{
				Description: "Test object",
				Required:    true,
			},
			required: true,
		},
		{
			name: "Object not required",
			typeDef: Object{
				Description: "Test object",
				Required:    false,
			},
			required: false,
		},
		{
			name: "Array required",
			typeDef: Array{
				Description: "Test array",
				Required:    true,
			},
			required: true,
		},
		{
			name: "IntArray required",
			typeDef: IntArray{
				Description: "Test int array",
				Required:    true,
			},
			required: true,
		},
		{
			name: "FloatArray not required",
			typeDef: FloatArray{
				Description: "Test float array",
				Required:    false,
			},
			required: false,
		},
		{
			name: "StringArray required",
			typeDef: StringArray{
				Description: "Test string array",
				Required:    true,
			},
			required: true,
		},
		{
			name: "BoolArray not required",
			typeDef: BoolArray{
				Description: "Test bool array",
				Required:    false,
			},
			required: false,
		},
		{
			name: "ObjectArray required",
			typeDef: ObjectArray{
				Description: "Test object array",
				Required:    true,
			},
			required: true,
		},
		{
			name: "IntEnum not required",
			typeDef: IntEnum{
				Description: "Test int enum",
				Required:    false,
			},
			required: false,
		},
		{
			name: "StringEnum required",
			typeDef: StringEnum{
				Description: "Test string enum",
				Required:    true,
			},
			required: true,
		},
		{
			name: "Int not required",
			typeDef: Int{
				Description: "Test int",
				Required:    false,
			},
			required: false,
		},
		{
			name: "Float required",
			typeDef: Float{
				Description: "Test float",
				Required:    true,
			},
			required: true,
		},
		{
			name: "String not required",
			typeDef: String{
				Description: "Test string",
				Required:    false,
			},
			required: false,
		},
		{
			name: "Bool required",
			typeDef: Bool{
				Description: "Test bool",
				Required:    true,
			},
			required: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.typeDef.IsRequired(); got != tt.required {
				t.Errorf("%T.IsRequired() = %v, want %v", tt.typeDef, got, tt.required)
			}
		})
	}
}
*/
