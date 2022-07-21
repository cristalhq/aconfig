package aconfig

import (
	"embed"
	"fmt"
	"io"
	"net/url"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"
)

type LogLevel int8

func (l *LogLevel) UnmarshalText(text []byte) error {
	switch string(text) {
	case "debug":
		*l = -1
	case "info":
		*l = 0
	case "warn":
		*l = 1
	case "error":
		*l = 2
	default:
		return fmt.Errorf("unknown log level: %s", text)
	}
	return nil
}

func TestDefaults(t *testing.T) {
	var cfg TestConfig
	loader := LoaderFor(&cfg, Config{
		SkipFiles: true,
		SkipEnv:   true,
		SkipFlags: true,
	})
	failIfErr(t, loader.Load())

	want := TestConfig{
		Str:      "str-def",
		Bytes:    []byte("bytes-def"),
		Int:      int32Ptr(123),
		HTTPPort: 8080,
		Sub: SubConfig{
			Float: 123.123,
		},
		Anon: struct {
			IsAnon bool `default:"true"`
		}{
			IsAnon: true,
		},
		StrSlice: []string{"1", "2", "3"},
		Slice:    []int{1, 2, 3},
		Map1:     map[string]int{"a": 1, "b": 2, "c": 3},
		Map2:     map[int]string{1: "a", 2: "b", 3: "c"},
		EmbeddedConfig: EmbeddedConfig{
			Em: "em-def",
		},
	}
	mustEqual(t, cfg, want)
}

func TestDefaults_AllTypes(t *testing.T) {
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

		Level LogLevel `default:"warn"`
	}

	var cfg AllTypesConfig
	loader := LoaderFor(&cfg, Config{
		SkipFiles: true,
		SkipEnv:   true,
		SkipFlags: true,
	})
	failIfErr(t, loader.Load())

	want := AllTypesConfig{
		Bool:    true,
		String:  "str",
		Int:     1,
		Int8:    12,
		Int16:   123,
		Int32:   13,
		Int64:   23,
		Uint:    1234,
		Uint8:   124,
		Uint16:  134,
		Uint32:  234,
		Uint64:  24,
		Float32: 1234.213,
		Float64: 1234.234,
		Dur:     time.Hour + 2*time.Minute + 3*time.Second,
		// TODO: support time
		// Time :2000-04-05 10:20:30 +0000 UTC,
		Level: LogLevel(1),
	}
	mustEqual(t, cfg, want)
}

func TestDefaults_OtherNumberFormats(t *testing.T) {
	type OtherNumberFormats struct {
		Int    int   `default:"0b111"`
		Int8   int8  `default:"0o123"`
		Int8x2 int8  `default:"0123"`
		Int16  int16 `default:"0x123"`

		Uint   uint   `default:"0b111"`
		Uint8  uint8  `default:"0o123"`
		Uint16 uint16 `default:"0123"`
		Uint32 uint32 `default:"0x123"`
	}

	var cfg OtherNumberFormats
	loader := LoaderFor(&cfg, Config{
		SkipFiles: true,
		SkipEnv:   true,
		SkipFlags: true,
	})
	failIfErr(t, loader.Load())

	want := OtherNumberFormats{
		Int:    7,
		Int8:   83,
		Int8x2: 83,
		Int16:  291,

		Uint:   7,
		Uint8:  83,
		Uint16: 83,
		Uint32: 291,
	}
	mustEqual(t, cfg, want)
}

func TestJSON(t *testing.T) {
	filepath := createTestFile(t)

	var cfg structConfig
	loader := LoaderFor(&cfg, Config{
		SkipDefaults: true,
		SkipEnv:      true,
		SkipFlags:    true,
		Files:        []string{filepath},
	})
	failIfErr(t, loader.Load())

	want := wantConfig
	mustEqual(t, cfg, want)
}

func TestJSONWithOmitempty(t *testing.T) {
	type TestConfig struct {
		APIKey string `json:"b,omitempty"`
	}

	var cfg TestConfig
	loader := LoaderFor(&cfg, Config{
		SkipDefaults:       true,
		SkipEnv:            true,
		SkipFlags:          true,
		AllowUnknownFields: true,
		Files:              []string{createTestFile(t)},
	})
	failIfErr(t, loader.Load())
}

func TestCustomFile(t *testing.T) {
	filepath := createTestFile(t, "custom.config")

	var cfg structConfig
	loader := LoaderFor(&cfg, Config{
		SkipDefaults: true,
		SkipEnv:      true,
		SkipFlags:    true,
		Files:        []string{filepath},
		FileDecoders: map[string]FileDecoder{
			".config": &jsonDecoder{},
		},
	})
	failIfErr(t, loader.Load())

	want := wantConfig
	mustEqual(t, cfg, want)
}

func TestFile(t *testing.T) {
	filepath := "testdata/config.json"

	var cfg TestConfig
	loader := LoaderFor(&cfg, Config{
		SkipDefaults: true,
		SkipEnv:      true,
		SkipFlags:    true,
		Files:        []string{filepath},
	})
	failIfErr(t, loader.Load())

	want := TestConfig{
		Str:      "str-json",
		Bytes:    []byte("Ynl0ZXMtanNvbg=="),
		Int:      int32Ptr(101),
		HTTPPort: 65000,
		Sub: SubConfig{
			Float: 999.111,
		},
		Anon: struct {
			IsAnon bool `default:"true"`
		}{
			IsAnon: true,
		},
	}
	mustEqual(t, cfg, want)
}

//go:embed testdata
var configEmbed embed.FS

func TestFileEmbed(t *testing.T) {
	filepath := "testdata/config.json"

	var cfg TestConfig
	loader := LoaderFor(&cfg, Config{
		SkipDefaults: true,
		SkipEnv:      true,
		SkipFlags:    true,
		Files:        []string{filepath},
		FileSystem:   configEmbed,
	})
	failIfErr(t, loader.Load())

	want := TestConfig{
		Str:      "str-json",
		Bytes:    []byte("Ynl0ZXMtanNvbg=="),
		Int:      int32Ptr(101),
		HTTPPort: 65000,
		Sub: SubConfig{
			Float: 999.111,
		},
		Anon: struct {
			IsAnon bool `default:"true"`
		}{
			IsAnon: true,
		},
	}
	mustEqual(t, cfg, want)
}

func TestFileMerging(t *testing.T) {
	file1 := "testdata/config1.json"
	file2 := "testdata/config2.json"
	file3 := "testdata/config3.json"

	var cfg TestConfig
	loader := LoaderFor(&cfg, Config{
		SkipDefaults: true,
		SkipEnv:      true,
		SkipFlags:    true,
		MergeFiles:   true,
		Files:        []string{file1, file2, file3},
	})
	failIfErr(t, loader.Load())

	want := TestConfig{
		Str:      "111",
		HTTPPort: 222,
		Sub: SubConfig{
			Float: 333.333,
		},
	}
	mustEqual(t, cfg, want)
}

func TestFileFlag(t *testing.T) {
	file1 := "testdata/config1.json"

	flags := []string{
		"-file_flag=testdata/config2.json",
	}

	var cfg TestConfig
	loader := LoaderFor(&cfg, Config{
		SkipDefaults: true,
		SkipEnv:      true,
		MergeFiles:   true,
		FileFlag:     "file_flag",
		Files:        []string{file1},
		Args:         flags,
	})
	failIfErr(t, loader.Load())

	want := TestConfig{
		Str:      "111",
		HTTPPort: 222,
	}
	mustEqual(t, cfg, want)
}

func TestBadFileFlag(t *testing.T) {
	flags := []string{
		"-file_flag=",
	}

	var cfg TestConfig
	loader := LoaderFor(&cfg, Config{
		SkipDefaults: true,
		SkipEnv:      true,
		FileFlag:     "file_flag",
		Args:         flags,
	})
	failIfOk(t, loader.Load())
}

func TestNoFileFlagValue(t *testing.T) {
	file1 := "testdata/config1.json"

	var cfg TestConfig
	loader := LoaderFor(&cfg, Config{
		SkipDefaults: true,
		SkipEnv:      true,
		FileFlag:     "file_flag",
		Files:        []string{file1},
		Args:         []string{}, // no file_flag
	})
	failIfErr(t, loader.Load())

	want := TestConfig{
		Str:      "111",
		HTTPPort: 111,
	}
	mustEqual(t, cfg, want)
}

func TestEnv(t *testing.T) {
	setEnv(t, "TST_STR", "str-env")
	setEnv(t, "TST_BYTES", "bytes-env")
	setEnv(t, "TST_INT", "121")
	setEnv(t, "TST_HTTP_PORT", "3000")
	setEnv(t, "TST_SUB_FLOAT", "222.333")
	setEnv(t, "TST_ANON_IS_ANON", "true")
	setEnv(t, "TST_EM", "em-env")
	defer os.Clearenv()

	var cfg TestConfig
	loader := LoaderFor(&cfg, Config{
		SkipDefaults: true,
		SkipFiles:    true,
		SkipFlags:    true,
		EnvPrefix:    "TST",
	})
	failIfErr(t, loader.Load())

	want := TestConfig{
		Str:      "str-env",
		Bytes:    []byte("bytes-env"),
		Int:      int32Ptr(121),
		HTTPPort: 3000,
		Sub: SubConfig{
			Float: 222.333,
		},
		Anon: struct {
			IsAnon bool `default:"true"`
		}{
			IsAnon: true,
		},
		EmbeddedConfig: EmbeddedConfig{
			Em: "em-env",
		},
	}

	mustEqual(t, cfg, want)
}

func TestFlag(t *testing.T) {
	var cfg TestConfig
	loader := LoaderFor(&cfg, Config{
		SkipDefaults: true,
		SkipFiles:    true,
		SkipEnv:      true,
		FlagPrefix:   "tst",
	})

	flags := []string{
		"-tst.str=str-flag",
		"-tst.bytes=Ynl0ZXMtZmxhZw==",
		"-tst.int=1001",
		"-tst.http_port=30000",
		"-tst.sub.float=123.321",
		"-tst.anon.is_anon=true",
		"-tst.em=em-flag",
	}

	failIfErr(t, loader.Flags().Parse(flags))

	failIfErr(t, loader.Load())

	want := TestConfig{
		Str:      "str-flag",
		Bytes:    []byte("Ynl0ZXMtZmxhZw=="),
		Int:      int32Ptr(1001),
		HTTPPort: 30000,
		Sub: SubConfig{
			Float: 123.321,
		},
		Anon: struct {
			IsAnon bool `default:"true"`
		}{
			IsAnon: true,
		},
		EmbeddedConfig: EmbeddedConfig{
			Em: "em-flag",
		},
	}

	mustEqual(t, cfg, want)
}

func TestExactName(t *testing.T) {
	setEnv(t, "STR", "str-env")
	setEnv(t, "TST_STR", "bar-env")
	defer os.Clearenv()

	type Foo struct {
		String string `env:"STR,exact"`
	}
	type ExactConfig struct {
		Foo Foo
		Bar string `env:"STR"`
	}
	var cfg ExactConfig

	loader := LoaderFor(&cfg, Config{
		SkipDefaults:     true,
		SkipFiles:        true,
		SkipFlags:        true,
		AllowUnknownEnvs: true,
		EnvPrefix:        "TST",
	})
	failIfErr(t, loader.Load())

	want := ExactConfig{
		Foo: Foo{
			String: "str-env",
		},
		Bar: "bar-env",
	}

	mustEqual(t, cfg, want)
}

func TestSkipName(t *testing.T) {
	setEnv(t, "STR", "str-env")
	setEnv(t, "BAR", "bar-env")
	defer os.Clearenv()

	type Foo struct {
		String string `env:"STR"`
	}
	type ExactConfig struct {
		Foo Foo    `env:"-"`
		Bar string `default:"def" env:"-"`
	}
	var cfg ExactConfig

	loader := LoaderFor(&cfg, Config{
		SkipFiles: true,
		SkipFlags: true,
	})
	failIfErr(t, loader.Load())

	want := ExactConfig{
		Foo: Foo{
			String: "str-env",
		},
		Bar: "def",
	}

	mustEqual(t, cfg, want)
}

func TestDuplicatedName(t *testing.T) {
	setEnv(t, "FOO_BAR", "str-env")
	defer os.Clearenv()

	type Foo struct {
		Bar string
	}
	type ExactConfig struct {
		Foo    Foo
		FooBar string
	}
	var cfg ExactConfig

	loader := LoaderFor(&cfg, Config{
		SkipFlags:       true,
		AllowDuplicates: true,
	})
	failIfErr(t, loader.Load())

	want := ExactConfig{
		Foo: Foo{
			Bar: "str-env",
		},
		FooBar: "str-env",
	}

	mustEqual(t, cfg, want)
}

func TestFailOnDuplicatedName(t *testing.T) {
	type Foo struct {
		Bar string
	}
	type ExactConfig struct {
		Foo    Foo
		FooBar string
	}
	var cfg ExactConfig

	loader := LoaderFor(&cfg, Config{
		SkipFlags: true,
	})

	err := loader.Load()
	failIfOk(t, err)

	if !strings.Contains(err.Error(), "is duplicated") {
		t.Fatalf("got %s", err.Error())
	}
}

func TestFailOnDuplicatedFlag(t *testing.T) {
	type Foo struct {
		Bar string `flag:"yes"`
		Baz string `flag:"yes"`
	}

	err := LoaderFor(&Foo{}, Config{}).Load()
	failIfOk(t, err)

	want := `init loader: duplicate flag "yes"`
	mustEqual(t, err.Error(), want)
}

func TestUsage(t *testing.T) {
	loader := LoaderFor(&EmbeddedConfig{}, Config{})

	var builder strings.Builder
	flags := loader.Flags()
	flags.SetOutput(&builder)
	flags.PrintDefaults()

	got := builder.String()
	want := `  -em string
    	use... em...field. (default "em-def")
`

	mustEqual(t, got, want)
}

func TestBadDefauts(t *testing.T) {
	f := func(cfg interface{}) {
		t.Helper()

		loader := LoaderFor(cfg, Config{
			SkipFiles: true,
			SkipEnv:   true,
			SkipFlags: true,
		})
		failIfOk(t, loader.Load())
	}

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
		var cfg TestConfig
		loader := LoaderFor(&cfg, Config{
			SkipDefaults:       true,
			SkipEnv:            true,
			SkipFlags:          true,
			FailOnFileNotFound: true,
			Files:              []string{filepath},
		})
		failIfOk(t, loader.Load())
	}

	t.Run("no_such_file.json", func(*testing.T) {
		f("no_such_file.json")
	})

	t.Run("bad_config.json", func(t *testing.T) {
		filepath := t.TempDir() + "unknown.ext"
		file, err := os.Create(filepath)
		failIfErr(t, err)
		defer file.Close()

		_, err = file.WriteString(`{almost": "json`)
		failIfErr(t, err)

		f(filepath)
	})

	t.Run("unknown.ext", func(t *testing.T) {
		filepath := t.TempDir() + "unknown.ext"
		file, err := os.Create(filepath)
		failIfErr(t, err)
		defer file.Close()

		f(filepath)
	})
}

func TestFailOnFileNotFound(t *testing.T) {
	f := func(filepath string) {
		t.Helper()

		loader := LoaderFor(&TestConfig{}, Config{
			SkipDefaults:       true,
			SkipEnv:            true,
			SkipFlags:          true,
			FailOnFileNotFound: false,
			Files:              []string{filepath},
		})

		failIfErr(t, loader.Load())
	}

	f("testdata/config.json")
	f("testdata/not_found.json")
}

func TestBadEnvs(t *testing.T) {
	setEnv(t, "TST_HTTP_PORT", "30a00")
	defer os.Clearenv()

	loader := LoaderFor(&TestConfig{}, Config{
		SkipDefaults: true,
		SkipFiles:    true,
		SkipFlags:    true,
		EnvPrefix:    "TST",
	})

	failIfOk(t, loader.Load())
}

func TestBadFlags(t *testing.T) {
	loader := LoaderFor(&TestConfig{}, Config{
		SkipDefaults: true,
		SkipFiles:    true,
		SkipEnv:      true,
		FlagPrefix:   "tst",
	})

	args := []string{"-tst.param=10a01"}

	failIfErr(t, loader.Flags().Parse(args))
	failIfOk(t, loader.Load())
}

func TestUnknownFields(t *testing.T) {
	filepath := "testdata/unknown_fields.json"

	var cfg TestConfig
	loader := LoaderFor(&cfg, Config{
		SkipDefaults: true,
		SkipEnv:      true,
		SkipFlags:    true,
		Files:        []string{filepath},
	})

	err := loader.Load()
	failIfOk(t, err)

	if !strings.Contains(err.Error(), "unknown field in file") {
		t.Fatalf("got %s", err.Error())
	}
}

func TestUnknownEnvs(t *testing.T) {
	setEnv(t, "TST_STR", "defined")
	setEnv(t, "TST_UNKNOWN", "42")
	setEnv(t, "JUST_ENV", "JUST_VALUE")
	defer os.Clearenv()

	var cfg TestConfig
	loader := LoaderFor(&cfg, Config{
		SkipDefaults: true,
		SkipFiles:    true,
		SkipFlags:    true,
		EnvPrefix:    "TST",
	})

	err := loader.Load()
	failIfOk(t, err)

	if !strings.Contains(err.Error(), "unknown environment var") {
		t.Fatalf("got %s", err.Error())
	}
}

func TestUnknownEnvsWithEmptyPrefix(t *testing.T) {
	setEnv(t, "STR", "defined")
	setEnv(t, "UNKNOWN", "42")
	defer os.Clearenv()

	var cfg TestConfig
	loader := LoaderFor(&cfg, Config{
		SkipDefaults: true,
		SkipFiles:    true,
		SkipFlags:    true,
	})

	failIfErr(t, loader.Load())
}

func TestUnknownFlags(t *testing.T) {
	loader := LoaderFor(&TestConfig{}, Config{
		SkipDefaults: true,
		SkipFiles:    true,
		SkipEnv:      true,
		FlagPrefix:   "tst",
	})

	flags := []string{
		"-tst.str=str-flag",
		"-tst.unknown=1001",
		"-just_env=just_value",
	}

	// just for tests
	flagSet := loader.Flags()
	flagSet.SetOutput(io.Discard)

	// define flag with a loader's prefix which is unknown
	flagSet.Int("tst.unknown", 42, "")
	flagSet.String("just_env", "just_def", "")

	failIfErr(t, flagSet.Parse(flags))

	err := loader.Load()
	failIfOk(t, err)

	if !strings.Contains(err.Error(), "unknown flag") {
		t.Fatalf("got %s", err.Error())
	}
}

func TestUnknownFlagsWithEmptyPrefix(t *testing.T) {
	loader := LoaderFor(&TestConfig{}, Config{
		SkipDefaults: true,
		SkipFiles:    true,
		SkipEnv:      true,
	})

	flags := []string{
		"-str=str-flag",
		"-unknown=1001",
	}

	// just for tests
	flagSet := loader.Flags()
	flagSet.SetOutput(io.Discard)

	// define flag with a loader's prefix which is unknown
	flagSet.Int("unknown", 42, "")

	failIfErr(t, flagSet.Parse(flags))
	failIfErr(t, loader.Load())
}

// flag.FlagSet already fails on undefined flag
func TestUnknownFlagsStdlib(t *testing.T) {
	loader := LoaderFor(&TestConfig{}, Config{
		SkipDefaults: true,
		SkipFiles:    true,
		SkipEnv:      true,
		FlagPrefix:   "tst",
	})

	flags := []string{
		"-tst.str=str-flag",
		"-tst.unknown=1001",
	}

	// just for tests
	flagSet := loader.Flags()
	flagSet.SetOutput(io.Discard)

	failIfOk(t, flagSet.Parse(flags))
}

func TestCustomEnvsAndArgs(t *testing.T) {
	var cfg TestConfig
	loader := LoaderFor(&cfg, Config{
		SkipDefaults: true,
		Envs:         []string{"PARAM=2"},
		Args:         []string{"-str=4"},
	})

	failIfErr(t, loader.Load())

	want := TestConfig{
		Str:   "4",
		Param: 2,
	}
	mustEqual(t, cfg, want)
}

func TestCustomNames(t *testing.T) {
	type TestConfig struct {
		A int `default:"-1" env:"ONE"`
		B int `default:"-1" flag:"two"`
		C int `default:"-1" env:"three" flag:"four"`
	}

	setEnv(t, "ONE", "1")
	setEnv(t, "three", "3")
	defer os.Clearenv()

	var cfg TestConfig
	loader := LoaderFor(&cfg, Config{
		Args: []string{"-two=2", "-four=4"},
	})

	failIfErr(t, loader.Load())

	mustEqual(t, cfg.A, 1)
	mustEqual(t, cfg.B, 2)
	mustEqual(t, cfg.C, 4)
}

func TestDontGenerateTags(t *testing.T) {
	type testConfig struct {
		A        string `json:"aaa"`
		B        string `yaml:"aaa" toml:"bbb"`
		DooDoo   string
		HTTPPort int    `yaml:"port"`
		D        string `env:"aaa"`
		E        string `flag:"aaa"`
	}

	want := map[string]string{
		"A::json":        "aaa",
		"B::yaml":        "aaa",
		"C::toml":        "c",
		"DooDoo::toml":   "DooDoo",
		"DooDoo::flag":   "doo_doo",
		"HTTPPort::flag": "http_port",
		"HTTPPort::json": "HTTPPort",
		"HTTPPort::yaml": "port",
		"D::env":         "aaa",
		"E::flag":        "aaa",
		"E::json":        "E",
	}
	cfg := Config{
		DontGenerateTags: true,
	}
	LoaderFor(&testConfig{}, cfg).WalkFields(func(f Field) bool {
		for _, tag := range []string{"json", "yaml", "env", "flag"} {
			k := f.Name() + "::" + tag
			if v, ok := want[k]; ok && v != f.Tag(tag) {
				t.Fatalf("%v: got %v, want %v", tag, f.Tag(tag), v)
				return false
			}
		}
		return true
	})
}

func TestWalkFields(t *testing.T) {
	type TestConfig struct {
		A int `default:"-1" env:"one" marco:"polo"`
		B struct {
			C int `default:"-1" flag:"two" usage:"pretty simple usage duh" json:"kek" yaml:"lel" toml:"mde"`
			D struct {
				E int `default:"-1" env:"three" json:"kek" yaml:"lel" toml:"mde"`
			}
		}
	}

	fields := []struct {
		Name         string
		ParentName   string
		DefaultValue string
		EnvName      string
		FlagName     string
		Usage        string
	}{
		{
			Name:         "A",
			EnvName:      "one",
			DefaultValue: "-1",
		},
		{
			Name:         "B.C",
			ParentName:   "B",
			FlagName:     "two",
			DefaultValue: "-1",
			Usage:        "pretty simple usage duh",
		},
		{
			Name:         "B.D.E",
			ParentName:   "B.D",
			EnvName:      "three",
			DefaultValue: "-1",
		},
	}

	i := 0

	LoaderFor(&TestConfig{}, Config{}).WalkFields(func(f Field) bool {
		wantFields := fields[i]
		mustEqual(t, f.Name(), wantFields.Name)
		mustEqual(t, f.Name(), wantFields.Name)
		if parent, ok := f.Parent(); ok {
			mustEqual(t, parent.Name(), wantFields.ParentName)
		}
		mustEqual(t, f.Tag("default"), wantFields.DefaultValue)
		mustEqual(t, f.Tag("usage"), wantFields.Usage)
		i++
		return true
	})

	mustEqual(t, i, 3)

	i = 0
	LoaderFor(&TestConfig{}, Config{}).WalkFields(func(f Field) bool {
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
	loader := LoaderFor(&TestConfig{}, Config{
		SkipFlags: true,
		Args:      []string{},
	})
	failIfErr(t, loader.Load())

	if flags := loader.Flags().NFlag(); flags != 0 {
		t.Errorf("want empty, got %v", flags)
	}
}

func TestPassBadStructs(t *testing.T) {
	f := func(cfg interface{}) {
		t.Helper()

		defer func() {
			t.Helper()
			if err := recover(); err == nil {
				t.Fatal()
			}
		}()

		_ = LoaderFor(cfg, Config{})
	}

	f(nil)
	f(map[string]string{})
	f([]string{})
	f([4]string{})
	f(func() {})

	type S struct {
		Foo int
	}
	f(S{})
}

func TestBadRequiredTag(t *testing.T) {
	type TestConfig struct {
		Field string `required:"boom"`
	}

	f := func(cfg interface{}) {
		t.Helper()

		defer func() {
			t.Helper()
			if err := recover(); err == nil {
				t.Fatal()
			}
		}()

		_ = LoaderFor(cfg, Config{})
	}

	f(&TestConfig{})
}

func setEnv(t *testing.T, key, value string) {
	failIfErr(t, os.Setenv(key, value))
}

func int32Ptr(a int32) *int32 {
	return &a
}

func createTestFile(t *testing.T, name ...string) string {
	t.Helper()
	mustEqual(t, len(name) < 2, true)

	dir := t.TempDir()
	t.Cleanup(func() {
		os.RemoveAll(dir)
	})

	filepath := dir + "/testfile.json"
	if len(name) == 1 {
		filepath = dir + name[0]
	}

	f, err := os.Create(filepath)
	failIfErr(t, err)
	defer f.Close()
	_, err = f.WriteString(testfileContent)
	failIfErr(t, err)
	return filepath
}

type TestConfig struct {
	Str      string `default:"str-def"`
	Bytes    []byte `default:"bytes-def"`
	Int      *int32 `default:"123"`
	HTTPPort int    `default:"8080"`
	Param    int    // no default tag, so default value
	Sub      SubConfig
	Anon     struct {
		IsAnon bool `default:"true"`
	}

	StrSlice []string       `default:"1,2,3" usage:"just pass strings"`
	Slice    []int          `default:"1,2,3" usage:"just pass elements"`
	Map1     map[string]int `default:"a:1,b:2,c:3"`
	Map2     map[int]string `default:"1:a,2:b,3:c"`

	EmbeddedConfig
}

type EmbeddedConfig struct {
	Em string `default:"em-def" usage:"use... em...field."`
}

type SubConfig struct {
	Float float64 `default:"123.123"`
}

type structConfig struct {
	A string
	C int
	E float64
	B []byte
	I *int32
	J *int64
	Y structY

	AA structA `json:"A"`
	StructM

	M interface{} `json:"M"`

	P *structP `json:"P"`
}

type structY struct {
	X string
	Z []int
	A structD
}

type structA struct {
	X  string  `json:"x"`
	BB structB `json:"B"`
}

type structB struct {
	CC structC  `json:"C"`
	DD []string `json:"D"`
}

type structC struct {
	MM string `json:"m"`
	BB []byte `json:"b"`
}

type structD struct {
	I bool
}

type StructM struct {
	M string
}

type structP struct {
	P string `json:"P"`
}

const testfileContent = `{
    "a": "b",
    "c": 10,
    "e": 123.456,
    "b": "abc",
    "i": 42,
	"j": 420,

    "y": {
        "x": "y",
		"z": [1, "2", "3"],
		"a": {
			"i": true
		}
    },

    "A": {
        "x": "y",
        "B": {
            "C": {
                "m": "n",
                "b": "boo"
            },
            "D": ["x", "y", "z"]
        }
	},

	"m": "n",

	"M":["q", "w"],
	
	"P": {
		"P": "r"
	}
}
`

var wantConfig = func() structConfig {
	i := int32(42)
	j := int64(420)
	mInterface := make([]interface{}, 2)
	for iI, vI := range []string{"q", "w"} {
		mInterface[iI] = vI
	}

	return structConfig{
		A: "b",
		C: 10,
		E: 123.456,
		B: []byte("abc"),
		I: &i,
		J: &j,
		Y: structY{
			X: "y",
			Z: []int{1, 2, 3},
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
		M: mInterface,
		P: &structP{
			P: "r",
		},
	}
}()

type ConfigTest struct {
	VCenter ConfigVCenter `json:"vcenter" env:"VCENTER"`
}

type ConfigVCenter struct {
	User        string                  `json:"user" env:"USER"`
	Password    string                  `json:"password" env:"PASSWORD"`
	Port        string                  `json:"port" env:"PORT"`
	Datacenters []ConfigVCenterDCRegion `json:"datacenters" env:"-"`
}

type ConfigVCenterDCRegion struct {
	Region    string            `json:"region"`
	Addresses []ConfigVCenterDC `json:"addresses"`
}

type ConfigVCenterDC struct {
	Zone       string `json:"zone"`
	Address    string `json:"address"`
	Datacenter string `json:"datacenter"`
}

func TestSliceStructs(t *testing.T) {
	var cfg ConfigTest
	loader := LoaderFor(&cfg, Config{
		SkipDefaults: true,
		SkipEnv:      true,
		SkipFlags:    true,
		Files:        []string{"testdata/complex.json"},
	})

	failIfErr(t, loader.Load())

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
	mustEqual(t, cfg, want)
}

func TestMapOfMap(t *testing.T) {
	type TestConfig struct {
		Options map[string]float64
	}
	var cfg TestConfig

	loader := LoaderFor(&cfg, Config{
		SkipDefaults: true,
		SkipEnv:      true,
		SkipFlags:    true,
		Files:        []string{"testdata/toy.json"},
	})

	failIfErr(t, loader.Load())

	want := TestConfig{
		Options: map[string]float64{
			"foo": 0.4,
			"bar": 0.25,
		},
	}

	mustEqual(t, cfg, want)
}

func TestBad(t *testing.T) {
	type TestConfig struct {
		Params url.Values
	}
	var cfg TestConfig
	os.Setenv("PARAMS", "foo:bar")
	defer os.Unsetenv("PARAMS")

	loader := LoaderFor(&cfg, Config{
		SkipFlags: true,
	})
	failIfErr(t, loader.Load())

	p, err := url.ParseQuery("foo=bar")
	if err != nil {
		t.Fatal(err)
	}
	want := TestConfig{
		Params: p,
	}

	mustEqual(t, cfg, want)
}

func TestFileConfigFlagDelim(t *testing.T) {
	type TestConfig struct {
		Options struct {
			Foo float64
			Bar float64
		}
	}
	var cfg TestConfig

	loader := LoaderFor(&cfg, Config{
		SkipDefaults:  true,
		SkipEnv:       true,
		SkipFlags:     true,
		FlagDelimiter: "_",

		Files: []string{"testdata/toy.json"},
	})

	failIfErr(t, loader.Load())

	want := TestConfig{Options: struct {
		Foo float64
		Bar float64
	}{0.4, 0.25}}

	mustEqual(t, cfg, want)
}

func failIfOk(t testing.TB, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("must be non-nil")
	}
}

func failIfErr(t testing.TB, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

func mustEqual(t testing.TB, got, want interface{}) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("\nhave %+v\nwant %+v", got, want)
	}
}
