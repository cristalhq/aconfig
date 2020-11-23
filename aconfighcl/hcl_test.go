package aconfighcl_test

import (
	"os"
	"reflect"
	"testing"

	"github.com/cristalhq/aconfig"
	"github.com/cristalhq/aconfig/aconfighcl"

	"github.com/hashicorp/hcl/v2/hclsimple"
)

func TestHCL(t *testing.T) {
	var want TestConfig

	filename := createFile(t)
	loadFile(t, filename, &want)

	var cfg TestConfig
	loader := aconfig.LoaderFor(&cfg).
		SkipDefaults().
		SkipEnvironment().
		SkipFlags().
		WithFileDecoder(".hcl", aconfighcl.New()).
		Build()

	if err := loader.LoadWithFile(filename); err != nil {
		t.Fatal(err)
	}

	if got := cfg; !reflect.DeepEqual(got, want) {
		t.Fatalf("want %v, got %v", want, got)
	}
}

func createFile(t *testing.T) string {
	filename := t.TempDir() + "/config.hcl"

	file, err := os.Create(filename)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	file.WriteString(`str      = "str-hcl"
bytes    = [89,110,108,48,90,88,77,116,90,87,53,50]
int      = 101
http_port = 65000
sub {
	float = 999.111
}`)

	return filename
}

func loadFile(t *testing.T, file string, dst interface{}) {
	t.Helper()

	err := hclsimple.DecodeFile(file, nil, dst)
	if err != nil {
		t.Fatal(err)
	}
}

type TestConfig struct {
	Str      string    `hcl:"str"`
	Bytes    []byte    `hcl:"bytes"`
	Int      *int32    `hcl:"int"`
	HTTPPort int       `hcl:"http_port"`
	Param    int       // no default tag, so default value
	Sub      SubConfig `hcl:"sub,block"`
	Anon     struct {
		IsAnon bool `default:"true"`
	}

	Slice []int          `usage:"just pass elements"`
	Map1  map[string]int ``
	Map2  map[int]string ``

	EmbeddedConfig
}

type EmbeddedConfig struct {
	Em string `usage:"use... em...field." hcl:"em"`
}

type SubConfig struct {
	Float float64 `hcl:"float"`
}
