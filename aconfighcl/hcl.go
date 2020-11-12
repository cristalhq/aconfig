package aconfighcl

import (
	"github.com/cristalhq/aconfig"

	"github.com/hashicorp/hcl/v2/hclsimple"
)

var _ aconfig.FileDecoder = &Decoder{}

// Decoder of HCL files for aconfig.
type Decoder struct{}

// New HCL decoder for aconfig.
func New() *Decoder { return &Decoder{} }

// DecodeFile implements aconfig.FileDecoder.
func (d *Decoder) DecodeFile(filename string, dst interface{}) error {
	return hclsimple.DecodeFile(filename, nil, dst)
}
