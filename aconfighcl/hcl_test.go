package aconfighcl_test

import (
	"os"
	"reflect"
	"testing"

	"github.com/cristalhq/aconfig"
	"github.com/cristalhq/aconfig/aconfighcl"
)

func TestHCL(t *testing.T) {
	filepath := createTestFile(t, "test.hcl", testfileContent)

	var cfg structConfig
	loader := aconfig.LoaderFor(&cfg, aconfig.Config{
		SkipDefaults: true,
		SkipEnv:      true,
		SkipFlags:    true,
		FileDecoders: map[string]aconfig.FileDecoder{
			".hcl": aconfighcl.New(),
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
	VCenter ConfigVCenter `toml:"vcenter"`
}

type ConfigVCenter struct {
	User        string `toml:"user"`
	Password    string `toml:"password"`
	Port        string `toml:"port"`
	Datacenters []ConfigVCenterDCRegion
}

type ConfigVCenterDCRegion struct {
	Region    string `toml:"region"`
	Addresses []ConfigVCenterDC
}

type ConfigVCenterDC struct {
	Zone       string `toml:"zone"`
	Address    string `toml:"address"`
	Datacenter string `toml:"datacenter"`
}

func TestSliceStructs(t *testing.T) {
	content := `[vcenter]
	user = "user-test"
	password = "pass-test"
	port = 8_080
	
	  [[vcenter.datacenters]]
	  region = "region-test"
	
		[[vcenter.datacenters.addresses]]
		zone = "zone-test"
		address = "address-test"
		datacenter = "datacenter-test"	
`

	file := createTestFile(t, "test.hcl", content)

	var cfg ConfigTest
	loader := aconfig.LoaderFor(&cfg, aconfig.Config{
		SkipDefaults:       true,
		SkipEnv:            true,
		SkipFlags:          true,
		FailOnFileNotFound: true,
		Files:              []string{file},
		FileDecoders: map[string]aconfig.FileDecoder{
			".hcl": aconfighcl.New(),
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
	A string  `hcl:"a"`
	C int     `hcl:"c"`
	E float64 `hcl:"e"`
	B []byte  `hcl:"b"`
	I *int32  `hcl:"i"`
	J *int64  `hcl:"j"`
	Y structY `hcl:"y"`

	AA structA `hcl:"A"`
	StructM
	MI interface{} `hcl:"MI"`
}

type structY struct {
	X string   `hcl:"x"`
	Z []string `hcl:"z"`
	A structD  `hcl:"A"`
}

type structA struct {
	X  string  `hcl:"x"`
	BB structB `hcl:"B"`
}

type structB struct {
	CC structC  `hcl:"C"`
	DD []string `hcl:"D"`
}

type structC struct {
	MM string `hcl:"m"`
	BB []byte `hcl:"b"`
}

type structD struct {
	I bool `hcl:"i"`
}

type StructM struct {
	M string `hcl:"M"`
}

const testfileContent = `
"a" = "b"
"c" = 10
"e" = 123.456
"b" = "abc"
"i" = 42
"j" = 420

"y" = {
	"x" = "y"
	"z" = ["1", "2", "3"]
	"A" = {
		"i" = true
	}
}
"A" = {
	"x" = "y"
	"B" = {
		"C" = {
			"m" = "n"
			"b" = "boo"
		}
		"D" = ["x", "y", "z"]
	}
}

"M" = "n"
"MI" = ["q", "w"]
`
