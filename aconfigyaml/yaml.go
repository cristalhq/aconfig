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
		flatten("", key, value, res)
	}
	return res, nil
}

// copied and adapted from aconfig/utils.go
//
func flatten(prefix, key string, curr interface{}, res map[string]interface{}) {
	switch curr := curr.(type) {
	case map[interface{}]interface{}:
		for k, v := range curr {
			if k, ok := k.(string); ok {
				flatten(prefix+key+".", k, v, res)
			}
		}
	case []interface{}:
		b := &strings.Builder{}
		for i, v := range curr {
			if i > 0 {
				b.WriteByte(',')
			}
			b.WriteString(fmt.Sprint(v))
		}
		res[prefix+key] = b.String()
	case string:
		res[prefix+key] = curr
	case bool:
		res[prefix+key] = fmt.Sprint(curr)
	case float64:
		res[prefix+key] = fmt.Sprint(curr)
	case int, int8, int16, int32:
		res[prefix+key] = fmt.Sprint(curr)
	default:
		panic(fmt.Sprintf("%s::%s got %T %v", prefix, key, curr, curr))
	}
}
