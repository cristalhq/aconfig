package aconfigdotenv

import (
	"github.com/joho/godotenv"
)

// Decoder of DotENV files for aconfig.
type Decoder struct{}

// New .ENV decoder for aconfig.
func New() *Decoder { return &Decoder{} }

// Format of the decoder.
func (d *Decoder) Format() string {
	return "env"
}

// DecodeFile implements aconfig.FileDecoder.
func (d *Decoder) DecodeFile(filename string) (map[string]interface{}, error) {
	raw, err := godotenv.Read(filename)
	if err != nil {
		return nil, err
	}

	res := make(map[string]interface{}, len(raw))
	for key, value := range raw {
		res[key] = value
	}
	return res, nil
}
