package aconfigyaml

import (
	"os"

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

	var dst map[string]interface{}
	if err := yaml.NewDecoder(f).Decode(&dst); err != nil {
		return nil, err
	}
	return dst, nil
}
