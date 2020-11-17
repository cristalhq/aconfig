package aconfigtoml

import (
	"os"

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

	var dst map[string]interface{}
	if _, err := toml.DecodeReader(f, &dst); err != nil {
		return nil, err
	}
	return dst, nil
}
