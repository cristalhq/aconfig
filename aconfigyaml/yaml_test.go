package aconfigyaml_test

import (
	"reflect"
	"testing"

	"github.com/cristalhq/aconfig"
	"github.com/cristalhq/aconfig/aconfigyaml"
)

func TestYAML(t *testing.T) {
	var cfg, want TestConfig

	i := int32(42)
	want = TestConfig{
		A: "b",
		C: 10,
		E: 123.456,
		B: []byte("abc"),
		P: &i,
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
		StructM: StructM{
			M: "n",
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
	B []byte
	P *int32
	Y structY

	AA structA `yaml:"A"`
	StructM
}

type structY struct {
	X string
	Z []string
	A struct {
		I bool
	}
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

type StructM struct {
	M string
}
