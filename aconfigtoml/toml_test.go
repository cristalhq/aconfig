package aconfigtoml_test

import (
	"os"
	"reflect"
	"testing"

	"github.com/cristalhq/aconfig"
	"github.com/cristalhq/aconfig/aconfigtoml"

	"github.com/BurntSushi/toml"
)

func TestTOML(t *testing.T) {
	var want TestConfig

	filename := createFile(t)
	loadFile(t, filename, &want)

	var cfg TestConfig
	loader := aconfig.LoaderFor(&cfg).
		SkipDefaults().
		SkipEnvironment().
		SkipFlags().
		WithFileDecoder(".toml", aconfigtoml.New()).
		Build()

	if err := loader.LoadWithFile(filename); err != nil {
		t.Fatal(err)
	}

	if got := cfg; !reflect.DeepEqual(got, want) {
		t.Fatalf("want %v, got %v", want, got)
	}
}

func createFile(t *testing.T) string {
	filename := t.TempDir() + "/config.toml"

	file, err := os.Create(filename)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	file.WriteString(`Str = "str-json"
Int = 101
HTTPPort = 65000
[Sub]
Float = 999.111 `)

	return filename
}

func loadFile(t *testing.T, file string, dst interface{}) {
	t.Helper()

	f, err := os.Open(file)
	if err != nil {
		t.Fatal(err)
	}
	_, err = toml.DecodeReader(f, dst)
	if err != nil {
		t.Fatal(err)
	}
}

type TestConfig struct {
	Str      string    `toml:"str"`
	Bytes    []byte    `toml:"bytes"`
	Int      *int32    `toml:"int"`
	HTTPPort int       `toml:"http_port"`
	Param    int       `toml:"param"` // no default tag, so default value
	Sub      SubConfig `toml:"sub"`
	Anon     struct {
		IsAnon bool `default:"true"`
	}

	Slice []int          `usage:"just pass elements"`
	Map1  map[string]int ``
	Map2  map[int]string ``

	EmbeddedConfig
}

type EmbeddedConfig struct {
	Em string `usage:"use... em...field." toml:"em"`
}

type SubConfig struct {
	Float float64 `toml:"float"`
}
