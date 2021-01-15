package aconfigtoml

import (
	"fmt"
	"os"
	"strings"

	"github.com/BurntSushi/toml"
)

// Decoder of TOML files for aconfig.
type Decoder struct{}

// New TOML decoder for aconfig.
func New() *Decoder { return &Decoder{} }

// DecodeFile implements aconfig.FileDecoder.
func (d *Decoder) DecodeFile(filename string) (map[string]interface{}, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	var raw map[string]interface{}
	if _, err := toml.DecodeReader(f, &raw); err != nil {
		return nil, err
	}

	res := map[string]interface{}{}

	for key, value := range raw {
		flatten("", key, value, res)
	}
	return res, nil
}

// copied from aconfig/utils.go
//
func flatten(prefix, key string, curr interface{}, res map[string]interface{}) {
	switch curr := curr.(type) {
	case map[string]interface{}:
		for k, v := range curr {
			flatten(prefix+key+".", k, v, res)
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
	case int64:
		res[prefix+key] = fmt.Sprint(curr)
	default:
		panic(fmt.Sprintf("%s::%s got %T %v", prefix, key, curr, curr))
	}
}
