package aconfigyaml

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v2"
)

// Decoder of YAML files for aconfig.
type Decoder struct{}

// New YAML decoder for aconfig.
func New() *Decoder { return &Decoder{} }

// DecodeFile implements aconfig.FileDecoder.
func (d *Decoder) DecodeFile(filename string) (map[string]interface{}, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	var raw map[string]interface{}
	if err := yaml.NewDecoder(f).Decode(&raw); err != nil {
		return nil, err
	}

	res := map[string]interface{}{}

	for key, value := range raw {
		// fmt.Printf("k: %s, v: %v %[2]T\n", key, value)
		flatten("", key, value, res)
	}
	// fmt.Printf("map: %#v\n\n", res)
	return res, nil
}

// copied from aconfig/utils.go
//
func flatten(prefix, key string, curr interface{}, res map[string]interface{}) {
	// log.Printf("%s::%s got %T %v", prefix, key, curr, curr)
	switch curr := curr.(type) {
	case map[string]interface{}:
		for k, v := range curr {
			flatten(prefix+key+".", k, v, res)
		}

	case map[interface{}]interface{}:
		for k, v := range curr {
			if k, ok := k.(string); ok {
				flatten(prefix+key+".", k, v, res)
			}
		}
	case []interface{}:
		// fmt.Printf("wow %v\n", curr)
		b := &strings.Builder{}
		for i, v := range curr {
			if i > 0 {
				b.WriteByte(',')
			}
			b.WriteString(fmt.Sprint(v))
		}
		res[prefix+key] = b.String()
	case []string:
		res[prefix+key] = curr
	case string:
		res[prefix+key] = curr
	case float32:
		res[prefix+key] = fmt.Sprintf("%v", curr)
	case float64:
		res[prefix+key] = fmt.Sprintf("%v", curr)
	case int, int8, int16, int32, int64:
		res[prefix+key] = fmt.Sprintf("%v", curr)
	default:
		panic(fmt.Sprintf("%s::%s got %T %v", prefix, key, curr, curr))
	}
}
