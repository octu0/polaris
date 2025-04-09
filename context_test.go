package polaris

import (
	"testing"
)

func TestCtxObjectArray(t *testing.T) {
	tests := []struct {
		name        string
		paramSchema Object
		req         JSONMap
		key         string
		wantLen     int
		checkValues func(t *testing.T, ctxs []*Ctx)
	}{
		{
			name: "Get ObjectArray parameter",
			paramSchema: Object{
				Properties: Properties{
					"users": ObjectArray{
						Description:     "User list",
						ItemDescription: "User",
						Required:        true,
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
					},
				},
			},
			req: JSONMap{
				"users": []any{
					map[string]any{
						"name": "Tanaka",
						"age":  30,
					},
					map[string]any{
						"name": "Sato",
						"age":  25,
					},
				},
			},
			key:     "users",
			wantLen: 2,
			checkValues: func(t *testing.T, ctxs []*Ctx) {
				if len(ctxs) != 2 {
					t.Fatalf("Expected 2 contexts, got %d", len(ctxs))
				}

				// First context
				if name := ctxs[0].String("name"); name != "Tanaka" {
					t.Errorf("First context name = %s, want Tanaka", name)
				}
				if age := ctxs[0].Int("age"); age != 30 {
					t.Errorf("First context age = %d, want 30", age)
				}

				// Second context
				if name := ctxs[1].String("name"); name != "Sato" {
					t.Errorf("Second context name = %s, want Sato", name)
				}
				if age := ctxs[1].Int("age"); age != 25 {
					t.Errorf("Second context age = %d, want 25", age)
				}
			},
		},
		{
			name: "Parameter does not exist",
			paramSchema: Object{
				Properties: Properties{
					"users": ObjectArray{
						Description:     "User list",
						ItemDescription: "User",
						Required:        true,
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
					},
				},
			},
			req:     JSONMap{},
			key:     "users",
			wantLen: 0,
			checkValues: func(t *testing.T, ctxs []*Ctx) {
				if len(ctxs) != 0 {
					t.Errorf("Expected empty context array, got %d items", len(ctxs))
				}
			},
		},
		{
			name: "Parameter is of different type",
			paramSchema: Object{
				Properties: Properties{
					"users": ObjectArray{
						Description:     "User list",
						ItemDescription: "User",
						Required:        true,
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
					},
				},
			},
			req: JSONMap{
				"users": "invalid value",
			},
			key:     "users",
			wantLen: 0,
			checkValues: func(t *testing.T, ctxs []*Ctx) {
				if len(ctxs) != 0 {
					t.Errorf("Expected empty context array, got %d items", len(ctxs))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := JSONMap{}
			ctx := &Ctx{
				req:         tt.req,
				paramSchema: tt.paramSchema,
				resp:        &resp,
			}

			got := ctx.ObjectArray(tt.key)
			if len(got) != tt.wantLen {
				t.Errorf("Ctx.ObjectArray() length = %d, want %d", len(got), tt.wantLen)
			}

			if tt.checkValues != nil {
				tt.checkValues(t, got)
			}
		})
	}
}

func TestJSONMapObjectArray(t *testing.T) {
	tests := []struct {
		name         string
		jsonMap      JSONMap
		key          string
		defaultValue []JSONMap
		wantLen      int
		checkValues  func(t *testing.T, maps []JSONMap)
	}{
		{
			name: "Get object array",
			jsonMap: JSONMap{
				"items": []any{
					map[string]any{
						"id":   1,
						"name": "Item 1",
					},
					map[string]any{
						"id":   2,
						"name": "Item 2",
					},
				},
			},
			key:          "items",
			defaultValue: []JSONMap{},
			wantLen:      2,
			checkValues: func(t *testing.T, maps []JSONMap) {
				if len(maps) != 2 {
					t.Fatalf("Expected 2 maps, got %d", len(maps))
				}

				// First map
				if id, ok := maps[0]["id"].(int); !ok || id != 1 {
					t.Errorf("First map id = %v, want 1", maps[0]["id"])
				}
				if name, ok := maps[0]["name"].(string); !ok || name != "Item 1" {
					t.Errorf("First map name = %v, want Item 1", maps[0]["name"])
				}

				// Second map
				if id, ok := maps[1]["id"].(int); !ok || id != 2 {
					t.Errorf("Second map id = %v, want 2", maps[1]["id"])
				}
				if name, ok := maps[1]["name"].(string); !ok || name != "Item 2" {
					t.Errorf("Second map name = %v, want Item 2", maps[1]["name"])
				}
			},
		},
		{
			name:         "Key does not exist",
			jsonMap:      JSONMap{},
			key:          "items",
			defaultValue: []JSONMap{{}, {}},
			wantLen:      2,
			checkValues: func(t *testing.T, maps []JSONMap) {
				if len(maps) != 2 {
					t.Errorf("Expected 2 maps, got %d", len(maps))
				}
			},
		},
		{
			name: "Value is not an array",
			jsonMap: JSONMap{
				"items": "invalid value",
			},
			key:          "items",
			defaultValue: []JSONMap{{}, {}},
			wantLen:      2,
			checkValues: func(t *testing.T, maps []JSONMap) {
				if len(maps) != 2 {
					t.Errorf("Expected 2 maps, got %d", len(maps))
				}
			},
		},
		{
			name: "Array elements are not objects",
			jsonMap: JSONMap{
				"items": []any{
					"string 1",
					"string 2",
				},
			},
			key:          "items",
			defaultValue: []JSONMap{},
			wantLen:      2,
			checkValues: func(t *testing.T, maps []JSONMap) {
				if len(maps) != 2 {
					t.Errorf("Expected 2 maps, got %d", len(maps))
				}
				// Both maps should be empty
				if len(maps[0]) != 0 {
					t.Errorf("First map should be empty, got %v", maps[0])
				}
				if len(maps[1]) != 0 {
					t.Errorf("Second map should be empty, got %v", maps[1])
				}
			},
		},
		{
			name: "Empty array",
			jsonMap: JSONMap{
				"items": []any{},
			},
			key:          "items",
			defaultValue: []JSONMap{{}, {}},
			wantLen:      0,
			checkValues: func(t *testing.T, maps []JSONMap) {
				if len(maps) != 0 {
					t.Errorf("Expected empty array, got %d items", len(maps))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.jsonMap.ObjectArray(tt.key, tt.defaultValue)
			if len(got) != tt.wantLen {
				t.Errorf("JSONMap.ObjectArray() length = %d, want %d", len(got), tt.wantLen)
			}

			if tt.checkValues != nil {
				tt.checkValues(t, got)
			}
		})
	}
}
