package aconfigtoml

import (
	"os"

	"github.com/BurntSushi/toml"
)

// Decoder of TOML files for aconfig.
type Decoder struct{}

// New TOML decoder for aconfig.
func New() *Decoder { return &Decoder{} }

// Format of the decoder.
func (d *Decoder) Format() string {
	return "toml"
}

// DecodeFile implements aconfig.FileDecoder.
func (d *Decoder) DecodeFile(filename string) (map[string]interface{}, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var raw map[string]interface{}
	if _, err := toml.DecodeReader(f, &raw); err != nil {
		return nil, err
	}
	return raw, nil
}
