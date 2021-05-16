package aconfig

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
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

func getEnv() map[string]interface{} {
	env := os.Environ()
	res := make(map[string]interface{}, len(env))

	for _, s := range env {
		for j := 0; j < len(s); j++ {
			if s[j] == '=' {
				key, value := s[:j], s[j+1:]
				res[key] = value
				break
			}
		}
	}
	return res
}

func getFlags(flagSet *flag.FlagSet) map[string]interface{} {
	res := map[string]interface{}{}
	flagSet.Visit(func(f *flag.Flag) {
		res[f.Name] = f.Value
	})
	return res
}

func getActualFlag(name string, flagSet *flag.FlagSet) *flag.Flag {
	var found *flag.Flag
	flagSet.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = f
		}
	})
	return found
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

	name := strings.Join(words, "_")
	if tag == envNameTag {
		return strings.ToUpper(name)
	}
	return strings.ToLower(name)
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

// ...
func (d *jsonDecoder) EncodeTo(w io.Writer, data map[string]interface{}) error {
	return json.NewEncoder(w).Encode(data)
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

func find(actualFields map[string]interface{}, name string) map[string]interface{} {
	if strings.LastIndex(name, ".") == -1 {
		return actualFields
	}

	subName := name[:strings.LastIndex(name, ".")]
	value, ok := actualFields[subName]
	if !ok {
		actualFields = find(actualFields, subName)
		value, ok = actualFields[subName]
		if !ok {
			return actualFields
		}
	}

	switch val := value.(type) {
	case map[string]interface{}:
		for k, v := range val {
			actualFields[subName+"."+k] = v
		}
		delete(actualFields, subName)
	case map[interface{}]interface{}:
		for k, v := range val {
			actualFields[subName+"."+fmt.Sprint(k)] = v
		}
		delete(actualFields, subName)
	case []map[string]interface{}:
		for _, m := range val {
			for k, v := range m {
				actualFields[subName+"."+k] = v
			}
		}
		delete(actualFields, subName)
	case []map[interface{}]interface{}:
		for _, m := range val {
			for k, v := range m {
				actualFields[subName+"."+fmt.Sprint(k)] = v
			}
		}
		delete(actualFields, subName)
	}
	return actualFields
}
