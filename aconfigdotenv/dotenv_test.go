package aconfigdotenv_test

import (
	"os"
	"reflect"
	"testing"

	"github.com/cristalhq/aconfig"
	"github.com/cristalhq/aconfig/aconfigdotenv"
)

func TestDotEnv(t *testing.T) {
	filepath := createTestFile(t)

	var cfg structConfig
	loader := aconfig.LoaderFor(&cfg, aconfig.Config{
		SkipDefaults: true,
		SkipEnv:      true,
		SkipFlags:    true,
		FileDecoders: map[string]aconfig.FileDecoder{
			".env": aconfigdotenv.New(),
		},
		Files: []string{filepath},
	})

	if err := loader.Load(); err != nil {
		t.Fatal(err)
	}

	i := int32(42)
	j := int64(420)
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

	filepath := dir + "/testfile.env"

	f, err := os.Create(filepath)
	if err != nil {
		t.Fatal(err)
	}
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
	BB []byte `env:"b"`
}

type StructM struct {
	M string
}

const testfileContent = `
A=b
C=10
E=123.456
B=abc
I=42
J=420

Y_X=y
Y_Z=1,2,3

A_x=y
A_B_C_m=n
A_B_C_b=boo
A_B_D=x,y,z

M=n
`
