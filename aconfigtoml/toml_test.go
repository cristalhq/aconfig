package aconfigtoml_test

import (
	"reflect"
	"testing"

	"github.com/cristalhq/aconfig"
	"github.com/cristalhq/aconfig/aconfigtoml"
)

func TestTOML(t *testing.T) {
	var cfg, want TestConfig
	want = TestConfig{
		A: "b",
		C: 10,
		E: 123.456,
		Y: structY{
			X: "y",
			Z: []string{"1", "2", "3"},
		},
		AA: structA{
			X: "y",
			BB: structB{
				CC: structC{
					MM: "n",
				},
				DD: []string{"x", "y", "z"},
			},
		},
	}

	loader := aconfig.LoaderFor(&cfg, aconfig.Config{
		SkipDefaults:    true,
		SkipEnvironment: true,
		SkipFlags:       true,
		FileDecoders: map[string]aconfig.FileDecoder{
			".toml": aconfigtoml.New(),
		},
		Files: []string{"testfile.toml"},
	})

	if err := loader.Load(); err != nil {
		t.Fatal(err)
	}
	if got := cfg; !reflect.DeepEqual(got, want) {
		t.Fatalf("want %v, got %v", want, got)
	}
}

type TestConfig struct {
	A string
	C int
	E float64
	Y structY

	AA structA `toml:"A"`
}

type structY struct {
	X string
	Z []string
}

type structA struct {
	X  string  `toml:"x"`
	BB structB `toml:"B"`
}

type structB struct {
	CC structC  `toml:"C"`
	DD []string `toml:"D"`
}

type structC struct {
	MM string `toml:"m"`
}

// func TestTOML(t *testing.T) {
// 	var cfg, want TestConfig

// 	filename := createFile(t)
// 	loadFile(t, filename, &want)

// 	loader := aconfig.LoaderFor(&cfg).
// 		SkipDefaults().SkipEnvironment().SkipFlags().
// 		WithFileDecoder(".toml", aconfigtoml.New()).
// 		WithFiles([]string{filename}).
// 		Build()

// 	if err := loader.Load(); err != nil {
// 		t.Fatal(err)
// 	}

// 	if got := cfg; !reflect.DeepEqual(got, want) {
// 		t.Fatalf("want %v, got %v", want, got)
// 	}
// }

// func createFile(t *testing.T) string {
// 	filename := t.TempDir() + "/config.toml"

// 	file, err := os.Create(filename)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	defer file.Close()

// 	file.WriteString(`str = "str-toml"
// int = 101
// http_port = 65000
// [sub]
// float = 999.111 `)

// 	return filename
// }

// func loadFile(t *testing.T, file string, dst interface{}) {
// 	t.Helper()

// 	f, err := os.Open(file)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	_, err = toml.DecodeReader(f, dst)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// }

// type TestConfig struct {
// 	Str      string    `toml:"str"`
// 	Bytes    []byte    `toml:"bytes"`
// 	Int      *int32    `toml:"int"`
// 	HTTPPort int       `toml:"http_port"`
// 	Param    int       `toml:"param"` // no default tag, so default value
// 	Sub      SubConfig `toml:"sub"`
// 	Anon     struct {
// 		IsAnon bool `default:"true"`
// 	}

// 	Slice []int          `usage:"just pass elements"`
// 	Map1  map[string]int ``
// 	Map2  map[int]string ``

// 	EmbeddedConfig
// }

// type EmbeddedConfig struct {
// 	Em string `usage:"use... em...field." toml:"em"`
// }

// type SubConfig struct {
// 	Float float64 `toml:"float"`
// }
