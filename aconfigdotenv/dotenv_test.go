package aconfigdotenv_test

import (
	"reflect"
	"testing"

	"github.com/cristalhq/aconfig"
	"github.com/cristalhq/aconfig/aconfigdotenv"
)

func TestDOTENV(t *testing.T) {
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
			".env": aconfigdotenv.New(),
		},
		Files: []string{"testfile.env"},
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

	AA structA `env:"A"`
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
	X  string  `env:"x"`
	BB structB `env:"B"`
}

type structB struct {
	CC structC  `env:"C"`
	DD []string `env:"D"`
}

type structC struct {
	MM string `env:"m"`
}

type StructM struct {
	M string
}
