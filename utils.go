package aconfig

import (
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
		switch true {
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
			runes[len(runes)-1] = append(runes[len(runes)-1], r)
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
