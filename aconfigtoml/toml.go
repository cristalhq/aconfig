package aconfigtoml

import (
	"os"

	"github.com/cristalhq/aconfig"

	"github.com/BurntSushi/toml"
)

var _ aconfig.FileDecoder = &Decoder{}

// Decoder of TOML files for aconfig.
type Decoder struct{}

// DecodeFile implements aconfig.FileDecoder.
func (d *Decoder) DecodeFile(filename string, dst interface{}) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	_, err = toml.DecodeReader(f, dst)
	return err
}
