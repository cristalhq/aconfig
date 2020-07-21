package aconfig

import (
	"encoding/json"
	"flag"
	"os"
	"testing"
	"time"
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
		UseDefaults: true,
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
		UseDefaults: true,
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
	filepath := "testdata/config1.json"
	loader := NewLoader(LoaderConfig{
		UseFile: true,
		Files:   []string{filepath},
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

func TestLoadEnv(t *testing.T) {
	setEnv(t, "TST_STR", "str-env")
	setEnv(t, "TST_INT", "121")
	setEnv(t, "TST_HTTPPORT", "3000")
	setEnv(t, "TST_SUB_FLOAT", "222.333")
	setEnv(t, "TST_EM", "em-env")
	defer os.Clearenv()

	loader := NewLoader(LoaderConfig{
		UseEnv:    true,
		EnvPrefix: "tst",
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
	flag.Parse()

	loader := NewLoader(LoaderConfig{
		UseFlag:    true,
		FlagPrefix: "tst",
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

func loadFile(t *testing.T, file string, dst interface{}) {
	f, err := os.Open(file)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.NewDecoder(f).Decode(dst); err != nil {
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
