package aconfigyaml_test

import (
	"os"
	"reflect"
	"testing"

	"github.com/cristalhq/aconfig"
	"github.com/cristalhq/aconfig/aconfigyaml"

	"gopkg.in/yaml.v2"
)

func TestYAML(t *testing.T) {
	var cfg, want TestConfig

	filename := createFile(t)
	loadFile(t, filename, &want)

	loader := aconfig.LoaderFor(&cfg).
		SkipDefaults().SkipEnvironment().SkipFlags().
		WithFileDecoder(".yaml", aconfigyaml.New()).
		WithFiles([]string{filename}).
		Build()

	if err := loader.Load(); err != nil {
		t.Fatal(err)
	}

	if got := cfg; !reflect.DeepEqual(got, want) {
		t.Fatalf("want %v, got %v", want, got)
	}
}

func createFile(t *testing.T) string {
	filename := t.TempDir() + "/config.yaml"

	file, err := os.Create(filename)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	file.WriteString(`str: "str-yaml"
int: 101
http_port: 65000
sub:
  float: 999`)

	return filename
}

func loadFile(t *testing.T, file string, dst interface{}) {
	t.Helper()

	f, err := os.Open(file)
	if err != nil {
		t.Fatal(err)
	}
	err = yaml.NewDecoder(f).Decode(dst)
	if err != nil {
		t.Fatal(err)
	}
}

type TestConfig struct {
	Str      string    `yaml:"str"`
	Bytes    []byte    `yaml:"bytes"`
	Int      *int32    `yaml:"int"`
	HTTPPort int       `yaml:"http_port"`
	Param    int       `yaml:"param"` // no default tag, so default value
	Sub      SubConfig `yaml:"sub"`
	Anon     struct {
		IsAnon bool `default:"true"`
	}

	Slice []int          `usage:"just pass elements"`
	Map1  map[string]int ``
	Map2  map[int]string ``

	EmbeddedConfig
}

type EmbeddedConfig struct {
	Em string `usage:"use... em...field." yaml:"em"`
}

type SubConfig struct {
	Float float64 `yaml:"float"`
}
