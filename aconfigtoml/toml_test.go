package aconfigtoml_test

import (
	"embed"
	"os"
	"reflect"
	"testing"

	"github.com/cristalhq/aconfig"
	"github.com/cristalhq/aconfig/aconfigtoml"
)

//go:embed testdata
var configEmbed embed.FS

func TestTOMLEmbed(t *testing.T) {
	var cfg struct {
		Foo string
		Bar string
	}
	loader := aconfig.LoaderFor(&cfg, aconfig.Config{
		SkipDefaults:       true,
		SkipEnv:            true,
		SkipFlags:          true,
		FailOnFileNotFound: true,
		FileDecoders: map[string]aconfig.FileDecoder{
			".toml": aconfigtoml.New(),
		},
		Files:      []string{"testdata/config.toml"},
		FileSystem: configEmbed,
	})

	if err := loader.Load(); err != nil {
		t.Fatal(err)
	}

	if cfg.Foo != "value1" {
		t.Fatalf("have: %v", cfg.Foo)
	}
	if cfg.Bar != "value2" {
		t.Fatalf("have: %v", cfg.Bar)
	}
}

func TestTOML(t *testing.T) {
	filepath := createTestFile(t)

	var cfg structConfig
	loader := aconfig.LoaderFor(&cfg, aconfig.Config{
		SkipDefaults: true,
		SkipEnv:      true,
		SkipFlags:    true,
		FileDecoders: map[string]aconfig.FileDecoder{
			".toml": aconfigtoml.New(),
		},
		Files: []string{filepath},
	})

	if err := loader.Load(); err != nil {
		t.Fatal(err)
	}

	i := int32(42)
	j := int64(420)
	mInterface := make([]interface{}, 2)
	for iI, vI := range []string{"q", "w"} {
		mInterface[iI] = vI
	}
	want := structConfig{
		A: "b",
		C: 10,
		E: 123.456,
		B: []byte("abc"),
		I: &i,
		J: &j,
		Y: structY{
			X: "y",
			Z: []string{"1", "2", "3"},
			A: structD{
				I: true,
			},
		},
		AA: structA{
			X: "y",
			BB: structB{
				CC: structC{
					MM: "n",
					BB: []byte("boo"),
				},
				DD: []string{"x", "y", "z"},
			},
		},
		StructM: StructM{
			M: "n",
		},
		MI: mInterface,
	}

	if got := cfg; !reflect.DeepEqual(want, got) {
		t.Fatalf("want %v, got %v", want, got)
	}
}

func createTestFile(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Cleanup(func() {
		os.RemoveAll(dir)
	})

	filepath := dir + "/testfile.toml"

	f, err := os.Create(filepath)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	_, err = f.WriteString(testfileContent)
	if err != nil {
		t.Fatal(err)
	}
	return filepath
}

type structConfig struct {
	A string
	C int
	E float64
	B []byte
	I *int32
	J *int64
	Y structY

	AA structA `toml:"A"`
	StructM
	MI interface{} `toml:"MI"`
}

type structY struct {
	X string
	Z []string
	A structD
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
	BB []byte `toml:"b"`
}

type structD struct {
	I bool
}

type StructM struct {
	M string
}

const testfileContent = `
a = "b"
c = 10
e = 123.456
b = "abc"
i = 42
j = 420
m = "n"
MI = ["q", "w"]

[y]
x = "y"
z = [ 1, 2, 3 ]
    [y.a]
    i = true

[A]
    x = "y"

	[A.B]
	D = ["x", "y", "z"]

    [A.B.C]
	m = "n"
	b = "boo"
`
