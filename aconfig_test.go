package aconfig

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/BurntSushi/toml"
	"gopkg.in/yaml.v2"
)

type TestConfig struct {
	Str      string `default:"str-def"`
	Int      *int32 `default:"123"`
	HTTPPort int    `default:"8080"`
	Param    int    // no default tag, so default value
	Sub      SubConfig

	Slice []int          `default:"1,2,3" usage:"just pass elements"`
	Map1  map[string]int `default:"a:1,b:2,c:3"`
	Map2  map[int]string `default:"1:a,2:b,3:c"`

	EmbeddedConfig
}

type EmbeddedConfig struct {
	Em string `default:"em-def" usage:"use... em...field."`
}

type SubConfig struct {
	Float float64 `default:"123.123"`
}

type AllTypesConfig struct {
	Bool   bool   `default:"true"`
	String string `default:"str"`

	Int   int   `default:"1"`
	Int8  int8  `default:"12"`
	Int16 int16 `default:"123"`
	Int32 int32 `default:"13"`
	Int64 int64 `default:"23"`

	Uint   uint   `default:"1234"`
	Uint8  uint8  `default:"124"`
	Uint16 uint16 `default:"134"`
	Uint32 uint32 `default:"234"`
	Uint64 uint64 `default:"24"`

	Float32 float32 `default:"1234.213"`
	Float64 float64 `default:"1234.234"`

	Dur  time.Duration `default:"1h2m3s"`
	Time time.Time     `default:"2000-04-05 10:20:30 +0000 UTC"`
}

type OtherNumbersConfig struct {
	Int    int   `default:"0b111"`
	Int8   int8  `default:"0o123"`
	Int8x2 int8  `default:"0123"`
	Int16  int16 `default:"0x123"`

	Uint   uint   `default:"0b111"`
	Uint8  uint8  `default:"0o123"`
	Uint16 uint16 `default:"0123"`
	Uint32 uint32 `default:"0x123"`
}

type DurationConfig struct {
	MyDur MyDuration `default:"1h2m3s" json:"my_dur"`
}

type MyDuration string

func (m MyDuration) Duration() (time.Duration, error) {
	return time.ParseDuration(string(m))
}

func TestLoadDefaults(t *testing.T) {
	loader := LoaderFor(&TestConfig{}).
		SkipFiles().
		SkipEnvironment().
		SkipFlags().
		Build()

	var cfg TestConfig
	if err := loader.Load(&cfg); err != nil {
		t.Fatal(err)
	}

	var want TestConfig
	loadFile(t, "testdata/test_config_def.json", &want)

	if got := cfg; !reflect.DeepEqual(got, want) {
		t.Fatalf("want %v, got %v", want, got)
	}
}

func TestLoadDefault_AllTypesConfig(t *testing.T) {
	loader := LoaderFor(&AllTypesConfig{}).
		SkipFiles().
		SkipEnvironment().
		SkipFlags().
		Build()

	var cfg AllTypesConfig
	if err := loader.Load(&cfg); err != nil {
		t.Fatal(err)
	}

	var want AllTypesConfig
	loadFile(t, "testdata/all_types_config.json", &want)

	if got := cfg; got != want {
		t.Fatalf("want %v, got %v", want, got)
	}
}

func TestLoadDefault_DurationConfig(t *testing.T) {
	loader := LoaderFor(&DurationConfig{}).
		SkipFiles().
		SkipEnvironment().
		SkipFlags().
		Build()

	var cfg DurationConfig
	if err := loader.Load(&cfg); err != nil {
		t.Fatal(err)
	}

	var want DurationConfig
	loadFile(t, "testdata/my_duration_config.json", &want)

	if got := cfg; got != want {
		t.Fatalf("want %v, got %v", want, got)
	}
}

func TestLoadDefault_OtherNumbersConfig(t *testing.T) {
	loader := LoaderFor(&OtherNumbersConfig{}).
		SkipFiles().
		SkipEnvironment().
		SkipFlags().
		Build()

	var cfg OtherNumbersConfig
	if err := loader.Load(&cfg); err != nil {
		t.Fatal(err)
	}

	var want OtherNumbersConfig
	loadFile(t, "testdata/other_numbers_config.json", &want)

	if got := cfg; got != want {
		t.Fatalf("want %v, got %v", want, got)
	}
}

func TestLoadFile(t *testing.T) {
	f := func(filepath string) {
		t.Helper()

		loader := LoaderFor(&TestConfig{}).
			SkipDefaults().
			SkipEnvironment().
			SkipFlags().
			WithFiles([]string{filepath}).
			Build()

		var cfg, want TestConfig
		if err := loader.Load(&cfg); err != nil {
			t.Fatal(err)
		}

		loadFile(t, filepath, &want)

		if got := cfg; !reflect.DeepEqual(got, want) {
			t.Fatalf("want %v, got %v", want, got)
		}
	}

	f("testdata/config1.json")
	f("testdata/config1.yaml")
	f("testdata/config1.toml")
}

func TestLoadFile_WithFiles(t *testing.T) {
	f := func(filepath string) {
		t.Helper()

		loader := LoaderFor(&TestConfig{}).
			SkipDefaults().
			SkipEnvironment().
			SkipFlags().
			Build()

		var cfg, want TestConfig
		if err := loader.LoadWithFile(&cfg, filepath); err != nil {
			t.Fatal(err)
		}

		loadFile(t, filepath, &want)

		if got := cfg; !reflect.DeepEqual(got, want) {
			t.Fatalf("want %v, got %v", want, got)
		}
	}

	f("testdata/config1.json")
	f("testdata/config1.yaml")
	f("testdata/config1.toml")
}

func TestLoadEnv(t *testing.T) {
	setEnv(t, "TST_STR", "str-env")
	setEnv(t, "TST_INT", "121")
	setEnv(t, "TST_HTTPPORT", "3000")
	setEnv(t, "TST_SUB_FLOAT", "222.333")
	setEnv(t, "TST_EM", "em-env")
	defer os.Clearenv()

	loader := LoaderFor(&TestConfig{}).
		SkipDefaults().
		SkipFiles().
		SkipFlags().
		WithEnvPrefix("tst").
		Build()

	var cfg TestConfig
	if err := loader.Load(&cfg); err != nil {
		t.Fatal(err)
	}

	var want TestConfig
	loadFile(t, "testdata/test_config_env.json", &want)

	if got := cfg; !reflect.DeepEqual(got, want) {
		t.Fatalf("want %v, got %v", want, got)
	}
}

func TestLoadFlag(t *testing.T) {
	loader := LoaderFor(&TestConfig{}).
		SkipDefaults().
		SkipFiles().
		SkipEnvironment().
		WithFlagPrefix("tst").
		Build()

	flags := []string{
		"-tst.str=str-flag",
		"-tst.int=1001",
		"-tst.int=1001",
		"-tst.httpport=30000",
		"-tst.sub.float=123.321",
		"-tst.em=em-flag",
	}

	// hack for test :(
	if err := loader.Flags().Parse(flags); err != nil {
		t.Fatal(err)
	}

	var cfg TestConfig
	if err := loader.Load(&cfg); err != nil {
		t.Fatal(err)
	}

	var want TestConfig
	loadFile(t, "testdata/test_config_flag.json", &want)

	if got := cfg; !reflect.DeepEqual(got, want) {
		t.Fatalf("want %v, got %v", want, got)
	}
}

func TestUsage(t *testing.T) {
	loader := LoaderFor(&EmbeddedConfig{}).Build()

	var builder strings.Builder
	flags := loader.Flags()
	flags.SetOutput(&builder)
	flags.PrintDefaults()

	got := builder.String()
	want := `  -em string
    	use... em...field. (default "em-def")
`

	if got != want {
		t.Fatalf("want %v, got %v", want, got)
	}
}

func TestBadDefauts(t *testing.T) {
	f := func(cfg interface{}) {
		t.Helper()

		loader := LoaderFor(cfg).
			SkipFiles().
			SkipEnvironment().
			SkipFlags().
			Build()

		if err := loader.Load(cfg); err == nil {
			t.Fatal(err)
		}
	}

	f(&struct {
		Bool bool `default:"omg"`
	}{})

	f(&struct {
		Bool bool `default:"omg"`
	}{})

	f(&struct {
		Int int `default:"1a"`
	}{})

	f(&struct {
		Int8 int8 `default:"12a"`
	}{})

	f(&struct {
		Int16 int16 `default:"123a"`
	}{})

	f(&struct {
		Int32 int32 `default:"13a"`
	}{})

	f(&struct {
		Int64 int64 `default:"23a"`
	}{})

	f(&struct {
		Uint uint `default:"1234a"`
	}{})

	f(&struct {
		Uint8 uint8 `default:"124a"`
	}{})

	f(&struct {
		Uint16 uint16 `default:"134a"`
	}{})

	f(&struct {
		Uint32 uint32 `default:"234a"`
	}{})

	f(&struct {
		Uint64 uint64 `default:"24a"`
	}{})

	f(&struct {
		Float32 float32 `default:"1234x213"`
	}{})

	f(&struct {
		Float64 float64 `default:"1234x234"`
	}{})

	f(&struct {
		Dur time.Duration `default:"1h_2m3s"`
	}{})

	f(&struct {
		Slice []int `default:"1,a,2"`
	}{})

	f(&struct {
		Map map[string]int `default:"1:a,2:2"`
	}{})

	f(&struct {
		Map map[int]string `default:"a:1"`
	}{})

	f(&struct {
		Map map[int]string `default:"a1"`
	}{})

	f(&struct {
		Array [2]string `default:"a1"`
	}{})
}

func TestBadFiles(t *testing.T) {
	f := func(filepath string) {
		t.Helper()

		loader := LoaderFor(&TestConfig{}).
			SkipDefaults().
			SkipEnvironment().
			SkipFlags().
			StopOnFileError().
			WithFiles([]string{filepath}).
			Build()

		var cfg TestConfig
		if err := loader.Load(&cfg); err == nil {
			t.Fatal(err)
		}
	}

	f("testdata/no_such_file.json")
	f("testdata/bad_config.json")
	f("testdata/unknown.ext")
}

func TestBadEnvs(t *testing.T) {
	setEnv(t, "TST_HTTPPORT", "30a00")
	defer os.Clearenv()

	loader := LoaderFor(&TestConfig{}).
		SkipDefaults().
		SkipFiles().
		SkipFlags().
		WithEnvPrefix("tst").
		Build()

	var cfg TestConfig
	if err := loader.Load(&cfg); err == nil {
		t.Fatal(err)
	}
}

func TestBadFlags(t *testing.T) {
	type Config struct {
		Field int
	}

	loader := LoaderFor(&Config{}).
		SkipDefaults().
		SkipFiles().
		SkipEnvironment().
		WithFlagPrefix("tst").
		Build()

	// hack for test :(
	if err := loader.Flags().Parse([]string{"-tst.field=10a01"}); err != nil {
		t.Fatal(err)
	}

	var cfg Config
	if err := loader.Load(&cfg); err == nil {
		t.Fatal(err)
	}
}

func TestCustomNames(t *testing.T) {
	type Config struct {
		A int `default:"-1" env:"one"`
		B int `default:"-1" flag:"two"`
		C int `default:"-1" env:"three" flag:"four"`
	}

	loader := LoaderFor(&Config{}).Build()

	setEnv(t, "ONE", "1")
	setEnv(t, "three", "3")
	defer os.Clearenv()

	if err := loader.Flags().Parse([]string{"-two=2", "-four=4"}); err != nil {
		t.Fatal(err)
	}

	var cfg Config
	if err := loader.Load(&cfg); err != nil {
		t.Fatal(err)
	}

	if want := 1; cfg.A != want {
		t.Errorf("got %#v, want %#v", cfg.A, want)
	}
	if want := 2; cfg.B != want {
		t.Errorf("got %#v, want %#v", cfg.B, want)
	}
	if want := 4; cfg.C != want {
		t.Errorf("got %#v, want %#v", cfg.C, want)
	}
}

func TestWalkFields(t *testing.T) {
	type Config struct {
		A int `default:"-1" env:"one" marco:"polo"`
		B struct {
			C int `default:"-1" flag:"two" usage:"pretty simple usage duh"`
			D struct {
				E int `default:"-1" env:"three"`
			}
		}
	}

	fields := []Field{
		&fieldData{
			name:         "A",
			envName:      "one",
			defaultValue: "-1",
		},
		&fieldData{
			name:         "B.C",
			flagName:     "two",
			defaultValue: "-1",
			usage:        "pretty simple usage duh",
		},
		&fieldData{
			name:         "B.D.E",
			envName:      "three",
			defaultValue: "-1",
		},
	}

	i := 0

	LoaderFor(&Config{}).Build().WalkFields(func(f Field) bool {
		wantFields := fields[i]
		if f.Name() != wantFields.Name() {
			t.Fatalf("got name %v, want %v", f.Name(), wantFields.Name())
		}
		if f.DefaultValue() != wantFields.DefaultValue() {
			t.Fatalf("got default %#v, want %#v", f.DefaultValue(), wantFields.DefaultValue())
		}
		if f.Usage() != wantFields.Usage() {
			t.Fatalf("got usage %#v, want %#v", f.Usage(), wantFields.Usage())
		}
		i++
		return true
	})

	if want := 3; i != want {
		t.Fatalf("got %v, want %v", i, want)
	}

	i = 0
	LoaderFor(&Config{}).Build().WalkFields(func(f Field) bool {
		if i > 0 {
			return false
		}
		if got := f.Tag("marco"); got != "polo" {
			t.Fatalf("got %v, want %v", got, "polo")
		}
		i++
		return true
	})
	if i != 1 {
		t.Fatal()
	}
}

func TestDontFillFlagsIfDisabled(t *testing.T) {
	type Config struct {
		A int `default:"1"`
	}

	loader := LoaderFor(&Config{}).
		SkipFlags().
		Build()

	if flags := loader.Flags().NFlag(); flags != 0 {
		t.Errorf("want empty, got %v", flags)
	}
}

func TestPassNonStructs(t *testing.T) {
	f := func(cfg interface{}) {
		t.Helper()

		defer func() {
			t.Helper()
			if err := recover(); err == nil {
				t.Fatal()
			}
		}()

		if err := LoaderFor(nil).Build().Load(cfg); err != nil {
			t.Fatal(err)
		}
	}

	f(nil)
	f(map[string]string{})
	f([]string{})
	f([4]string{})
	f(func() {})
}

func TestPanicWhenNotBuilt(t *testing.T) {
	f := func(fn func()) {
		t.Helper()

		defer func() {
			t.Helper()
			if err := recover(); err == nil {
				t.Fatal()
			}
		}()
		fn()
	}

	// ok to pass nils
	f(func() {
		_ = LoaderFor(nil).Load(nil)
	})
	f(func() {
		_ = LoaderFor(nil).LoadWithFile(nil, "")
	})
	f(func() {
		_ = LoaderFor(nil).Flags()
	})
	f(func() {
		LoaderFor(nil).WalkFields(nil)
	})
}

func loadFile(t *testing.T, file string, dst interface{}) {
	f, err := os.Open(file)
	if err != nil {
		t.Fatal(err)
	}
	ext := strings.ToLower(filepath.Ext(file))
	switch ext {
	case ".yaml", ".yml":
		err = yaml.NewDecoder(f).Decode(dst)
	case ".json":
		err = json.NewDecoder(f).Decode(dst)
	case ".toml":
		_, err = toml.DecodeReader(f, dst)
	}
	if err != nil {
		t.Fatal(err)
	}
}

func setEnv(t *testing.T, key, value string) {
	if err := os.Setenv(key, value); err != nil {
		t.Fatal(err)
	}
}
