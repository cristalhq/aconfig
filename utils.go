package aconfig

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
	"unicode"
)

func assertStruct(x interface{}) {
	if x == nil {
		panic("aconfig: nil should not be passed to the Loader")
	}
	value := reflect.ValueOf(x)
	for value.Type().Kind() == reflect.Ptr {
		value = value.Elem()
	}
	if value.Kind() != reflect.Struct {
		panic("aconfig: only struct can be passed to the Loader")
	}
}

func getEnv() map[string]string {
	env := os.Environ()
	res := make(map[string]string, len(env))

	for _, s := range env {
		for j := 0; j < len(s); j++ {
			if s[j] != '=' {
				continue
			}

			key, value := s[:j], s[j+1:]
			res[key] = value
		}
	}
	return res
}

func makeName(name string, parent *fieldData) string {
	if parent == nil {
		return name
	}
	return parent.name + "." + name
}

func (l *Loader) makeTagValue(field reflect.StructField, tag string, words []string) string {
	if v := field.Tag.Get(tag); v != "" {
		return v
	}
	switch tag {
	case jsonNameTag, yamlNameTag, tomlNameTag, hclNameTag:
		if l.config.DontGenerateTags {
			return field.Name
		}
	}
	return strings.ToLower(makeParsingName(words))
}

func (l *Loader) makeEnvName(field reflect.StructField, words []string) string {
	if v := field.Tag.Get(envNameTag); v != "" {
		return v
	}
	if l.config.DontGenerateTags {
		return ""
	}
	return strings.ToUpper(makeParsingName(words))
}

func makeParsingName(words []string) string {
	var name strings.Builder
	for i, w := range words {
		if i > 0 {
			name.WriteByte('_')
		}
		name.WriteString(strings.ToLower(w))
	}
	return name.String()
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

	words := make([]string, 0, len(runes))
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
	defer f.Close()

	var raw map[string]interface{}
	if err := json.NewDecoder(f).Decode(&raw); err != nil {
		return nil, err
	}
	return raw, nil
}

func normalize(curr interface{}) interface{} {
	switch curr := curr.(type) {
	case []interface{}:
		b := &strings.Builder{}
		for i, v := range curr {
			if i > 0 {
				b.WriteByte(',')
			}
			b.WriteString(fmt.Sprint(v))
		}
		return b.String()
	case string:
		return curr
	case float64:
		return fmt.Sprint(curr)
	case bool:
		return fmt.Sprint(curr)
	default:
		panic(fmt.Sprintf("Can't normalize %T %v", curr, curr))
	}
}
