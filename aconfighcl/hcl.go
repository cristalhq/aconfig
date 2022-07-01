package aconfighcl

import (
	"io/fs"
	"io/ioutil"

	"github.com/hashicorp/hcl"
)

// Decoder of HCL files for aconfig.
type Decoder struct {
	fsys fs.FS
}

// New HCL decoder for aconfig.
func New() *Decoder { return &Decoder{} }

// Format of the decoder.
func (d *Decoder) Format() string {
	return "hcl"
}

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

// DecodeFile implements aconfig.FileDecoder.
func (d *Decoder) Init(fsys fs.FS) {
	d.fsys = fsys
}
