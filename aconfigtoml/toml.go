package aconfigtoml

import (
	"io/fs"

	"github.com/pelletier/go-toml/v2"
)

// Decoder of TOML files for aconfig.
type Decoder struct {
	fsys fs.FS
}

// New TOML decoder for aconfig.
func New() *Decoder { return &Decoder{} }

// Format of the decoder.
func (d *Decoder) Format() string {
	return "toml"
}

// DecodeFile implements aconfig.FileDecoder.
func (d *Decoder) DecodeFile(filename string) (map[string]interface{}, error) {
	f, err := d.fsys.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var raw map[string]interface{}
	if err := toml.NewDecoder(f).Decode(&raw); err != nil {
		return nil, err
	}
	return raw, nil
}

// DecodeFile implements aconfig.FileDecoder.
func (d *Decoder) Init(fsys fs.FS) {
	d.fsys = fsys
}
