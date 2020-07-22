package aconfig

import (
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/BurntSushi/toml"
	"gopkg.in/yaml.v2"
)

type TestConfig struct {
	Str      string `default:"str"`
	Int      int32  `default:"123"`
	HTTPPort int    `default:"8080"`
	Sub      SubConfig
	//PtrSub   *SubConfig

	EmbeddedConfig
}

type EmbeddedConfig struct {
	Em string `default:"xxx"`
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

type DurationConfig struct {
	MyDur MyDuration `default:"1h2m3s" json:"my_dur"`
}

type MyDuration string

func (m MyDuration) Duration() (time.Duration, error) {
	return time.ParseDuration(string(m))
}

func TestLoadDefault_AllTypesConfig(t *testing.T) {
	loader := NewLoader(LoaderConfig{
		SkipFile: true,
		SkipEnv:  true,
		SkipFlag: true,
	})
	var cfg, want AllTypesConfig
	if err := loader.Load(&cfg); err != nil {
		t.Fatal(err)
	}

	loadFile(t, "testdata/all_types_config.json", &want)

	if got := cfg; got != want {
		t.Fatalf("want %v, got %v", want, got)
	}
}

func TestLoadDefault_DurationConfig(t *testing.T) {
	loader := NewLoader(LoaderConfig{
		SkipFile: true,
		SkipEnv:  true,
		SkipFlag: true,
	})
	var cfg, want DurationConfig
	if err := loader.Load(&cfg); err != nil {
		t.Fatal(err)
	}

	loadFile(t, "testdata/my_duration_config.json", &want)

	if got := cfg; got != want {
		t.Fatalf("want %v, got %v", want, got)
	}
}

func TestLoadFile(t *testing.T) {
	f := func(filepath string) {
		t.Helper()

		loader := NewLoader(LoaderConfig{
			SkipDefaults: true,
			SkipEnv:      true,
			SkipFlag:     true,
			Files:        []string{filepath},
		})
		var cfg, want TestConfig
		if err := loader.Load(&cfg); err != nil {
			t.Fatal(err)
		}

		loadFile(t, filepath, &want)

		if got := cfg; got != want {
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

	loader := NewLoader(LoaderConfig{
		SkipDefaults: true,
		SkipFile:     true,
		SkipFlag:     true,
		EnvPrefix:    "tst",
	})

	var cfg TestConfig
	if err := loader.Load(&cfg); err != nil {
		t.Fatal(err)
	}

	var want TestConfig
	loadFile(t, "testdata/test_config_env.json", &want)

	if got := cfg; got != want {
		t.Fatalf("want %v, got %v", want, got)
	}
}

func TestLoadFlag(t *testing.T) {
	setFlag("tst.str", "str-flag")
	setFlag("tst.int", "1001")
	setFlag("tst.httpport", "30000")
	setFlag("tst.sub.float", "123.321")
	setFlag("tst.ptrsub.float", "321.123")
	setFlag("tst.em", "em-flag")

	loader := NewLoader(LoaderConfig{
		SkipDefaults: true,
		SkipFile:     true,
		SkipEnv:      true,
		FlagPrefix:   "tst",
	})

	var cfg TestConfig
	if err := loader.Load(&cfg); err != nil {
		t.Fatal(err)
	}

	var want TestConfig
	loadFile(t, "testdata/test_config_flag.json", &want)

	if got := cfg; got != want {
		t.Fatalf("want %v, got %v", want, got)
	}
}

func TestBadDefauts(t *testing.T) {
	f := func(testCase func(*Loader) error) {
		t.Helper()
		loader := NewLoader(LoaderConfig{
			SkipFile: true,
			SkipEnv:  true,
			SkipFlag: true,
		})
		if err := testCase(loader); err == nil {
			t.Fatal(err)
		}
	}

	f(func(loader *Loader) error {
		return loader.Load(&struct {
			Bool bool `default:"omg"`
		}{})
	})

	f(func(loader *Loader) error {
		return loader.Load(&struct {
			Int int `default:"1a"`
		}{})
	})

	f(func(loader *Loader) error {
		return loader.Load(&struct {
			Int8 int8 `default:"12a"`
		}{})
	})

	f(func(loader *Loader) error {
		return loader.Load(&struct {
			Int16 int16 `default:"123a"`
		}{})
	})

	f(func(loader *Loader) error {
		return loader.Load(&struct {
			Int32 int32 `default:"13a"`
		}{})
	})

	f(func(loader *Loader) error {
		return loader.Load(&struct {
			Int64 int64 `default:"23a"`
		}{})
	})

	f(func(loader *Loader) error {
		return loader.Load(&struct {
			Uint uint `default:"1234a"`
		}{})
	})

	f(func(loader *Loader) error {
		return loader.Load(&struct {
			Uint8 uint8 `default:"124a"`
		}{})
	})

	f(func(loader *Loader) error {
		return loader.Load(&struct {
			Uint16 uint16 `default:"134a"`
		}{})
	})

	f(func(loader *Loader) error {
		return loader.Load(&struct {
			Uint32 uint32 `default:"234a"`
		}{})
	})

	f(func(loader *Loader) error {
		return loader.Load(&struct {
			Uint64 uint64 `default:"24a"`
		}{})
	})

	f(func(loader *Loader) error {
		return loader.Load(&struct {
			Float32 float32 `default:"1234x213"`
		}{})
	})

	f(func(loader *Loader) error {
		return loader.Load(&struct {
			Float64 float64 `default:"1234x234"`
		}{})
	})

	f(func(loader *Loader) error {
		return loader.Load(&struct {
			Dur time.Duration `default:"1h_2m3s"`
		}{})
	})
}

func TestBadFiles(t *testing.T) {
	f := func(filepath string) {
		t.Helper()

		loader := NewLoader(LoaderConfig{
			SkipDefaults: true,
			SkipEnv:      true,
			SkipFlag:     true,
			Files:        []string{filepath},
		})
		var cfg TestConfig
		if err := loader.Load(&cfg); err == nil {
			t.Fatal(err)
		}
	}

	f("testdata/no_such_file.json")
	f("testdata/unknown.ext")
}

func TestBadEnvs(t *testing.T) {
	setEnv(t, "TST_HTTPPORT", "30a00")
	defer os.Clearenv()

	loader := NewLoader(LoaderConfig{
		SkipDefaults: true,
		SkipFile:     true,
		SkipFlag:     true,
		EnvPrefix:    "tst",
	})

	var cfg TestConfig
	if err := loader.Load(&cfg); err == nil {
		t.Fatal(err)
	}
}

func TestBadFlags(t *testing.T) {
	type Config struct {
		Field int
	}
	setFlag("tst.field", "10a01")

	loader := NewLoader(LoaderConfig{
		SkipDefaults: true,
		SkipFile:     true,
		SkipEnv:      true,
		FlagPrefix:   "tst",
	})

	var cfg Config
	if err := loader.Load(&cfg); err == nil {
		t.Fatal(err)
	}
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

func setFlag(flg, value string) {
	flag.String(flg, value, "testing")
}
