package aconfigyaml_test

import (
	"reflect"
	"testing"

	"github.com/cristalhq/aconfig"
	"github.com/cristalhq/aconfig/aconfigyaml"
)

func TestYAML(t *testing.T) {
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
			".yaml": aconfigyaml.New(),
		},
		Files: []string{"testfile.yaml"},
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

	AA structA `yaml:"A"`
}

type structY struct {
	X string
	Z []string
}

type structA struct {
	X  string  `yaml:"x"`
	BB structB `yaml:"B"`
}

type structB struct {
	CC structC  `yaml:"C"`
	DD []string `yaml:"D"`
}

type structC struct {
	MM string `yaml:"m"`
}
