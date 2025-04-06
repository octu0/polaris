package polaris

import (
	"cloud.google.com/go/vertexai/genai"
)

type JSONMap map[string]any
type Resp = JSONMap

func (m JSONMap) Int(key string, defaultValue int) int {
	if v, ok := m[key].(int); ok {
		return v
	}
	if v, ok := m[key].(float64); ok {
		return int(v)
	}
	return defaultValue
}

func (m JSONMap) String(key string, defaultValue string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return defaultValue
}

func (m JSONMap) Bool(key string, defaultValue bool) bool {
	if v, ok := m[key].(bool); ok {
		return v
	}
	return defaultValue
}

func (m JSONMap) Float64(key string, defaultValue float64) float64 {
	if v, ok := m[key].(float64); ok {
		return v
	}
	if v, ok := m[key].(int); ok {
		return float64(v)
	}
	return defaultValue
}

func (m JSONMap) IntArray(key string, defaultValue []int) []int {
	if v, ok := m[key].([]any); ok {
		ret := make([]int, len(v))
		for i, vv := range v {
			if intVal, ok := vv.(int); ok {
				ret[i] = intVal
			}
		}
		return ret
	}
	return defaultValue
}

func (m JSONMap) Float64Array(key string, defaultValue []float64) []float64 {
	if v, ok := m[key].([]any); ok {
		ret := make([]float64, len(v))
		for i, vv := range v {
			if intVal, ok := vv.(float64); ok {
				ret[i] = intVal
			}
		}
		return ret
	}
	return defaultValue
}

func (m JSONMap) ObjectArray(key string, defaultValue []JSONMap) []JSONMap {
	if v, ok := m[key].([]any); ok {
		ret := make([]JSONMap, len(v))
		for i, o := range v {
			r := JSONMap{}
			if kv, ok := o.(map[string]any); ok {
				for k, a := range kv {
					r.Set(k, a)
				}
			}
			ret[i] = r
		}
		return ret
	}
	return defaultValue
}

func (m JSONMap) Object(key string, defaultValue JSONMap) JSONMap {
	if v, ok := m[key].(map[string]any); ok {
		ret := make(JSONMap, len(v))
		for k, a := range v {
			ret.Set(k, a)
		}
		return ret
	}
	return defaultValue
}

func (m JSONMap) Set(key string, value any) {
	switch v := value.(type) {
	case []map[string]any:
		objArray := make([]any, len(v))
		for i, m := range v {
			objArray[i] = m
		}
		m[key] = objArray
	case []int:
		intArray := make([]any, len(v))
		for i, vv := range v {
			intArray[i] = vv
		}
		m[key] = intArray
	case []float64:
		floatArray := make([]any, len(v))
		for i, vv := range v {
			floatArray[i] = vv
		}
		m[key] = floatArray
	default:
		m[key] = v
	}
}

func (m JSONMap) ToMap() map[string]any {
	ret := make(map[string]any, len(m))
	for key, value := range m {
		switch v := value.(type) {
		case JSONMap:
			nest := make(map[string]any, len(v))
			for kk, vv := range v {
				nest[kk] = vv
			}
			ret[key] = nest
		default:
			ret[key] = value
		}
	}
	return ret
}

type Ctx struct {
	req         JSONMap
	paramSchema Object
	resp        *JSONMap
}

func (c *Ctx) Int(key string) int {
	t := c.paramSchema.Properties[key]
	if tt, ok := t.(Int); ok {
		return c.req.Int(key, tt.Default)
	}
	if _, ok := t.(IntEnum); ok {
		return c.req.Int(key, 0)
	}
	return c.req.Int(key, 0)
}

func (c *Ctx) Float64(key string) float64 {
	t := c.paramSchema.Properties[key]
	if tt, ok := t.(Float); ok {
		return c.req.Float64(key, tt.Default)
	}
	return c.req.Float64(key, 0.0)
}

func (c *Ctx) String(key string) string {
	t := c.paramSchema.Properties[key]
	if tt, ok := t.(String); ok {
		return c.req.String(key, tt.Default)
	}
	if _, ok := t.(StringEnum); ok {
		return c.req.String(key, "")
	}
	return c.req.String(key, "")
}

func (c *Ctx) Bool(key string) bool {
	t := c.paramSchema.Properties[key]
	if tt, ok := t.(Bool); ok {
		return c.req.Bool(key, tt.Default)
	}
	return c.req.Bool(key, false)
}

func (c *Ctx) IntArray(key string) []int {
	t := c.paramSchema.Properties[key]
	if _, ok := t.(IntArray); ok {
		return c.req.IntArray(key, []int{})
	}
	if arr, ok := t.(Array); ok {
		if arr.Items.Schema().Type == genai.TypeInteger {
			return c.req.IntArray(key, []int{})
		}
	}
	return c.req.IntArray(key, []int{})
}

func (c *Ctx) FloatArray(key string) []float64 {
	t := c.paramSchema.Properties[key]
	if _, ok := t.(FloatArray); ok {
		return c.req.Float64Array(key, []float64{})
	}
	if arr, ok := t.(Array); ok {
		if arr.Items.Schema().Type == genai.TypeNumber {
			return c.req.Float64Array(key, []float64{})
		}
	}
	return c.req.Float64Array(key, []float64{})
}

func (c *Ctx) Object(key string) *Ctx {
	t := c.paramSchema.Properties[key]
	if obj, ok := t.(Object); ok {
		r := JSONMap{}
		c.resp.Set(key, r)
		return &Ctx{c.req.Object(key, JSONMap{}), obj, &r}
	}
	return nil
}

func (c *Ctx) Set(r Resp) {
	for k, v := range r {
		c.resp.Set(k, v)
	}
}
