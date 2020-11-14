package aconfighcl

import (
	"github.com/hashicorp/hcl/v2/hclsimple"
)

// Decoder of HCL files for aconfig.
type Decoder struct{}

// New HCL decoder for aconfig.
func New() *Decoder { return &Decoder{} }

// DecodeFile implements aconfig.FileDecoder.
func (d *Decoder) DecodeFile(filename string, dst interface{}) error {
	return hclsimple.DecodeFile(filename, nil, dst)
}
