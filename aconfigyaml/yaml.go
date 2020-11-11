package aconfigyaml

import (
	"os"

	"github.com/cristalhq/aconfig"

	"gopkg.in/yaml.v2"
)

var _ aconfig.FileDecoder = &Decoder{}

// Decoder of YAML files for aconfig.
type Decoder struct{}

// DecodeFile implements aconfig.FileDecoder.
func (d *Decoder) DecodeFile(filename string, dst interface{}) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	return yaml.NewDecoder(f).Decode(dst)
}
