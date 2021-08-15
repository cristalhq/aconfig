package aconfigyaml_test

import (
	"os"
	"reflect"
	"testing"

	"github.com/cristalhq/aconfig"
	"github.com/cristalhq/aconfig/aconfigyaml"
)

func TestYAML(t *testing.T) {
	filepath := createTestFile(t, "file.yaml", testfileContent)

	var cfg structConfig
	loader := aconfig.LoaderFor(&cfg, aconfig.Config{
		SkipDefaults: true,
		SkipEnv:      true,
		SkipFlags:    true,
		FileDecoders: map[string]aconfig.FileDecoder{
			".yaml": aconfigyaml.New(),
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

type ConfigTest struct {
	VCenter ConfigVCenter `yaml:"vcenter"`
}

type ConfigVCenter struct {
	User        string `yaml:"user"`
	Password    string `yaml:"password"`
	Port        string `yaml:"port"`
	Datacenters []ConfigVCenterDCRegion
}

type ConfigVCenterDCRegion struct {
	Region    string `yaml:"region"`
	Addresses []ConfigVCenterDC
}

type ConfigVCenterDC struct {
	Zone       string `yaml:"zone"`
	Address    string `yaml:"address"`
	Datacenter string `yaml:"datacenter"`
}

func TestSliceStructs(t *testing.T) {
	content := `vcenter:
  user: user-test
  password: pass-test
  port: 8080
  datacenters:
    - region: region-test
      addresses:
      - zone: zone-test
        address: address-test
        datacenter: datacenter-test
`

	file := createTestFile(t, "test.yaml", content)

	var cfg ConfigTest
	loader := aconfig.LoaderFor(&cfg, aconfig.Config{
		SkipDefaults:       true,
		SkipEnv:            true,
		SkipFlags:          true,
		FailOnFileNotFound: true,
		Files:              []string{file},
		FileDecoders: map[string]aconfig.FileDecoder{
			".yaml": aconfigyaml.New(),
		},
	})

	if err := loader.Load(); err != nil {
		t.Fatal(err)
	}

	want := ConfigTest{
		VCenter: ConfigVCenter{
			User:     "user-test",
			Password: "pass-test",
			Port:     "8080",
			Datacenters: []ConfigVCenterDCRegion{
				{
					Region: "region-test",
					Addresses: []ConfigVCenterDC{
						{
							Zone:       "zone-test",
							Address:    "address-test",
							Datacenter: "datacenter-test",
						},
					},
				},
			},
		},
	}
	if got := cfg; !reflect.DeepEqual(want, got) {
		t.Fatalf("want %v, got %v", want, got)
	}
}

func createTestFile(t *testing.T, name, content string) string {
	t.Helper()
	dir := t.TempDir()
	t.Cleanup(func() {
		os.RemoveAll(dir)
	})

	filepath := dir + name

	f, err := os.Create(filepath)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	_, err = f.WriteString(content)
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

	AA structA `yaml:"A"`
	StructM
	MI interface{} `yaml:"MI"`
}

type structY struct {
	X string
	Z []string
	A structD
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
	BB []byte `yaml:"b"`
}

type structD struct {
	I bool
}

type StructM struct {
	M string
}

const testfileContent = `
a: "b"
c: 10
e: 123.456
b: "abc"
i: 42
j: 420

y:
    x: "y"
    z: ["1", "2", "3"]
    a:
        "i": true

A:
    x: "y"
    B: 
        C:
            m: "n"
            b: "boo"
        D: ["x", "y", "z"]

m: "n"

MI: ["q", "w"]
`
