package aconfighcl

import (
	"fmt"

	"github.com/hashicorp/hcl/v2/hclsimple"
)

// Decoder of HCL files for aconfig.
type Decoder struct{}

// New HCL decoder for aconfig.
func New() *Decoder { return &Decoder{} }

// DecodeFile implements aconfig.FileDecoder.
func (d *Decoder) DecodeFile(filename string) (map[string]interface{}, error) {
	var raw map[string]interface{}
	if err := hclsimple.DecodeFile(filename, nil, &raw); err != nil {
		return nil, err
	}

	res := map[string]interface{}{}

	for key, value := range raw {
		fmt.Printf("k: %s, v: %v %[2]T\n", key, value)
		flatten("", key, value, res)
	}
	fmt.Printf("map: %#v\n\n", res)
	return res, nil
}

func flatten(prefix, key string, curr interface{}, res map[string]interface{}) {
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
		res[prefix+key] = curr
	case string:
		res[prefix+key] = curr
	case float64:
		res[prefix+key] = fmt.Sprintf("%v", curr)
	case int, int8, int16, int32, int64:
		res[prefix+key] = fmt.Sprintf("%v", curr)
	}
}
