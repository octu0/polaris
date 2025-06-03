package polaris

import (
	"google.golang.org/genai"
)

type jsonMap map[string]any
type (
	Req  = jsonMap
	Resp = jsonMap
)

func (m jsonMap) Int(key string, defaultValue int) int {
	if v, ok := m[key].(int); ok {
		return v
	}
	if v, ok := m[key].(float64); ok {
		return int(v)
	}
	return defaultValue
}

func (m jsonMap) String(key string, defaultValue string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return defaultValue
}

func (m jsonMap) Bool(key string, defaultValue bool) bool {
	if v, ok := m[key].(bool); ok {
		return v
	}
	if s, ok := m[key].(string); ok {
		switch s {
		case "yes", "true", "1":
			return true
		case "no", "false", "0":
			return false
		}
	}
	return defaultValue
}

func (m jsonMap) Float64(key string, defaultValue float64) float64 {
	if v, ok := m[key].(float64); ok {
		return v
	}
	if v, ok := m[key].(int); ok {
		return float64(v)
	}
	return defaultValue
}

func (m jsonMap) IntArray(key string, defaultValue []int) []int {
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

func (m jsonMap) Float64Array(key string, defaultValue []float64) []float64 {
	if v, ok := m[key].([]any); ok {
		ret := make([]float64, len(v))
		for i, vv := range v {
			if f64Val, ok := vv.(float64); ok {
				ret[i] = f64Val
			}
		}
		return ret
	}
	return defaultValue
}

func (m jsonMap) StringArray(key string, defaultValue []string) []string {
	if v, ok := m[key].([]any); ok {
		ret := make([]string, len(v))
		for i, vv := range v {
			if strVal, ok := vv.(string); ok {
				ret[i] = strVal
			}
		}
		return ret
	}
	return defaultValue
}

func (m jsonMap) ObjectArray(key string, defaultValue []jsonMap) []jsonMap {
	if v, ok := m[key].([]any); ok {
		ret := make([]jsonMap, len(v))
		for i, o := range v {
			r := jsonMap{}
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

func (m jsonMap) Object(key string, defaultValue jsonMap) jsonMap {
	if v, ok := m[key].(map[string]any); ok {
		ret := make(jsonMap, len(v))
		for k, a := range v {
			ret.Set(k, a)
		}
		return ret
	}
	return defaultValue
}

func (m jsonMap) Array(key string, defaultValue []jsonMap) []jsonMap {
	if v, ok := m[key].([]any); ok {
		ret := make([]jsonMap, len(v))
		for i, vv := range v {
			if mm, isMap := vv.(map[string]any); isMap {
				ret[i] = mm
			}
		}
		return ret
	}
	return defaultValue
}

func (m jsonMap) Set(key string, value any) {
	switch v := value.(type) {
	case []map[string]any:
		// [{"key": "value"},...] is not of the form []map[string]any
		// structpb panics when explicit type used, repackaged as []any
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
	case []string:
		stringArray := make([]any, len(v))
		for i, vv := range v {
			stringArray[i] = vv
		}
		m[key] = stringArray
	default:
		m[key] = v
	}
}

func (m jsonMap) ToMap() map[string]any {
	ret := jsonMap(make(map[string]any, len(m)))
	for key, value := range m {
		ret.Set(key, value)
	}
	return map[string]any(ret)
}

type ReqCtx struct {
	req         jsonMap
	paramSchema Object
}

func (c *ReqCtx) Int(key string) int {
	t := c.paramSchema.Properties[key]
	if tt, ok := t.(Int); ok {
		return c.req.Int(key, tt.Default)
	}
	if _, ok := t.(IntEnum); ok {
		return c.req.Int(key, 0)
	}
	return c.req.Int(key, 0)
}

func (c *ReqCtx) Float64(key string) float64 {
	t := c.paramSchema.Properties[key]
	if tt, ok := t.(Float); ok {
		return c.req.Float64(key, tt.Default)
	}
	return c.req.Float64(key, 0.0)
}

func (c *ReqCtx) String(key string) string {
	t := c.paramSchema.Properties[key]
	if tt, ok := t.(String); ok {
		return c.req.String(key, tt.Default)
	}
	if _, ok := t.(StringEnum); ok {
		return c.req.String(key, "")
	}
	return c.req.String(key, "")
}

func (c *ReqCtx) Bool(key string) bool {
	t := c.paramSchema.Properties[key]
	if tt, ok := t.(Bool); ok {
		return c.req.Bool(key, tt.Default)
	}
	return c.req.Bool(key, false)
}

func (c *ReqCtx) IntArray(key string) []int {
	t := c.paramSchema.Properties[key]
	if _, ok := t.(IntArray); ok {
		return c.req.IntArray(key, []int{})
	}
	if arr, ok := t.(Array); ok {
		if arr.Items.Schema().Type == string(genai.TypeInteger) {
			return c.req.IntArray(key, []int{})
		}
	}
	return c.req.IntArray(key, []int{})
}

func (c *ReqCtx) FloatArray(key string) []float64 {
	t := c.paramSchema.Properties[key]
	if _, ok := t.(FloatArray); ok {
		return c.req.Float64Array(key, []float64{})
	}
	if arr, ok := t.(Array); ok {
		if arr.Items.Schema().Type == string(genai.TypeNumber) {
			return c.req.Float64Array(key, []float64{})
		}
	}
	return c.req.Float64Array(key, []float64{})
}

func (c *ReqCtx) StringArray(key string) []string {
	t := c.paramSchema.Properties[key]
	if _, ok := t.(StringArray); ok {
		return c.req.StringArray(key, []string{})
	}
	if arr, ok := t.(Array); ok {
		if arr.Items.Schema().Type == string(genai.TypeString) {
			return c.req.StringArray(key, []string{})
		}
	}
	return c.req.StringArray(key, []string{})
}

func (c *ReqCtx) Object(key string) *ReqCtx {
	t := c.paramSchema.Properties[key]
	if obj, ok := t.(Object); ok {
		return &ReqCtx{c.req.Object(key, jsonMap{}), obj}
	}
	return nil
}

func (c *ReqCtx) ObjectArray(key string) []*ReqCtx {
	t := c.paramSchema.Properties[key]
	if oa, ok := t.(ObjectArray); ok {
		data := c.req.ObjectArray(key, []jsonMap{})
		if len(data) < 1 {
			return nil
		}

		ret := make([]*ReqCtx, len(data))
		for i, jsonMap := range data {
			ret[i] = &ReqCtx{jsonMap, Object{Properties: oa.Items}}
		}
		return ret
	}
	return nil
}

func (c *ReqCtx) Req() Req {
	return c.req
}
