package aconfighcl

import (
	"github.com/hashicorp/hcl/v2/hclsimple"
)

// Decoder of HCL files for aconfig.
type Decoder struct{}

// New HCL decoder for aconfig.
func New() *Decoder { return &Decoder{} }

// DecodeFile implements aconfig.FileDecoder.
func (d *Decoder) DecodeFile(filename string) (map[string]interface{}, error) {
	var dst map[string]interface{}
	if err := hclsimple.DecodeFile(filename, nil, &dst); err != nil {
		return nil, err
	}
	return dst, nil
}
