package aconfig

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
	"unicode"
)

func makeName(name string, parent *fieldData) string {
	if parent == nil {
		return name
	}
	return parent.name + "." + name
}

func makeEnvName(field reflect.StructField, parent *fieldData, words []string) string {
	envName := field.Tag.Get(envNameTag)
	if envName == "" {
		envName = makeParsingName(words)
	}
	if parent != nil {
		envName = parent.envName + "_" + envName
	}
	return strings.ToUpper(envName)
}

func makeFlagName(field reflect.StructField, parent *fieldData, words []string) string {
	flagName := field.Tag.Get(flagNameTag)
	if flagName == "" {
		flagName = makeParsingName(words)
	}
	if parent != nil {
		flagName = parent.flagName + "." + flagName
	}
	return strings.ToLower(flagName)
}

func makeParsingName(words []string) string {
	name := ""
	for i, w := range words {
		if i > 0 {
			name += "_"
		}
		name += strings.ToLower(w)
	}
	return name
}

// based on https://github.com/fatih/camelcase
func splitNameByWords(src string) []string {
	var runes [][]rune
	lastClass, class := 0, 0

	// split into fields based on class of unicode character
	for _, r := range src {
		switch {
		case unicode.IsLower(r):
			class = 1
		case unicode.IsUpper(r):
			class = 2
		case unicode.IsDigit(r):
			class = 3
		default:
			class = 4
		}
		if class == lastClass {
			sz := len(runes) - 1
			runes[sz] = append(runes[sz], r)
		} else {
			runes = append(runes, []rune{r})
		}
		lastClass = class
	}

	// handle upper case -> lower case sequences, e.g.
	// "PDFL", "oader" -> "PDF", "Loader"
	for i := 0; i < len(runes)-1; i++ {
		if unicode.IsUpper(runes[i][0]) && unicode.IsLower(runes[i+1][0]) {
			runes[i+1] = append([]rune{runes[i][len(runes[i])-1]}, runes[i+1]...)
			runes[i] = runes[i][:len(runes[i])-1]
		}
	}

	words := []string{}
	for _, s := range runes {
		if len(s) > 0 {
			words = append(words, string(s))
		}
	}
	return words
}

type jsonDecoder struct{}

// DecodeFile implements FileDecoder.
func (d *jsonDecoder) DecodeFile(filename string) (map[string]interface{}, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	var raw map[string]interface{}
	if err := json.NewDecoder(f).Decode(&raw); err != nil {
		return nil, err
	}

	res := map[string]interface{}{}

	for key, value := range raw {
		// fmt.Printf("k: %s, v: %v %[2]T\n", key, value)
		flatten("", key, value, res)
	}
	// fmt.Printf("map: %#v\n\n", res)
	return res, nil
}

func flatten(prefix, key string, curr interface{}, res map[string]interface{}) {
	switch curr := curr.(type) {
	case map[string]interface{}:
		for k, v := range curr {
			flatten(prefix+key+".", k, v, res)
		}
	case []interface{}:
		res[prefix+key] = curr
	case string:
		res[prefix+key] = curr
	case float64:
		res[prefix+key] = fmt.Sprintf("%v", curr)
	}
}
