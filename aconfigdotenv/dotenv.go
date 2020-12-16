package aconfigdotenv

import (
	"github.com/joho/godotenv"
)

// Decoder of .ENV files for aconfig.
type Decoder struct{}

// New .ENV decoder for aconfig.
func New() *Decoder { return &Decoder{} }

// DecodeFile implements aconfig.FileDecoder.
func (d *Decoder) DecodeFile(filename string) (map[string]interface{}, error) {
	raw, err := godotenv.Read(filename)
	if err != nil {
		return nil, err
	}

	res := map[string]interface{}{}

	for key, value := range raw {
		res[key] = value
	}
	return res, nil
}
