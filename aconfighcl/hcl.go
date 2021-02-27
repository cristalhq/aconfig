package aconfighcl

import (
	"io/ioutil"

	"github.com/hashicorp/hcl"
)

// Decoder of HCL files for aconfig.
type Decoder struct{}

// New HCL decoder for aconfig.
func New() *Decoder { return &Decoder{} }

// DecodeFile implements aconfig.FileDecoder.
func (d *Decoder) DecodeFile(filename string) (map[string]interface{}, error) {
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	f, err := hcl.ParseBytes(b)
	if err != nil {
		return nil, err
	}

	var raw map[string]interface{}
	if err := hcl.DecodeObject(&raw, f); err != nil {
		return nil, err
	}
	return raw, nil
}
